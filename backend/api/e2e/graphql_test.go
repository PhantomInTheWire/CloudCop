package e2e

import (
	"context"
	"testing"
	"time"

	"cloudcop/api/graph"
	"cloudcop/api/internal/scanner"
	"cloudcop/api/internal/scanner/s3"
	"cloudcop/api/internal/security"

	"github.com/aws/aws-sdk-go-v2/aws"
)

// TestGraphQL_StartScan_E2E tests the GraphQL StartScan resolver end-to-end
func TestGraphQL_StartScan_E2E(t *testing.T) {
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

	// 1. Setup Security Service
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

	// Register scanners (reuse S3 for simplicity)
	svc.RegisterScanner("s3", func(cfg aws.Config, region, accountID string) scanner.ServiceScanner {
		return s3.NewScanner(cfg, region, accountID)
	})

	// 2. Setup Bucket
	s3Client, err := cfg.NewS3Client(ctx)
	if err != nil {
		t.Fatalf("Failed to create S3 client: %v", err)
	}
	bucketName := "graphql-test-bucket-" + time.Now().Format("150405")
	createMisconfiguredBucket(ctx, t, s3Client, bucketName)
	defer cleanupBucket(ctx, s3Client, bucketName)

	// 3. Create Resolver
	resolver := &graph.Resolver{
		Security: svc,
	}

	// 4. Invoke StartScan via Resolver
	mutation := resolver.Mutation()

	t.Log("Invoking StartScan mutation...")
	scan, err := mutation.StartScan(ctx, TestAccountID, []string{"s3"}, []string{DefaultRegion})
	if err != nil {
		t.Fatalf("StartScan failed: %v", err)
	}

	if scan.Status != "completed" {
		t.Errorf("Expected status completed, got %s", scan.Status)
	}

	// 5. Verify Summary via Resolver
	scanResolver := resolver.Scan()
	summary, err := scanResolver.Summary(ctx, scan)
	if err != nil {
		t.Fatalf("Failed to get summary: %v", err)
	}

	if summary == nil {
		t.Log("Warning: No summary returned (AI service might be down)")
	} else {
		t.Logf("GraphQL Summary Risk Score: %d", summary.RiskScore)
		t.Logf("GraphQL Summary Text: %s", summary.SummaryText)
		if len(summary.Actions) > 0 {
			t.Logf("GraphQL Remediation Commands: %v", summary.Actions[0].Commands)
		}
	}
}
