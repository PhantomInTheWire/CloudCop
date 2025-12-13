/*
Package awsauth provides AWS authentication services using STS AssumeRole.
It supports secure temporary credential management with ExternalID validation.
*/
package awsauth

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

/*
AWSAuth manages AWS authentication and STS operations.
It handles both production mode (using STS AssumeRole) and self-hosted mode
(using direct AWS credentials from environment variables).
*/
type AWSAuth struct {
	cfg         aws.Config
	stsClient   *sts.Client
	selfHosting bool
	endpointURL string
}

/*
NewAWSAuth creates a new AWS authentication service.
It automatically detects the mode (production vs self-hosted) based on
the SELF_HOSTING environment variable and configures accordingly.
*/
func NewAWSAuth() (*AWSAuth, error) {
	ctx := context.Background()
	selfHosting := os.Getenv("SELF_HOSTING") == "1"
	endpointURL := os.Getenv("AWS_ENDPOINT_URL")

	var cfg aws.Config
	var err error

	if selfHosting {
		/*
			Self-hosted mode uses static credentials from environment variables.
			This is for deployments where customers manage their own AWS credentials.
		*/
		accessKey := os.Getenv("AWS_ACCESS_KEY_ID")
		secretKey := os.Getenv("AWS_SECRET_ACCESS_KEY")

		if accessKey == "" || secretKey == "" {
			return nil, errors.New("AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY required in self-hosted mode")
		}

		cfg, err = config.LoadDefaultConfig(ctx,
			config.WithRegion(os.Getenv("AWS_REGION")),
			config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
				accessKey,
				secretKey,
				"",
			)),
		)
	} else {
		/*
			Production mode uses default AWS credential chain.
			The platform's IAM role will be used to assume customer roles.
		*/
		cfg, err = config.LoadDefaultConfig(ctx,
			config.WithRegion(os.Getenv("AWS_REGION")),
		)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	/*
		Configure custom endpoint for LocalStack or other AWS-compatible services.
	*/
	if endpointURL != "" {
		cfg.BaseEndpoint = aws.String(endpointURL)
	}

	return &AWSAuth{
		cfg:         cfg,
		stsClient:   sts.NewFromConfig(cfg),
		selfHosting: selfHosting,
		endpointURL: endpointURL,
	}, nil
}

/*
AssumeRole performs STS AssumeRole to get temporary credentials for a customer AWS account.
This is disabled in self-hosted mode where direct credentials are used instead.
*/
func (a *AWSAuth) AssumeRole(ctx context.Context, input AssumeRoleInput) (*Credentials, error) {
	if a.selfHosting {
		return nil, errors.New("AssumeRole not available in self-hosted mode")
	}

	if input.AccountID == "" || input.ExternalID == "" {
		return nil, ErrInvalidExternalID
	}

	/*
		Construct the IAM role ARN to assume.
		The role name must match what's created by the CloudFormation template.
	*/
	roleARN := fmt.Sprintf("arn:aws:iam::%s:role/CloudCopSecurityScanRole", input.AccountID)
	sessionName := fmt.Sprintf("CloudCopSession-%d", time.Now().Unix())

	result, err := a.stsClient.AssumeRole(ctx, &sts.AssumeRoleInput{
		RoleArn:         aws.String(roleARN),
		RoleSessionName: aws.String(sessionName),
		ExternalId:      aws.String(input.ExternalID),
		DurationSeconds: aws.Int32(21600), // 6 hours
	})

	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrAssumeRoleFailed, err)
	}

	if result.Credentials == nil {
		return nil, errors.New("no credentials returned from STS")
	}

	return &Credentials{
		AccessKeyID:     aws.ToString(result.Credentials.AccessKeyId),
		SecretAccessKey: aws.ToString(result.Credentials.SecretAccessKey),
		SessionToken:    aws.ToString(result.Credentials.SessionToken),
		Expiration:      aws.ToTime(result.Credentials.Expiration),
	}, nil
}

/*
VerifyAccountAccess verifies that we can access the specified AWS account.
In production mode, it assumes the role and gets caller identity.
In self-hosted mode, it uses direct credentials to get caller identity.
*/
func (a *AWSAuth) VerifyAccountAccess(ctx context.Context, input AssumeRoleInput) (*AccountInfo, error) {
	var stsClient *sts.Client

	if a.selfHosting {
		/*
			In self-hosted mode, use existing credentials directly.
		*/
		stsClient = a.stsClient
	} else {
		/*
			In production mode, assume the role first to get temporary credentials.
		*/
		creds, err := a.AssumeRole(ctx, input)
		if err != nil {
			return nil, err
		}

		/*
			Create a new STS client with the temporary credentials.
		*/
		cfg := a.cfg.Copy()
		cfg.Credentials = credentials.NewStaticCredentialsProvider(
			creds.AccessKeyID,
			creds.SecretAccessKey,
			creds.SessionToken,
		)

		stsClient = sts.NewFromConfig(cfg)
	}

	/*
		Verify access by calling GetCallerIdentity.
		This confirms we have valid credentials and returns account information.
	*/
	identity, err := stsClient.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to verify account access: %w", err)
	}

	return &AccountInfo{
		AccountID: aws.ToString(identity.Account),
		ARN:       aws.ToString(identity.Arn),
		UserID:    aws.ToString(identity.UserId),
	}, nil
}

/*
GetAccountID retrieves the AWS account ID using current credentials.
This is primarily used in self-hosted mode during initial setup.
*/
func (a *AWSAuth) GetAccountID(ctx context.Context) (string, error) {
	identity, err := a.stsClient.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
	if err != nil {
		return "", fmt.Errorf("failed to get account ID: %w", err)
	}

	return aws.ToString(identity.Account), nil
}
