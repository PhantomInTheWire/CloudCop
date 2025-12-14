package compliance

import (
	"testing"
)

func TestGetCompliance(t *testing.T) {
	tests := []struct {
		name    string
		checkID string
		want    []string
		wantLen int
	}{
		{
			name:    "S3 bucket public access check",
			checkID: "s3_bucket_public_access",
			want:    []string{"CIS-2.1.5", "SOC2-CC6.1", "NIST-AC-3", "PCI-DSS-1.3"},
			wantLen: 4,
		},
		{
			name:    "S3 bucket encryption check",
			checkID: "s3_bucket_encryption",
			want:    []string{"CIS-2.1.1", "SOC2-CC6.1", "NIST-SC-13", "PCI-DSS-3.4", "GDPR-32"},
			wantLen: 5,
		},
		{
			name:    "IAM root MFA check",
			checkID: "iam_root_mfa",
			want:    []string{"CIS-1.5", "SOC2-CC6.1", "NIST-IA-2", "PCI-DSS-8.3"},
			wantLen: 4,
		},
		{
			name:    "EC2 IMDSv2 check",
			checkID: "ec2_imdsv2_required",
			want:    []string{"CIS-5.6", "SOC2-CC6.1", "NIST-AC-3"},
			wantLen: 3,
		},
		{
			name:    "Lambda environment secrets check",
			checkID: "lambda_env_secrets",
			want:    []string{"SOC2-CC6.1", "NIST-SC-28", "PCI-DSS-3.4", "GDPR-32"},
			wantLen: 4,
		},
		{
			name:    "ECS privileged container check",
			checkID: "ecs_privileged_container",
			want:    []string{"CIS-5.1", "SOC2-CC6.1", "NIST-AC-6"},
			wantLen: 3,
		},
		{
			name:    "DynamoDB encryption check",
			checkID: "dynamodb_encryption",
			want:    []string{"CIS-2.3.1", "SOC2-CC6.1", "NIST-SC-28", "PCI-DSS-3.4", "GDPR-32"},
			wantLen: 5,
		},
		{
			name:    "Non-existent check returns empty",
			checkID: "non_existent_check",
			want:    []string{},
			wantLen: 0,
		},
		{
			name:    "Empty check ID returns empty",
			checkID: "",
			want:    []string{},
			wantLen: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetCompliance(tt.checkID)
			if len(got) != tt.wantLen {
				t.Errorf("GetCompliance() returned %d items, want %d", len(got), tt.wantLen)
			}
			if tt.wantLen > 0 {
				// Verify expected values are present
				gotMap := make(map[string]bool)
				for _, v := range got {
					gotMap[v] = true
				}
				for _, want := range tt.want {
					if !gotMap[want] {
						t.Errorf("GetCompliance() missing expected value %s", want)
					}
				}
			}
		})
	}
}

func TestCheckMappings_Coverage(t *testing.T) {
	// Test that all expected check IDs have mappings
	expectedChecks := []string{
		// S3
		"s3_bucket_public_access", "s3_bucket_policy_public", "s3_bucket_encryption",
		"s3_bucket_versioning", "s3_bucket_logging", "s3_block_public_access",
		"s3_mfa_delete", "s3_lifecycle_policy", "s3_ssl_only", "s3_object_lock",
		// EC2
		"ec2_sg_unrestricted_ingress", "ec2_sg_dangerous_ports", "ec2_imdsv2_required",
		"ec2_ebs_encryption", "ec2_public_ip", "ec2_cloudwatch_monitoring",
		"ec2_detailed_monitoring", "ec2_iam_role", "ec2_unassociated_eip",
		"ec2_unused_sg_rules", "ec2_vpc_flow_logs", "ec2_imdsv1_usage",
		// IAM
		"iam_unused_access_keys", "iam_access_key_rotation", "iam_root_usage",
		"iam_user_mfa", "iam_root_mfa", "iam_overly_permissive",
		"iam_privilege_escalation", "iam_password_policy", "iam_unused_users",
		"iam_inline_policies", "iam_cross_account_trust", "iam_service_role_trust",
		"iam_admin_access_users", "iam_not_action", "iam_console_without_mfa",
		// Lambda
		"lambda_env_secrets", "lambda_excessive_iam", "lambda_cloudwatch_logs",
		"lambda_vpc_config", "lambda_dlq", "lambda_tracing",
		"lambda_reserved_concurrency", "lambda_timeout",
		// ECS
		"ecs_privileged_container", "ecs_public_registry", "ecs_task_iam_role",
		"ecs_awsvpc_mode", "ecs_secrets_in_env", "ecs_cloudwatch_logs",
		"ecs_task_versioning", "ecs_auto_scaling",
		// DynamoDB
		"dynamodb_encryption", "dynamodb_pitr", "dynamodb_backup",
		"dynamodb_ttl", "dynamodb_auto_scaling", "dynamodb_vpc_endpoint",
	}

	for _, checkID := range expectedChecks {
		t.Run(checkID, func(t *testing.T) {
			mappings := GetCompliance(checkID)
			if len(mappings) == 0 {
				t.Errorf("Check %s has no compliance mappings", checkID)
			}
		})
	}
}

func TestCheckMappings_Frameworks(t *testing.T) {
	// Verify that mappings contain valid framework references
	validPrefixes := []string{"CIS-", "SOC2-", "NIST-", "PCI-DSS-", "GDPR-"}

	for checkID, mappings := range checkMappings {
		for _, mapping := range mappings {
			hasValidPrefix := false
			for _, prefix := range validPrefixes {
				if len(mapping) >= len(prefix) && mapping[:len(prefix)] == prefix {
					hasValidPrefix = true
					break
				}
			}
			if !hasValidPrefix {
				t.Errorf("Check %s has invalid framework mapping: %s", checkID, mapping)
			}
		}
	}
}

func TestFrameworkConstants(t *testing.T) {
	tests := []struct {
		name      string
		framework Framework
		want      string
	}{
		{"CIS framework", CIS, "CIS"},
		{"SOC2 framework", SOC2, "SOC2"},
		{"GDPR framework", GDPR, "GDPR"},
		{"NIST framework", NIST, "NIST"},
		{"PCI-DSS framework", PCIDSS, "PCI-DSS"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.framework) != tt.want {
				t.Errorf("Framework constant = %v, want %v", tt.framework, tt.want)
			}
		})
	}
}

func TestCheckMappings_NoEmptyMappings(t *testing.T) {
	// Ensure no check has empty mappings array
	for checkID, mappings := range checkMappings {
		if len(mappings) == 0 {
			t.Errorf("Check %s has empty mappings array", checkID)
		}
		for i, mapping := range mappings {
			if mapping == "" {
				t.Errorf("Check %s has empty string at index %d", checkID, i)
			}
		}
	}
}

func TestCheckMappings_Consistency(t *testing.T) {
	// Test that similar checks have similar compliance mappings
	tests := []struct {
		name   string
		checks []string
		common string
	}{
		{
			name:   "Encryption checks should reference encryption standards",
			checks: []string{"s3_bucket_encryption", "ec2_ebs_encryption", "dynamodb_encryption"},
			common: "GDPR-32",
		},
		{
			name:   "MFA checks should reference authentication standards",
			checks: []string{"iam_user_mfa", "iam_root_mfa", "iam_console_without_mfa"},
			common: "NIST-IA-2",
		},
		{
			name:   "Logging checks should reference audit standards",
			checks: []string{"s3_bucket_logging", "lambda_cloudwatch_logs", "ecs_cloudwatch_logs"},
			common: "SOC2-CC7.2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, checkID := range tt.checks {
				mappings := GetCompliance(checkID)
				found := false
				for _, mapping := range mappings {
					if mapping == tt.common {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Check %s missing expected common mapping %s", checkID, tt.common)
				}
			}
		})
	}
}
