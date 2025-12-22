// Package e2e provides end-to-end testing utilities using LocalStack.
package e2e

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

const (
	// DefaultLocalStackEndpoint is the default LocalStack endpoint
	DefaultLocalStackEndpoint = "http://localhost:4566"
	// DefaultRegion for LocalStack tests
	DefaultRegion = "us-east-1"
	// TestAccountID is the fake AWS account ID for LocalStack
	TestAccountID = "000000000000"
)

// LocalStackConfig holds configuration for connecting to LocalStack
type LocalStackConfig struct {
	Endpoint  string
	Region    string
	AccountID string
}

// NewDefaultConfig creates a default LocalStack configuration
func NewDefaultConfig() *LocalStackConfig {
	endpoint := os.Getenv("LOCALSTACK_ENDPOINT")
	if endpoint == "" {
		endpoint = DefaultLocalStackEndpoint
	}
	return &LocalStackConfig{
		Endpoint:  endpoint,
		Region:    DefaultRegion,
		AccountID: TestAccountID,
	}
}

// GetAWSConfig returns an AWS configuration for LocalStack
func (c *LocalStackConfig) GetAWSConfig(ctx context.Context) (aws.Config, error) {
	//nolint:staticcheck // Using deprecated endpoint resolver for LocalStack compatibility
	customResolver := aws.EndpointResolverWithOptionsFunc(
		func(_, _ string, _ ...interface{}) (aws.Endpoint, error) {
			return aws.Endpoint{
				URL:               c.Endpoint,
				HostnameImmutable: true,
				SigningRegion:     c.Region,
			}, nil
		})

	//nolint:staticcheck // Using deprecated endpoint resolver for LocalStack compatibility
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(c.Region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			"test",
			"test",
			"",
		)),
		config.WithEndpointResolverWithOptions(customResolver),
	)
	if err != nil {
		return aws.Config{}, fmt.Errorf("loading AWS config: %w", err)
	}

	return cfg, nil
}

// NewS3Client creates an S3 client configured for LocalStack
func (c *LocalStackConfig) NewS3Client(ctx context.Context) (*s3.Client, error) {
	cfg, err := c.GetAWSConfig(ctx)
	if err != nil {
		return nil, err
	}
	return s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.UsePathStyle = true // Required for LocalStack
	}), nil
}

// NewEC2Client creates an EC2 client configured for LocalStack
func (c *LocalStackConfig) NewEC2Client(ctx context.Context) (*ec2.Client, error) {
	cfg, err := c.GetAWSConfig(ctx)
	if err != nil {
		return nil, err
	}
	return ec2.NewFromConfig(cfg), nil
}

// NewIAMClient creates an IAM client configured for LocalStack
func (c *LocalStackConfig) NewIAMClient(ctx context.Context) (*iam.Client, error) {
	cfg, err := c.GetAWSConfig(ctx)
	if err != nil {
		return nil, err
	}
	return iam.NewFromConfig(cfg), nil
}

// NewLambdaClient creates a Lambda client configured for LocalStack
func (c *LocalStackConfig) NewLambdaClient(ctx context.Context) (*lambda.Client, error) {
	cfg, err := c.GetAWSConfig(ctx)
	if err != nil {
		return nil, err
	}
	return lambda.NewFromConfig(cfg), nil
}

// NewDynamoDBClient creates a DynamoDB client configured for LocalStack
func (c *LocalStackConfig) NewDynamoDBClient(ctx context.Context) (*dynamodb.Client, error) {
	cfg, err := c.GetAWSConfig(ctx)
	if err != nil {
		return nil, err
	}
	return dynamodb.NewFromConfig(cfg), nil
}

// NewSTSClient creates an STS client configured for LocalStack
func (c *LocalStackConfig) NewSTSClient(ctx context.Context) (*sts.Client, error) {
	cfg, err := c.GetAWSConfig(ctx)
	if err != nil {
		return nil, err
	}
	return sts.NewFromConfig(cfg), nil
}

// StartLocalStack starts LocalStack using docker-compose
func StartLocalStack(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "docker", "compose", "-f", "docker-compose.yml", "up", "-d", "--wait")
	cmd.Dir = getE2EDir()
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("starting LocalStack: %w", err)
	}
	return nil
}

// StopLocalStack stops LocalStack using docker-compose
func StopLocalStack(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "docker", "compose", "-f", "docker-compose.yml", "down", "-v")
	cmd.Dir = getE2EDir()
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// WaitForLocalStack waits for LocalStack to be healthy
func WaitForLocalStack(ctx context.Context, timeout time.Duration) error {
	cfg := NewDefaultConfig()
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		stsClient, err := cfg.NewSTSClient(ctx)
		if err == nil {
			_, err = stsClient.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
			if err == nil {
				return nil
			}
		}
		time.Sleep(1 * time.Second)
	}

	return fmt.Errorf("LocalStack not ready after %v", timeout)
}

// IsLocalStackRunning checks if LocalStack is running
func IsLocalStackRunning(ctx context.Context) bool {
	cfg := NewDefaultConfig()
	stsClient, err := cfg.NewSTSClient(ctx)
	if err != nil {
		return false
	}
	_, err = stsClient.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
	return err == nil
}

func getE2EDir() string {
	// Try to find the e2e directory relative to current working directory
	wd, _ := os.Getwd()
	if strings.HasSuffix(wd, "e2e") {
		return wd
	}
	if strings.HasSuffix(wd, "api") {
		return wd + "/e2e"
	}
	return "."
}
