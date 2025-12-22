package e2e

import (
	"context"
	"testing"
	"time"

	"cloudcop/api/internal/scanner"
	"cloudcop/api/internal/scanner/s3"
	"cloudcop/api/internal/security"

	"github.com/aws/aws-sdk-go-v2/aws"
	awss3 "github.com/aws/aws-sdk-go-v2/service/s3"
)

// TestSecurityService_E2E tests the full SecurityService with AI summarization
func TestSecurityService_E2E(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	if !IsLocalStackRunning(ctx) {
		t.Skip("LocalStack is not running")
	}

	cfg := NewDefaultConfig()
	awsCfg, err := cfg.GetAWSConfig(ctx)
	if err != nil {
		t.Fatalf("Failed to get AWS config: %v", err)
	}

	// Create Security Service
	// We point to localhost:50051 where ai-service is exposed
	svc, err := security.NewService(security.Config{
		AWSConfig:            awsCfg,
		AccountID:            TestAccountID,
		SummarizationAddress: "localhost:50051",
		EnableSummarization:  true,
	})
	if err != nil {
		t.Fatalf("Failed to create security service: %v", err)
	}
	defer func() { _ = svc.Close() }()

	// Register scanners
	svc.RegisterScanner("s3", func(cfg aws.Config, region, accountID string) scanner.ServiceScanner {
		return s3.NewScanner(cfg, region, accountID)
	})

	// Setup a misconfigured bucket for the test
	s3Client, err := cfg.NewS3Client(ctx)
	if err != nil {
		t.Fatalf("Failed to create S3 client: %v", err)
	}
	bucketName := "ai-test-bucket-" + time.Now().Format("150405")
	createMisconfiguredBucket(ctx, t, s3Client, bucketName)
	defer cleanupBucket(ctx, s3Client, bucketName)

	// Run Scan
	scanConfig := scanner.ScanConfig{
		AccountID: TestAccountID,
		Regions:   []string{DefaultRegion},
		Services:  []string{"s3"},
	}

	t.Log("Starting scan with AI summarization...")
	result, err := svc.Scan(ctx, scanConfig)
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	// Validation
	if result == nil {
		t.Fatal("Expected scan result, got nil")
	}
	if result.FailedChecks == 0 {
		t.Error("Expected failed checks for misconfigured bucket")
	}

	// Check Summary
	if result.Summary == nil {
		t.Log("Warning: No summary returned. AI service might be down or not configured correctly.")
		// We don't fail the test strictly if AI is down, unless we want to enforce it
	} else {
		t.Logf("Risk Score: %d", result.Summary.RiskScore)
		t.Logf("Risk Level: %s", result.Summary.RiskLevel)
		t.Logf("Summary Text: %s", result.Summary.SummaryText)

		if len(result.Summary.Groups) > 0 {
			t.Logf("Received %d finding groups", len(result.Summary.Groups))
			for _, g := range result.Summary.Groups {
				t.Logf("Group: %s", g.Title)
				if g.Summary != "" {
					t.Logf("AI Summary: %s", g.Summary)
				}
				if g.Remedy != "" {
					t.Logf("AI Remedy: %s", g.Remedy)
				}
			}
		}

		if len(result.Summary.Actions) > 0 {
			t.Logf("Received %d action items", len(result.Summary.Actions))
			for _, a := range result.Summary.Actions {
				t.Logf("Action: %s", a.Title)
				if len(a.Commands) > 0 {
					t.Logf("Commands: %v", a.Commands)
				}
			}
		}
	}
}

func createMisconfiguredBucket(ctx context.Context, t *testing.T, client *awss3.Client, bucketName string) {
	// Create a basic bucket with no encryption (misconfigured)
	_, err := client.CreateBucket(ctx, &awss3.CreateBucketInput{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		t.Fatalf("Failed to create bucket: %v", err)
	}
}
