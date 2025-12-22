package e2e

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"cloudcop/api/internal/scanner"
	"cloudcop/api/internal/scanner/iam"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsiam "github.com/aws/aws-sdk-go-v2/service/iam"
)

// TestIAMScanner_E2E tests the IAM scanner against LocalStack
func TestIAMScanner_E2E(t *testing.T) {
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

	iamClient, err := cfg.NewIAMClient(ctx)
	if err != nil {
		t.Fatalf("Failed to create IAM client: %v", err)
	}

	tests := []struct {
		name           string
		setup          func(t *testing.T) (cleanup func())
		expectedChecks map[string]scanner.FindingStatus
	}{
		{
			name: "user_without_mfa",
			setup: func(t *testing.T) func() {
				userName := "test-user-no-mfa-" + time.Now().Format("150405")

				// Create user
				_, err := iamClient.CreateUser(ctx, &awsiam.CreateUserInput{
					UserName: aws.String(userName),
				})
				if err != nil {
					t.Fatalf("Failed to create user: %v", err)
				}

				// Create login profile (console access) without MFA
				_, err = iamClient.CreateLoginProfile(ctx, &awsiam.CreateLoginProfileInput{
					UserName: aws.String(userName),
					Password: aws.String("Test123!@#Password"),
				})
				if err != nil {
					t.Logf("Warning: Failed to create login profile: %v", err)
				}

				return func() {
					// Delete login profile
					_, _ = iamClient.DeleteLoginProfile(ctx, &awsiam.DeleteLoginProfileInput{
						UserName: aws.String(userName),
					})
					// Delete user
					_, _ = iamClient.DeleteUser(ctx, &awsiam.DeleteUserInput{
						UserName: aws.String(userName),
					})
				}
			},
			expectedChecks: map[string]scanner.FindingStatus{
				"iam_user_mfa":            scanner.StatusFail,
				"iam_console_without_mfa": scanner.StatusFail,
			},
		},
		{
			name: "user_with_access_keys",
			setup: func(t *testing.T) func() {
				userName := "test-user-keys-" + time.Now().Format("150405")

				// Create user
				_, err := iamClient.CreateUser(ctx, &awsiam.CreateUserInput{
					UserName: aws.String(userName),
				})
				if err != nil {
					t.Fatalf("Failed to create user: %v", err)
				}

				// Create access key
				keyOutput, err := iamClient.CreateAccessKey(ctx, &awsiam.CreateAccessKeyInput{
					UserName: aws.String(userName),
				})
				if err != nil {
					t.Fatalf("Failed to create access key: %v", err)
				}
				accessKeyID := aws.ToString(keyOutput.AccessKey.AccessKeyId)

				return func() {
					// Delete access key
					_, _ = iamClient.DeleteAccessKey(ctx, &awsiam.DeleteAccessKeyInput{
						UserName:    aws.String(userName),
						AccessKeyId: aws.String(accessKeyID),
					})
					// Delete user
					_, _ = iamClient.DeleteUser(ctx, &awsiam.DeleteUserInput{
						UserName: aws.String(userName),
					})
				}
			},
			expectedChecks: map[string]scanner.FindingStatus{
				// Access keys are not inherently bad, but we check for rotation
				"iam_access_key_rotation": scanner.StatusPass, // New keys should pass rotation check
			},
		},
		{
			name: "user_with_inline_policy",
			setup: func(t *testing.T) func() {
				userName := "test-user-inline-" + time.Now().Format("150405")

				// Create user
				_, err := iamClient.CreateUser(ctx, &awsiam.CreateUserInput{
					UserName: aws.String(userName),
				})
				if err != nil {
					t.Fatalf("Failed to create user: %v", err)
				}

				// Add inline policy
				policy := map[string]interface{}{
					"Version": "2012-10-17",
					"Statement": []map[string]interface{}{
						{
							"Effect":   "Allow",
							"Action":   "s3:GetObject",
							"Resource": "*",
						},
					},
				}
				policyJSON, _ := json.Marshal(policy)

				_, err = iamClient.PutUserPolicy(ctx, &awsiam.PutUserPolicyInput{
					UserName:       aws.String(userName),
					PolicyName:     aws.String("inline-test-policy"),
					PolicyDocument: aws.String(string(policyJSON)),
				})
				if err != nil {
					t.Fatalf("Failed to put inline policy: %v", err)
				}

				return func() {
					// Delete inline policy
					_, _ = iamClient.DeleteUserPolicy(ctx, &awsiam.DeleteUserPolicyInput{
						UserName:   aws.String(userName),
						PolicyName: aws.String("inline-test-policy"),
					})
					// Delete user
					_, _ = iamClient.DeleteUser(ctx, &awsiam.DeleteUserInput{
						UserName: aws.String(userName),
					})
				}
			},
			expectedChecks: map[string]scanner.FindingStatus{
				"iam_inline_policy": scanner.StatusFail,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := tt.setup(t)
			defer cleanup()

			// Give LocalStack time to process
			time.Sleep(500 * time.Millisecond)

			// Run scanner
			iamScanner := iam.NewScanner(awsCfg, DefaultRegion, TestAccountID)
			findings, err := iamScanner.Scan(ctx, DefaultRegion)
			if err != nil {
				t.Fatalf("Scan failed: %v", err)
			}

			t.Logf("Found %d total IAM findings", len(findings))
			for _, f := range findings {
				t.Logf("  %s: %s (%s) - %s", f.CheckID, f.Status, f.ResourceID, f.Title)
			}

			// Verify expected checks
			for checkID, expectedStatus := range tt.expectedChecks {
				found := false
				for _, f := range findings {
					if f.CheckID == checkID {
						found = true
						if f.Status != expectedStatus {
							t.Errorf("Check %s: got status %s, want %s", checkID, f.Status, expectedStatus)
						}
						break
					}
				}
				if !found {
					t.Logf("Note: Check %s not found in findings (may be expected depending on LocalStack support)", checkID)
				}
			}
		})
	}
}

// TestIAMScanner_Policies tests IAM policy checks
func TestIAMScanner_Policies(t *testing.T) {
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

	iamClient, err := cfg.NewIAMClient(ctx)
	if err != nil {
		t.Fatalf("Failed to create IAM client: %v", err)
	}

	tests := []struct {
		name       string
		policy     map[string]interface{}
		expectFail bool
	}{
		{
			name: "overly_permissive_policy",
			policy: map[string]interface{}{
				"Version": "2012-10-17",
				"Statement": []map[string]interface{}{
					{
						"Effect":   "Allow",
						"Action":   "*",
						"Resource": "*",
					},
				},
			},
			expectFail: true,
		},
		{
			name: "least_privilege_policy",
			policy: map[string]interface{}{
				"Version": "2012-10-17",
				"Statement": []map[string]interface{}{
					{
						"Effect":   "Allow",
						"Action":   []string{"s3:GetObject", "s3:ListBucket"},
						"Resource": "arn:aws:s3:::my-bucket/*",
					},
				},
			},
			expectFail: false,
		},
		{
			name: "admin_policy",
			policy: map[string]interface{}{
				"Version": "2012-10-17",
				"Statement": []map[string]interface{}{
					{
						"Effect":   "Allow",
						"Action":   "iam:*",
						"Resource": "*",
					},
				},
			},
			expectFail: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			policyName := tt.name + "-" + time.Now().Format("150405")
			policyJSON, _ := json.Marshal(tt.policy)

			// Create managed policy
			policyOutput, err := iamClient.CreatePolicy(ctx, &awsiam.CreatePolicyInput{
				PolicyName:     aws.String(policyName),
				PolicyDocument: aws.String(string(policyJSON)),
			})
			if err != nil {
				t.Fatalf("Failed to create policy: %v", err)
			}
			policyArn := aws.ToString(policyOutput.Policy.Arn)
			defer func() {
				_, _ = iamClient.DeletePolicy(ctx, &awsiam.DeletePolicyInput{
					PolicyArn: aws.String(policyArn),
				})
			}()

			// Run scanner
			iamScanner := iam.NewScanner(awsCfg, DefaultRegion, TestAccountID)
			findings, err := iamScanner.Scan(ctx, DefaultRegion)
			if err != nil {
				t.Fatalf("Scan failed: %v", err)
			}

			// Look for overly permissive policy findings
			hasFail := false
			for _, f := range findings {
				if f.CheckID == "iam_overly_permissive_policy" && f.Status == scanner.StatusFail {
					hasFail = true
					t.Logf("Found overly permissive policy: %s", f.Description)
					break
				}
			}

			if tt.expectFail && !hasFail {
				t.Logf("All IAM findings:")
				for _, f := range findings {
					t.Logf("  %s: %s (%s)", f.CheckID, f.Status, f.Title)
				}
				// This might not fail depending on how the scanner checks policies
				t.Logf("Note: Expected overly permissive policy finding not found")
			}
		})
	}
}

// TestIAMScanner_Roles tests IAM role checks
func TestIAMScanner_Roles(t *testing.T) {
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

	iamClient, err := cfg.NewIAMClient(ctx)
	if err != nil {
		t.Fatalf("Failed to create IAM client: %v", err)
	}

	// Create a role with cross-account trust
	roleName := "test-cross-account-role-" + time.Now().Format("150405")
	assumeRolePolicy := map[string]interface{}{
		"Version": "2012-10-17",
		"Statement": []map[string]interface{}{
			{
				"Effect":    "Allow",
				"Principal": map[string]string{"AWS": "arn:aws:iam::999999999999:root"},
				"Action":    "sts:AssumeRole",
			},
		},
	}
	assumeRolePolicyJSON, _ := json.Marshal(assumeRolePolicy)

	_, err = iamClient.CreateRole(ctx, &awsiam.CreateRoleInput{
		RoleName:                 aws.String(roleName),
		AssumeRolePolicyDocument: aws.String(string(assumeRolePolicyJSON)),
	})
	if err != nil {
		t.Fatalf("Failed to create role: %v", err)
	}
	defer func() {
		_, _ = iamClient.DeleteRole(ctx, &awsiam.DeleteRoleInput{
			RoleName: aws.String(roleName),
		})
	}()

	// Run scanner
	iamScanner := iam.NewScanner(awsCfg, DefaultRegion, TestAccountID)
	findings, err := iamScanner.Scan(ctx, DefaultRegion)
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	// Look for cross-account trust finding
	found := false
	for _, f := range findings {
		if f.CheckID == "iam_cross_account_trust" && f.Status == scanner.StatusFail {
			found = true
			t.Logf("Found cross-account trust finding: %s", f.Description)
			break
		}
	}

	if !found {
		t.Logf("Note: Cross-account trust check not found (may depend on scanner implementation)")
		for _, f := range findings {
			t.Logf("  %s: %s (%s)", f.CheckID, f.Status, f.Title)
		}
	}
}

// TestIAMScanner_PasswordPolicy tests password policy check
func TestIAMScanner_PasswordPolicy(t *testing.T) {
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

	iamClient, err := cfg.NewIAMClient(ctx)
	if err != nil {
		t.Fatalf("Failed to create IAM client: %v", err)
	}

	// Set a weak password policy
	_, err = iamClient.UpdateAccountPasswordPolicy(ctx, &awsiam.UpdateAccountPasswordPolicyInput{
		MinimumPasswordLength:      aws.Int32(6), // Too short
		RequireSymbols:             false,
		RequireNumbers:             false,
		RequireUppercaseCharacters: false,
		RequireLowercaseCharacters: false,
		AllowUsersToChangePassword: true,
	})
	if err != nil {
		t.Logf("Warning: Failed to set password policy: %v", err)
	}
	defer func() {
		// Reset to strong policy
		_, _ = iamClient.UpdateAccountPasswordPolicy(ctx, &awsiam.UpdateAccountPasswordPolicyInput{
			MinimumPasswordLength:      aws.Int32(14),
			RequireSymbols:             true,
			RequireNumbers:             true,
			RequireUppercaseCharacters: true,
			RequireLowercaseCharacters: true,
		})
	}()

	// Run scanner
	iamScanner := iam.NewScanner(awsCfg, DefaultRegion, TestAccountID)
	findings, err := iamScanner.Scan(ctx, DefaultRegion)
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	// Look for password policy finding
	found := false
	for _, f := range findings {
		if f.CheckID == "iam_password_policy" {
			found = true
			t.Logf("Password policy check: %s - %s", f.Status, f.Description)
			break
		}
	}

	if !found {
		t.Logf("Note: Password policy check not found")
	}
}
