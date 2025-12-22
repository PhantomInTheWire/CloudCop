package e2e

import (
	"context"
	"testing"
	"time"

	"cloudcop/api/internal/scanner"
	"cloudcop/api/internal/scanner/dynamodb"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsdynamodb "github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// TestDynamoDBScanner_E2E tests the DynamoDB scanner against LocalStack
func TestDynamoDBScanner_E2E(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	if !IsLocalStackRunning(ctx) {
		t.Skip("LocalStack is not running. Start it with: docker compose -f e2e/docker-compose.yml up -d")
	}

	cfg := NewDefaultConfig()
	awsCfg, err := cfg.GetAWSConfig(ctx)
	if err != nil {
		t.Fatalf("Failed to get AWS config: %v", err)
	}

	dynamoClient, err := cfg.NewDynamoDBClient(ctx)
	if err != nil {
		t.Fatalf("Failed to create DynamoDB client: %v", err)
	}

	tests := []struct {
		name           string
		setup          func(t *testing.T) (tableName string, cleanup func())
		expectedChecks map[string]scanner.FindingStatus
	}{
		{
			name: "table_without_encryption",
			setup: func(t *testing.T) (string, func()) {
				tableName := "test-table-noenc-" + time.Now().Format("150405")

				// Create table without explicit encryption
				// Note: AWS now encrypts all DynamoDB tables by default, but LocalStack may differ
				_, err := dynamoClient.CreateTable(ctx, &awsdynamodb.CreateTableInput{
					TableName: aws.String(tableName),
					AttributeDefinitions: []types.AttributeDefinition{
						{
							AttributeName: aws.String("id"),
							AttributeType: types.ScalarAttributeTypeS,
						},
					},
					KeySchema: []types.KeySchemaElement{
						{
							AttributeName: aws.String("id"),
							KeyType:       types.KeyTypeHash,
						},
					},
					BillingMode: types.BillingModePayPerRequest,
				})
				if err != nil {
					t.Fatalf("Failed to create table: %v", err)
				}

				// Wait for table to be active
				waitForTable(ctx, dynamoClient, tableName)

				return tableName, func() {
					_, _ = dynamoClient.DeleteTable(ctx, &awsdynamodb.DeleteTableInput{
						TableName: aws.String(tableName),
					})
				}
			},
			expectedChecks: map[string]scanner.FindingStatus{
				// Default encryption should pass, but PITR may fail
				"dynamodb_pitr": scanner.StatusFail,
				"dynamodb_ttl":  scanner.StatusFail,
			},
		},
		{
			name: "table_with_pitr",
			setup: func(t *testing.T) (string, func()) {
				tableName := "test-table-pitr-" + time.Now().Format("150405")

				// Create table
				_, err := dynamoClient.CreateTable(ctx, &awsdynamodb.CreateTableInput{
					TableName: aws.String(tableName),
					AttributeDefinitions: []types.AttributeDefinition{
						{
							AttributeName: aws.String("id"),
							AttributeType: types.ScalarAttributeTypeS,
						},
					},
					KeySchema: []types.KeySchemaElement{
						{
							AttributeName: aws.String("id"),
							KeyType:       types.KeyTypeHash,
						},
					},
					BillingMode: types.BillingModePayPerRequest,
				})
				if err != nil {
					t.Fatalf("Failed to create table: %v", err)
				}

				waitForTable(ctx, dynamoClient, tableName)

				// Enable PITR
				_, err = dynamoClient.UpdateContinuousBackups(ctx, &awsdynamodb.UpdateContinuousBackupsInput{
					TableName: aws.String(tableName),
					PointInTimeRecoverySpecification: &types.PointInTimeRecoverySpecification{
						PointInTimeRecoveryEnabled: aws.Bool(true),
					},
				})
				if err != nil {
					t.Logf("Warning: Failed to enable PITR: %v", err)
				}

				return tableName, func() {
					_, _ = dynamoClient.DeleteTable(ctx, &awsdynamodb.DeleteTableInput{
						TableName: aws.String(tableName),
					})
				}
			},
			expectedChecks: map[string]scanner.FindingStatus{
				"dynamodb_pitr": scanner.StatusPass,
			},
		},
		{
			name: "table_with_ttl",
			setup: func(t *testing.T) (string, func()) {
				tableName := "test-table-ttl-" + time.Now().Format("150405")

				// Create table
				_, err := dynamoClient.CreateTable(ctx, &awsdynamodb.CreateTableInput{
					TableName: aws.String(tableName),
					AttributeDefinitions: []types.AttributeDefinition{
						{
							AttributeName: aws.String("id"),
							AttributeType: types.ScalarAttributeTypeS,
						},
					},
					KeySchema: []types.KeySchemaElement{
						{
							AttributeName: aws.String("id"),
							KeyType:       types.KeyTypeHash,
						},
					},
					BillingMode: types.BillingModePayPerRequest,
				})
				if err != nil {
					t.Fatalf("Failed to create table: %v", err)
				}

				waitForTable(ctx, dynamoClient, tableName)

				// Enable TTL
				_, err = dynamoClient.UpdateTimeToLive(ctx, &awsdynamodb.UpdateTimeToLiveInput{
					TableName: aws.String(tableName),
					TimeToLiveSpecification: &types.TimeToLiveSpecification{
						Enabled:       aws.Bool(true),
						AttributeName: aws.String("expireAt"),
					},
				})
				if err != nil {
					t.Logf("Warning: Failed to enable TTL: %v", err)
				}

				return tableName, func() {
					_, _ = dynamoClient.DeleteTable(ctx, &awsdynamodb.DeleteTableInput{
						TableName: aws.String(tableName),
					})
				}
			},
			expectedChecks: map[string]scanner.FindingStatus{
				"dynamodb_ttl": scanner.StatusPass,
			},
		},
		{
			name: "table_with_encryption_kms",
			setup: func(t *testing.T) (string, func()) {
				tableName := "test-table-kms-" + time.Now().Format("150405")

				// Create table with KMS encryption
				_, err := dynamoClient.CreateTable(ctx, &awsdynamodb.CreateTableInput{
					TableName: aws.String(tableName),
					AttributeDefinitions: []types.AttributeDefinition{
						{
							AttributeName: aws.String("id"),
							AttributeType: types.ScalarAttributeTypeS,
						},
					},
					KeySchema: []types.KeySchemaElement{
						{
							AttributeName: aws.String("id"),
							KeyType:       types.KeyTypeHash,
						},
					},
					BillingMode: types.BillingModePayPerRequest,
					SSESpecification: &types.SSESpecification{
						Enabled: aws.Bool(true),
						SSEType: types.SSETypeKms,
					},
				})
				if err != nil {
					t.Fatalf("Failed to create table: %v", err)
				}

				waitForTable(ctx, dynamoClient, tableName)

				return tableName, func() {
					_, _ = dynamoClient.DeleteTable(ctx, &awsdynamodb.DeleteTableInput{
						TableName: aws.String(tableName),
					})
				}
			},
			expectedChecks: map[string]scanner.FindingStatus{
				"dynamodb_encryption": scanner.StatusPass,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tableName, cleanup := tt.setup(t)
			defer cleanup()

			// Run scanner
			dynamoScanner := dynamodb.NewScanner(awsCfg, DefaultRegion, TestAccountID)
			findings, err := dynamoScanner.Scan(ctx, DefaultRegion)
			if err != nil {
				t.Fatalf("Scan failed: %v", err)
			}

			// Filter findings for our table
			tableFindings := filterFindingsByResource(findings, tableName)

			t.Logf("Found %d findings for table %s", len(tableFindings), tableName)
			for _, f := range tableFindings {
				t.Logf("  %s: %s (%s)", f.CheckID, f.Status, f.Title)
			}

			// Verify expected checks
			for checkID, expectedStatus := range tt.expectedChecks {
				finding := findFindingByCheckID(tableFindings, checkID)
				if finding == nil {
					t.Logf("Note: Check %s not found (may depend on LocalStack support)", checkID)
					continue
				}
				if finding.Status != expectedStatus {
					t.Errorf("Check %s: got status %s, want %s", checkID, finding.Status, expectedStatus)
				}
			}
		})
	}
}

// TestDynamoDBScanner_MultipleTables tests scanning multiple DynamoDB tables
func TestDynamoDBScanner_MultipleTables(t *testing.T) {
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

	dynamoClient, err := cfg.NewDynamoDBClient(ctx)
	if err != nil {
		t.Fatalf("Failed to create DynamoDB client: %v", err)
	}

	// Create multiple tables
	tableNames := []string{
		"multi-table-1-" + time.Now().Format("150405"),
		"multi-table-2-" + time.Now().Format("150405"),
		"multi-table-3-" + time.Now().Format("150405"),
	}

	for _, name := range tableNames {
		_, err := dynamoClient.CreateTable(ctx, &awsdynamodb.CreateTableInput{
			TableName: aws.String(name),
			AttributeDefinitions: []types.AttributeDefinition{
				{
					AttributeName: aws.String("id"),
					AttributeType: types.ScalarAttributeTypeS,
				},
			},
			KeySchema: []types.KeySchemaElement{
				{
					AttributeName: aws.String("id"),
					KeyType:       types.KeyTypeHash,
				},
			},
			BillingMode: types.BillingModePayPerRequest,
		})
		if err != nil {
			t.Fatalf("Failed to create table %s: %v", name, err)
		}
		defer func(n string) {
			_, _ = dynamoClient.DeleteTable(ctx, &awsdynamodb.DeleteTableInput{
				TableName: aws.String(n),
			})
		}(name)
		waitForTable(ctx, dynamoClient, name)
	}

	// Run scanner
	dynamoScanner := dynamodb.NewScanner(awsCfg, DefaultRegion, TestAccountID)
	findings, err := dynamoScanner.Scan(ctx, DefaultRegion)
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	// Verify we got findings for all tables
	for _, tableName := range tableNames {
		tableFindings := filterFindingsByResource(findings, tableName)
		if len(tableFindings) == 0 {
			t.Errorf("No findings for table %s", tableName)
		} else {
			t.Logf("Table %s: %d findings", tableName, len(tableFindings))
		}
	}
}

// Helper function to wait for table to be active
func waitForTable(ctx context.Context, client *awsdynamodb.Client, tableName string) {
	for i := 0; i < 30; i++ {
		output, err := client.DescribeTable(ctx, &awsdynamodb.DescribeTableInput{
			TableName: aws.String(tableName),
		})
		if err == nil && output.Table.TableStatus == types.TableStatusActive {
			return
		}
		time.Sleep(500 * time.Millisecond)
	}
}
