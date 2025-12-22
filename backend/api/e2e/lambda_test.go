package e2e

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"testing"
	"time"

	"cloudcop/api/internal/scanner"
	"cloudcop/api/internal/scanner/lambda"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsiam "github.com/aws/aws-sdk-go-v2/service/iam"
	awslambda "github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
)

// TestLambdaScanner_E2E tests the Lambda scanner against LocalStack
func TestLambdaScanner_E2E(t *testing.T) {
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

	lambdaClient, err := cfg.NewLambdaClient(ctx)
	if err != nil {
		t.Fatalf("Failed to create Lambda client: %v", err)
	}

	iamClient, err := cfg.NewIAMClient(ctx)
	if err != nil {
		t.Fatalf("Failed to create IAM client: %v", err)
	}

	// Create a basic Lambda execution role
	roleName := "lambda-test-role-" + time.Now().Format("150405")
	roleArn, cleanup := createLambdaRole(ctx, t, iamClient, roleName)
	defer cleanup()

	// Wait for role to be available
	time.Sleep(2 * time.Second)

	tests := []struct {
		name           string
		setup          func(t *testing.T) (functionName string, cleanup func())
		expectedChecks map[string]scanner.FindingStatus
	}{
		{
			name: "lambda_with_env_secrets",
			setup: func(t *testing.T) (string, func()) {
				functionName := "test-lambda-secrets-" + time.Now().Format("150405")

				// Create minimal Lambda deployment package
				zipContent := createDummyLambdaZip(t)

				// Create function with secrets in environment
				_, err := lambdaClient.CreateFunction(ctx, &awslambda.CreateFunctionInput{
					FunctionName: aws.String(functionName),
					Runtime:      types.RuntimePython312,
					Role:         aws.String(roleArn),
					Handler:      aws.String("handler.handler"),
					Code: &types.FunctionCode{
						ZipFile: zipContent,
					},
					Environment: &types.Environment{
						Variables: map[string]string{
							"API_KEY":           "sk-secret-key-12345",
							"AWS_SECRET_KEY":    "AKIAIOSFODNN7EXAMPLE",
							"DATABASE_PASSWORD": "SuperSecret123!",
						},
					},
					Timeout:    aws.Int32(30),
					MemorySize: aws.Int32(128),
				})
				if err != nil {
					t.Fatalf("Failed to create Lambda function: %v", err)
				}

				return functionName, func() {
					_, _ = lambdaClient.DeleteFunction(ctx, &awslambda.DeleteFunctionInput{
						FunctionName: aws.String(functionName),
					})
				}
			},
			expectedChecks: map[string]scanner.FindingStatus{
				"lambda_env_secrets": scanner.StatusFail,
			},
		},
		{
			name: "lambda_without_dlq",
			setup: func(t *testing.T) (string, func()) {
				functionName := "test-lambda-no-dlq-" + time.Now().Format("150405")

				zipContent := createDummyLambdaZip(t)

				// Create function without DLQ
				_, err := lambdaClient.CreateFunction(ctx, &awslambda.CreateFunctionInput{
					FunctionName: aws.String(functionName),
					Runtime:      types.RuntimePython312,
					Role:         aws.String(roleArn),
					Handler:      aws.String("handler.handler"),
					Code: &types.FunctionCode{
						ZipFile: zipContent,
					},
					Timeout:    aws.Int32(30),
					MemorySize: aws.Int32(128),
				})
				if err != nil {
					t.Fatalf("Failed to create Lambda function: %v", err)
				}

				return functionName, func() {
					_, _ = lambdaClient.DeleteFunction(ctx, &awslambda.DeleteFunctionInput{
						FunctionName: aws.String(functionName),
					})
				}
			},
			expectedChecks: map[string]scanner.FindingStatus{
				"lambda_dlq": scanner.StatusFail,
			},
		},
		{
			name: "lambda_short_timeout",
			setup: func(t *testing.T) (string, func()) {
				functionName := "test-lambda-short-timeout-" + time.Now().Format("150405")

				zipContent := createDummyLambdaZip(t)

				// Create function with very short timeout
				_, err := lambdaClient.CreateFunction(ctx, &awslambda.CreateFunctionInput{
					FunctionName: aws.String(functionName),
					Runtime:      types.RuntimePython312,
					Role:         aws.String(roleArn),
					Handler:      aws.String("handler.handler"),
					Code: &types.FunctionCode{
						ZipFile: zipContent,
					},
					Timeout:    aws.Int32(3), // Very short timeout
					MemorySize: aws.Int32(128),
				})
				if err != nil {
					t.Fatalf("Failed to create Lambda function: %v", err)
				}

				return functionName, func() {
					_, _ = lambdaClient.DeleteFunction(ctx, &awslambda.DeleteFunctionInput{
						FunctionName: aws.String(functionName),
					})
				}
			},
			expectedChecks: map[string]scanner.FindingStatus{
				"lambda_timeout": scanner.StatusFail,
			},
		},
		{
			name: "lambda_no_tracing",
			setup: func(t *testing.T) (string, func()) {
				functionName := "test-lambda-no-tracing-" + time.Now().Format("150405")

				zipContent := createDummyLambdaZip(t)

				// Create function without X-Ray tracing
				_, err := lambdaClient.CreateFunction(ctx, &awslambda.CreateFunctionInput{
					FunctionName: aws.String(functionName),
					Runtime:      types.RuntimePython312,
					Role:         aws.String(roleArn),
					Handler:      aws.String("handler.handler"),
					Code: &types.FunctionCode{
						ZipFile: zipContent,
					},
					Timeout:    aws.Int32(30),
					MemorySize: aws.Int32(128),
					TracingConfig: &types.TracingConfig{
						Mode: types.TracingModePassThrough,
					},
				})
				if err != nil {
					t.Fatalf("Failed to create Lambda function: %v", err)
				}

				return functionName, func() {
					_, _ = lambdaClient.DeleteFunction(ctx, &awslambda.DeleteFunctionInput{
						FunctionName: aws.String(functionName),
					})
				}
			},
			expectedChecks: map[string]scanner.FindingStatus{
				"lambda_tracing": scanner.StatusFail,
			},
		},
		{
			name: "lambda_with_tracing",
			setup: func(t *testing.T) (string, func()) {
				functionName := "test-lambda-with-tracing-" + time.Now().Format("150405")

				zipContent := createDummyLambdaZip(t)

				// Create function with X-Ray tracing enabled
				_, err := lambdaClient.CreateFunction(ctx, &awslambda.CreateFunctionInput{
					FunctionName: aws.String(functionName),
					Runtime:      types.RuntimePython312,
					Role:         aws.String(roleArn),
					Handler:      aws.String("handler.handler"),
					Code: &types.FunctionCode{
						ZipFile: zipContent,
					},
					Timeout:    aws.Int32(30),
					MemorySize: aws.Int32(128),
					TracingConfig: &types.TracingConfig{
						Mode: types.TracingModeActive,
					},
				})
				if err != nil {
					t.Fatalf("Failed to create Lambda function: %v", err)
				}

				return functionName, func() {
					_, _ = lambdaClient.DeleteFunction(ctx, &awslambda.DeleteFunctionInput{
						FunctionName: aws.String(functionName),
					})
				}
			},
			expectedChecks: map[string]scanner.FindingStatus{
				"lambda_tracing": scanner.StatusPass,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			functionName, cleanup := tt.setup(t)
			defer cleanup()

			// Wait for function to be fully created
			time.Sleep(1 * time.Second)

			// Run scanner
			lambdaScanner := lambda.NewScanner(awsCfg, DefaultRegion, TestAccountID)
			findings, err := lambdaScanner.Scan(ctx, DefaultRegion)
			if err != nil {
				t.Fatalf("Scan failed: %v", err)
			}

			// Filter findings for our function
			functionFindings := filterLambdaFindings(findings, functionName)

			t.Logf("Found %d findings for function %s", len(functionFindings), functionName)
			for _, f := range functionFindings {
				t.Logf("  %s: %s (%s)", f.CheckID, f.Status, f.Title)
			}

			// Verify expected checks
			for checkID, expectedStatus := range tt.expectedChecks {
				finding := findFindingByCheckID(functionFindings, checkID)
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

// TestLambdaScanner_MultipleFunctions tests scanning multiple Lambda functions
func TestLambdaScanner_MultipleFunctions(t *testing.T) {
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

	lambdaClient, err := cfg.NewLambdaClient(ctx)
	if err != nil {
		t.Fatalf("Failed to create Lambda client: %v", err)
	}

	iamClient, err := cfg.NewIAMClient(ctx)
	if err != nil {
		t.Fatalf("Failed to create IAM client: %v", err)
	}

	// Create role
	roleName := "lambda-multi-test-role-" + time.Now().Format("150405")
	roleArn, cleanup := createLambdaRole(ctx, t, iamClient, roleName)
	defer cleanup()
	time.Sleep(2 * time.Second)

	// Create multiple functions
	functionNames := []string{
		"multi-func-1-" + time.Now().Format("150405"),
		"multi-func-2-" + time.Now().Format("150405"),
		"multi-func-3-" + time.Now().Format("150405"),
	}

	zipContent := createDummyLambdaZip(t)

	for _, name := range functionNames {
		_, err := lambdaClient.CreateFunction(ctx, &awslambda.CreateFunctionInput{
			FunctionName: aws.String(name),
			Runtime:      types.RuntimePython312,
			Role:         aws.String(roleArn),
			Handler:      aws.String("handler.handler"),
			Code: &types.FunctionCode{
				ZipFile: zipContent,
			},
			Timeout:    aws.Int32(30),
			MemorySize: aws.Int32(128),
		})
		if err != nil {
			t.Fatalf("Failed to create function %s: %v", name, err)
		}
		defer func(n string) {
			_, _ = lambdaClient.DeleteFunction(ctx, &awslambda.DeleteFunctionInput{
				FunctionName: aws.String(n),
			})
		}(name)
	}

	time.Sleep(1 * time.Second)

	// Run scanner
	lambdaScanner := lambda.NewScanner(awsCfg, DefaultRegion, TestAccountID)
	findings, err := lambdaScanner.Scan(ctx, DefaultRegion)
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	// Verify we got findings for all functions
	for _, name := range functionNames {
		funcFindings := filterLambdaFindings(findings, name)
		if len(funcFindings) == 0 {
			t.Errorf("No findings for function %s", name)
		} else {
			t.Logf("Function %s: %d findings", name, len(funcFindings))
		}
	}
}

// Helper functions

func createLambdaRole(ctx context.Context, t *testing.T, client *awsiam.Client, roleName string) (string, func()) {
	assumeRolePolicy := map[string]interface{}{
		"Version": "2012-10-17",
		"Statement": []map[string]interface{}{
			{
				"Effect":    "Allow",
				"Principal": map[string]string{"Service": "lambda.amazonaws.com"},
				"Action":    "sts:AssumeRole",
			},
		},
	}
	assumeRolePolicyJSON, _ := json.Marshal(assumeRolePolicy)

	roleOutput, err := client.CreateRole(ctx, &awsiam.CreateRoleInput{
		RoleName:                 aws.String(roleName),
		AssumeRolePolicyDocument: aws.String(string(assumeRolePolicyJSON)),
	})
	if err != nil {
		t.Fatalf("Failed to create role: %v", err)
	}
	roleArn := aws.ToString(roleOutput.Role.Arn)

	return roleArn, func() {
		_, _ = client.DeleteRole(ctx, &awsiam.DeleteRoleInput{
			RoleName: aws.String(roleName),
		})
	}
}

func createDummyLambdaZip(t *testing.T) []byte {
	var buf bytes.Buffer
	zipWriter := zip.NewWriter(&buf)

	// Add a simple Python handler
	handler, err := zipWriter.Create("handler.py")
	if err != nil {
		t.Fatalf("Failed to create zip entry: %v", err)
	}
	_, err = handler.Write([]byte(`def handler(event, context):
    return {"statusCode": 200, "body": "Hello"}
`))
	if err != nil {
		t.Fatalf("Failed to write handler: %v", err)
	}

	if err := zipWriter.Close(); err != nil {
		t.Fatalf("Failed to close zip: %v", err)
	}

	return buf.Bytes()
}

func filterLambdaFindings(findings []scanner.Finding, functionName string) []scanner.Finding {
	var filtered []scanner.Finding
	for _, f := range findings {
		// Lambda function names are used as resource IDs, or ARNs containing the name
		if f.ResourceID == functionName || contains(f.ResourceID, functionName) {
			filtered = append(filtered, f)
		}
	}
	return filtered
}

func contains(s, substr string) bool {
	return bytes.Contains([]byte(s), []byte(substr))
}
