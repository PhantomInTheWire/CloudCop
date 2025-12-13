package awsauth

import (
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"
)

const (
	defaultRegion   = "us-west-2"
	roleSessionName = "CloudCopSession"
	roleNameFormat  = "CloudCopSecurityScanRole"
	sessionDuration = int64(21600) // 6 hours in seconds
)

// AWSAuth handles AWS authentication and STS operations
type AWSAuth struct {
	stsClient   *sts.STS
	selfHosted  bool
	region      string
	endpointURL string
}

// NewAWSAuth creates a new AWS authentication service
func NewAWSAuth() (*AWSAuth, error) {
	region := os.Getenv("AWS_REGION")
	if region == "" {
		region = defaultRegion
	}

	endpointURL := os.Getenv("AWS_ENDPOINT_URL")
	selfHosted := os.Getenv("SELF_HOSTING") != ""

	config := &aws.Config{
		Region: aws.String(region),
	}

	// Configure for LocalStack or custom endpoint
	if endpointURL != "" {
		config.Endpoint = aws.String(endpointURL)
	}

	var stsClient *sts.STS

	if selfHosted {
		// Self-Hosted: Use provided credentials directly
		accessKeyID := os.Getenv("AWS_ACCESS_KEY_ID")
		secretAccessKey := os.Getenv("AWS_SECRET_ACCESS_KEY")

		if accessKeyID == "" || secretAccessKey == "" {
			return nil, fmt.Errorf("AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY required in self-hosted mode")
		}

		config.Credentials = credentials.NewStaticCredentials(accessKeyID, secretAccessKey, "")
		sess, err := session.NewSession(config)
		if err != nil {
			return nil, fmt.Errorf("failed to create AWS session: %w", err)
		}
		stsClient = sts.New(sess)
	} else {
		// Production: Create session for STS operations
		sess, err := session.NewSession(config)
		if err != nil {
			return nil, fmt.Errorf("failed to create AWS session: %w", err)
		}
		stsClient = sts.New(sess)
	}

	return &AWSAuth{
		stsClient:   stsClient,
		selfHosted:  selfHosted,
		region:      region,
		endpointURL: endpointURL,
	}, nil
}

// AssumeRole assumes an AWS IAM role using STS with ExternalID
func (a *AWSAuth) AssumeRole(ctx context.Context, input AssumeRoleInput) (*Credentials, error) {
	if a.selfHosted {
		return nil, ErrSelfHostedMode
	}

	roleArn := fmt.Sprintf("arn:aws:iam::%s:role/%s", input.AccountID, roleNameFormat)

	assumeRoleInput := &sts.AssumeRoleInput{
		RoleArn:         aws.String(roleArn),
		RoleSessionName: aws.String(roleSessionName),
		ExternalId:      aws.String(input.ExternalID),
		DurationSeconds: aws.Int64(sessionDuration),
	}

	result, err := a.stsClient.AssumeRoleWithContext(ctx, assumeRoleInput)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrAssumeRoleFailed, err)
	}

	if result.Credentials == nil {
		return nil, fmt.Errorf("%w: no credentials returned", ErrAssumeRoleFailed)
	}

	return &Credentials{
		AccessKeyID:     *result.Credentials.AccessKeyId,
		SecretAccessKey: *result.Credentials.SecretAccessKey,
		SessionToken:    *result.Credentials.SessionToken,
		Expiration:      *result.Credentials.Expiration,
	}, nil
}

// VerifyAccountAccess verifies access to an AWS account and returns account info
func (a *AWSAuth) VerifyAccountAccess(ctx context.Context, input AssumeRoleInput) (*AccountInfo, error) {
	if !a.selfHosted {
		// Production mode: Verify by assuming the role
		creds, err := a.AssumeRole(ctx, input)
		if err != nil {
			return nil, err
		}

		// Create a new STS client with assumed role credentials
		config := &aws.Config{
			Region:      aws.String(a.region),
			Credentials: credentials.NewStaticCredentials(creds.AccessKeyID, creds.SecretAccessKey, creds.SessionToken),
		}

		if a.endpointURL != "" {
			config.Endpoint = aws.String(a.endpointURL)
		}

		sess, err := session.NewSession(config)
		if err != nil {
			return nil, fmt.Errorf("failed to create session with assumed role: %w", err)
		}

		stsTempClient := sts.New(sess)
		identity, err := stsTempClient.GetCallerIdentityWithContext(ctx, &sts.GetCallerIdentityInput{})
		if err != nil {
			return nil, fmt.Errorf("failed to get caller identity: %w", err)
		}

		return &AccountInfo{
			AccountID: *identity.Account,
			ARN:       *identity.Arn,
			UserID:    *identity.UserId,
		}, nil
	}

	// Self-hosted mode: Use existing credentials
	identity, err := a.stsClient.GetCallerIdentityWithContext(ctx, &sts.GetCallerIdentityInput{})
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidCredentials, err)
	}

	return &AccountInfo{
		AccountID: *identity.Account,
		ARN:       *identity.Arn,
		UserID:    *identity.UserId,
	}, nil
}

// GetAccountID retrieves the AWS account ID for direct credentials (self-hosted mode)
func (a *AWSAuth) GetAccountID(ctx context.Context, accessKey, secretKey string) (string, error) {
	config := &aws.Config{
		Credentials: credentials.NewStaticCredentials(accessKey, secretKey, ""),
		Region:      aws.String(a.region),
	}

	if a.endpointURL != "" {
		config.Endpoint = aws.String(a.endpointURL)
	}

	sess, err := session.NewSession(config)
	if err != nil {
		return "", fmt.Errorf("failed to create session: %w", err)
	}

	stsClient := sts.New(sess)
	identity, err := stsClient.GetCallerIdentityWithContext(ctx, &sts.GetCallerIdentityInput{})
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrInvalidCredentials, err)
	}

	return *identity.Account, nil
}
