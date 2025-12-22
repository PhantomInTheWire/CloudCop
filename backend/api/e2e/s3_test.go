// Package e2e provides end-to-end tests for CloudCop scanners using LocalStack.
package e2e

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"cloudcop/api/internal/scanner"
	"cloudcop/api/internal/scanner/s3"

	"github.com/aws/aws-sdk-go-v2/aws"
	awss3 "github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// TestS3Scanner_E2E tests the S3 scanner against LocalStack
func TestS3Scanner_E2E(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Check if LocalStack is running
	if !IsLocalStackRunning(ctx) {
		t.Skip("LocalStack is not running. Start it with: docker compose -f e2e/docker-compose.yml up -d")
	}

	cfg := NewDefaultConfig()
	awsCfg, err := cfg.GetAWSConfig(ctx)
	if err != nil {
		t.Fatalf("Failed to get AWS config: %v", err)
	}

	// Create S3 client for setup
	s3Client, err := cfg.NewS3Client(ctx)
	if err != nil {
		t.Fatalf("Failed to create S3 client: %v", err)
	}

	// Test cases with different bucket configurations
	tests := []struct {
		name           string
		setupBucket    func(t *testing.T, ctx context.Context, client *awss3.Client, bucketName string)
		expectedChecks map[string]scanner.FindingStatus // checkID -> expected status
	}{
		{
			name: "unencrypted_bucket",
			setupBucket: func(t *testing.T, ctx context.Context, client *awss3.Client, bucketName string) {
				// Create a basic bucket with no encryption
				_, err := client.CreateBucket(ctx, &awss3.CreateBucketInput{
					Bucket: aws.String(bucketName),
				})
				if err != nil {
					t.Fatalf("Failed to create bucket: %v", err)
				}
			},
			expectedChecks: map[string]scanner.FindingStatus{
				"s3_bucket_encryption": scanner.StatusFail,
				"s3_bucket_versioning": scanner.StatusFail,
				"s3_bucket_logging":    scanner.StatusFail,
				"s3_mfa_delete":        scanner.StatusFail,
				"s3_lifecycle_policy":  scanner.StatusFail,
				"s3_ssl_only":          scanner.StatusFail,
			},
		},
		{
			name: "encrypted_bucket_with_versioning",
			setupBucket: func(t *testing.T, ctx context.Context, client *awss3.Client, bucketName string) {
				// Create bucket
				_, err := client.CreateBucket(ctx, &awss3.CreateBucketInput{
					Bucket: aws.String(bucketName),
				})
				if err != nil {
					t.Fatalf("Failed to create bucket: %v", err)
				}

				// Enable encryption
				_, err = client.PutBucketEncryption(ctx, &awss3.PutBucketEncryptionInput{
					Bucket: aws.String(bucketName),
					ServerSideEncryptionConfiguration: &types.ServerSideEncryptionConfiguration{
						Rules: []types.ServerSideEncryptionRule{
							{
								ApplyServerSideEncryptionByDefault: &types.ServerSideEncryptionByDefault{
									SSEAlgorithm: types.ServerSideEncryptionAes256,
								},
							},
						},
					},
				})
				if err != nil {
					t.Fatalf("Failed to enable encryption: %v", err)
				}

				// Enable versioning
				_, err = client.PutBucketVersioning(ctx, &awss3.PutBucketVersioningInput{
					Bucket: aws.String(bucketName),
					VersioningConfiguration: &types.VersioningConfiguration{
						Status: types.BucketVersioningStatusEnabled,
					},
				})
				if err != nil {
					t.Fatalf("Failed to enable versioning: %v", err)
				}
			},
			expectedChecks: map[string]scanner.FindingStatus{
				"s3_bucket_encryption": scanner.StatusPass,
				"s3_bucket_versioning": scanner.StatusPass,
			},
		},
		{
			name: "public_bucket",
			setupBucket: func(t *testing.T, ctx context.Context, client *awss3.Client, bucketName string) {
				// Create bucket
				_, err := client.CreateBucket(ctx, &awss3.CreateBucketInput{
					Bucket: aws.String(bucketName),
				})
				if err != nil {
					t.Fatalf("Failed to create bucket: %v", err)
				}

				// Set public ACL (AllUsers)
				_, err = client.PutBucketAcl(ctx, &awss3.PutBucketAclInput{
					Bucket: aws.String(bucketName),
					ACL:    types.BucketCannedACLPublicRead,
				})
				if err != nil {
					// LocalStack may not fully support all ACL operations
					t.Logf("Warning: Failed to set public ACL (may be expected in LocalStack): %v", err)
				}
			},
			expectedChecks: map[string]scanner.FindingStatus{
				// Note: ACL checks may vary based on LocalStack behavior
				"s3_bucket_encryption": scanner.StatusFail,
			},
		},
		{
			name: "bucket_with_block_public_access",
			setupBucket: func(t *testing.T, ctx context.Context, client *awss3.Client, bucketName string) {
				// Create bucket
				_, err := client.CreateBucket(ctx, &awss3.CreateBucketInput{
					Bucket: aws.String(bucketName),
				})
				if err != nil {
					t.Fatalf("Failed to create bucket: %v", err)
				}

				// Enable all Block Public Access settings
				_, err = client.PutPublicAccessBlock(ctx, &awss3.PutPublicAccessBlockInput{
					Bucket: aws.String(bucketName),
					PublicAccessBlockConfiguration: &types.PublicAccessBlockConfiguration{
						BlockPublicAcls:       aws.Bool(true),
						BlockPublicPolicy:     aws.Bool(true),
						IgnorePublicAcls:      aws.Bool(true),
						RestrictPublicBuckets: aws.Bool(true),
					},
				})
				if err != nil {
					t.Fatalf("Failed to set block public access: %v", err)
				}
			},
			expectedChecks: map[string]scanner.FindingStatus{
				"s3_block_public_access": scanner.StatusPass,
			},
		},
		{
			name: "bucket_with_ssl_policy",
			setupBucket: func(t *testing.T, ctx context.Context, client *awss3.Client, bucketName string) {
				// Create bucket
				_, err := client.CreateBucket(ctx, &awss3.CreateBucketInput{
					Bucket: aws.String(bucketName),
				})
				if err != nil {
					t.Fatalf("Failed to create bucket: %v", err)
				}

				// Set bucket policy that enforces SSL
				policy := map[string]interface{}{
					"Version": "2012-10-17",
					"Statement": []map[string]interface{}{
						{
							"Sid":       "ForceSSLOnlyAccess",
							"Effect":    "Deny",
							"Principal": "*",
							"Action":    "s3:*",
							"Resource": []string{
								"arn:aws:s3:::" + bucketName,
								"arn:aws:s3:::" + bucketName + "/*",
							},
							"Condition": map[string]interface{}{
								"Bool": map[string]string{
									"aws:SecureTransport": "false",
								},
							},
						},
					},
				}
				policyJSON, _ := json.Marshal(policy)

				_, err = client.PutBucketPolicy(ctx, &awss3.PutBucketPolicyInput{
					Bucket: aws.String(bucketName),
					Policy: aws.String(string(policyJSON)),
				})
				if err != nil {
					t.Logf("Warning: Failed to set bucket policy: %v", err)
				}
			},
			expectedChecks: map[string]scanner.FindingStatus{
				"s3_ssl_only": scanner.StatusPass,
			},
		},
		{
			name: "bucket_with_logging",
			setupBucket: func(t *testing.T, ctx context.Context, client *awss3.Client, bucketName string) {
				// Create main bucket
				_, err := client.CreateBucket(ctx, &awss3.CreateBucketInput{
					Bucket: aws.String(bucketName),
				})
				if err != nil {
					t.Fatalf("Failed to create bucket: %v", err)
				}

				// Create logging target bucket
				logBucket := bucketName + "-logs"
				_, err = client.CreateBucket(ctx, &awss3.CreateBucketInput{
					Bucket: aws.String(logBucket),
				})
				if err != nil {
					t.Fatalf("Failed to create log bucket: %v", err)
				}

				// Enable logging
				_, err = client.PutBucketLogging(ctx, &awss3.PutBucketLoggingInput{
					Bucket: aws.String(bucketName),
					BucketLoggingStatus: &types.BucketLoggingStatus{
						LoggingEnabled: &types.LoggingEnabled{
							TargetBucket: aws.String(logBucket),
							TargetPrefix: aws.String("logs/"),
						},
					},
				})
				if err != nil {
					t.Logf("Warning: Failed to enable logging (may be expected in LocalStack): %v", err)
				}
			},
			expectedChecks: map[string]scanner.FindingStatus{
				"s3_bucket_logging": scanner.StatusPass,
			},
		},
		{
			name: "bucket_with_lifecycle",
			setupBucket: func(t *testing.T, ctx context.Context, client *awss3.Client, bucketName string) {
				// Create bucket
				_, err := client.CreateBucket(ctx, &awss3.CreateBucketInput{
					Bucket: aws.String(bucketName),
				})
				if err != nil {
					t.Fatalf("Failed to create bucket: %v", err)
				}

				// Add lifecycle rule
				_, err = client.PutBucketLifecycleConfiguration(ctx, &awss3.PutBucketLifecycleConfigurationInput{
					Bucket: aws.String(bucketName),
					LifecycleConfiguration: &types.BucketLifecycleConfiguration{
						Rules: []types.LifecycleRule{
							{
								ID:     aws.String("cleanup"),
								Status: types.ExpirationStatusEnabled,
								Filter: &types.LifecycleRuleFilter{
									Prefix: aws.String("temp/"),
								},
								Expiration: &types.LifecycleExpiration{
									Days: aws.Int32(30),
								},
							},
						},
					},
				})
				if err != nil {
					t.Logf("Warning: Failed to set lifecycle: %v", err)
				}
			},
			expectedChecks: map[string]scanner.FindingStatus{
				"s3_lifecycle_policy": scanner.StatusPass,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create unique bucket name for this test
			bucketName := "test-" + tt.name + "-" + time.Now().Format("20060102150405")

			// Setup the bucket
			tt.setupBucket(t, ctx, s3Client, bucketName)

			// Cleanup after test
			defer cleanupBucket(ctx, s3Client, bucketName)
			defer cleanupBucket(ctx, s3Client, bucketName+"-logs")

			// Create and run the scanner
			s3Scanner := s3.NewScanner(awsCfg, DefaultRegion, TestAccountID)
			findings, err := s3Scanner.Scan(ctx, DefaultRegion)
			if err != nil {
				t.Fatalf("Scan failed: %v", err)
			}

			// Filter findings for our bucket
			bucketFindings := filterFindingsByResource(findings, bucketName)

			// Log all findings for debugging
			t.Logf("Found %d findings for bucket %s", len(bucketFindings), bucketName)
			for _, f := range bucketFindings {
				t.Logf("  %s: %s (%s)", f.CheckID, f.Status, f.Title)
			}

			// Verify expected checks
			for checkID, expectedStatus := range tt.expectedChecks {
				finding := findFindingByCheckID(bucketFindings, checkID)
				if finding == nil {
					t.Errorf("Expected finding for check %s, but not found", checkID)
					continue
				}
				if finding.Status != expectedStatus {
					t.Errorf("Check %s: got status %s, want %s. Description: %s",
						checkID, finding.Status, expectedStatus, finding.Description)
				}
			}
		})
	}
}

// TestS3Scanner_MultipleBuckets tests scanning multiple buckets
func TestS3Scanner_MultipleBuckets(t *testing.T) {
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

	s3Client, err := cfg.NewS3Client(ctx)
	if err != nil {
		t.Fatalf("Failed to create S3 client: %v", err)
	}

	// Create multiple buckets
	bucketNames := []string{
		"multi-test-bucket-1-" + time.Now().Format("150405"),
		"multi-test-bucket-2-" + time.Now().Format("150405"),
		"multi-test-bucket-3-" + time.Now().Format("150405"),
	}

	for _, name := range bucketNames {
		_, err := s3Client.CreateBucket(ctx, &awss3.CreateBucketInput{
			Bucket: aws.String(name),
		})
		if err != nil {
			t.Fatalf("Failed to create bucket %s: %v", name, err)
		}
		defer cleanupBucket(ctx, s3Client, name)
	}

	// Run scanner
	s3Scanner := s3.NewScanner(awsCfg, DefaultRegion, TestAccountID)
	findings, err := s3Scanner.Scan(ctx, DefaultRegion)
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	// Verify we got findings for all buckets
	for _, bucketName := range bucketNames {
		bucketFindings := filterFindingsByResource(findings, bucketName)
		if len(bucketFindings) == 0 {
			t.Errorf("No findings for bucket %s", bucketName)
		} else {
			t.Logf("Bucket %s: %d findings", bucketName, len(bucketFindings))
		}
	}
}

// Helper functions

func cleanupBucket(ctx context.Context, client *awss3.Client, bucketName string) {
	// Delete all objects first
	listOutput, err := client.ListObjectsV2(ctx, &awss3.ListObjectsV2Input{
		Bucket: aws.String(bucketName),
	})
	if err == nil && listOutput.Contents != nil {
		for _, obj := range listOutput.Contents {
			_, _ = client.DeleteObject(ctx, &awss3.DeleteObjectInput{
				Bucket: aws.String(bucketName),
				Key:    obj.Key,
			})
		}
	}

	// Delete the bucket
	_, _ = client.DeleteBucket(ctx, &awss3.DeleteBucketInput{
		Bucket: aws.String(bucketName),
	})
}

func filterFindingsByResource(findings []scanner.Finding, resourceID string) []scanner.Finding {
	var filtered []scanner.Finding
	for _, f := range findings {
		if f.ResourceID == resourceID {
			filtered = append(filtered, f)
		}
	}
	return filtered
}

func findFindingByCheckID(findings []scanner.Finding, checkID string) *scanner.Finding {
	for _, f := range findings {
		if f.CheckID == checkID {
			return &f
		}
	}
	return nil
}
