// Package compliance provides compliance framework mappings for security checks.
package compliance

// Framework represents a compliance framework.
type Framework string

const (
	// CIS is the CIS AWS Foundations Benchmark.
	CIS Framework = "CIS"
	// SOC2 is the SOC 2 Type II framework.
	SOC2 Framework = "SOC2"
	// GDPR is the General Data Protection Regulation.
	GDPR Framework = "GDPR"
	// NIST is the NIST 800-53 framework.
	NIST Framework = "NIST"
	// PCIDSS is the Payment Card Industry Data Security Standard.
	PCIDSS Framework = "PCI-DSS"
)

// CheckMappings maps check IDs to their compliance framework requirements.
var CheckMappings = map[string][]string{
	// S3 Checks
	"s3_bucket_public_access": {"CIS-2.1.5", "SOC2-CC6.1", "NIST-AC-3", "PCI-DSS-1.3"},
	"s3_bucket_policy_public": {"CIS-2.1.5", "SOC2-CC6.1", "NIST-AC-3", "PCI-DSS-1.3"},
	"s3_bucket_encryption":    {"CIS-2.1.1", "SOC2-CC6.1", "NIST-SC-13", "PCI-DSS-3.4", "GDPR-32"},
	"s3_bucket_versioning":    {"CIS-2.1.3", "SOC2-CC6.1", "NIST-CP-9"},
	"s3_bucket_logging":       {"CIS-2.1.2", "SOC2-CC7.2", "NIST-AU-2", "PCI-DSS-10.1"},
	"s3_block_public_access":  {"CIS-2.1.4", "SOC2-CC6.1", "NIST-AC-3", "PCI-DSS-1.3"},
	"s3_mfa_delete":           {"CIS-2.1.3", "SOC2-CC6.1", "NIST-IA-2"},
	"s3_lifecycle_policy":     {"SOC2-CC6.1", "NIST-SI-12"},
	"s3_ssl_only":             {"CIS-2.1.2", "SOC2-CC6.7", "NIST-SC-8", "PCI-DSS-4.1"},
	"s3_object_lock":          {"SOC2-CC6.1", "NIST-CP-9"},

	// EC2 Checks
	"ec2_sg_unrestricted_ingress": {"CIS-5.1", "SOC2-CC6.1", "NIST-AC-4", "PCI-DSS-1.2"},
	"ec2_sg_dangerous_ports":      {"CIS-5.2", "SOC2-CC6.1", "NIST-AC-4", "PCI-DSS-1.2"},
	"ec2_imdsv2_required":         {"CIS-5.6", "SOC2-CC6.1", "NIST-AC-3"},
	"ec2_ebs_encryption":          {"CIS-2.2.1", "SOC2-CC6.1", "NIST-SC-28", "PCI-DSS-3.4", "GDPR-32"},
	"ec2_public_ip":               {"SOC2-CC6.1", "NIST-AC-4"},
	"ec2_cloudwatch_monitoring":   {"CIS-4.1", "SOC2-CC7.2", "NIST-AU-2"},
	"ec2_detailed_monitoring":     {"SOC2-CC7.2", "NIST-AU-6"},
	"ec2_iam_role":                {"CIS-4.2", "SOC2-CC6.3", "NIST-AC-6"},
	"ec2_unassociated_eip":        {"SOC2-CC6.1", "NIST-CM-8"},
	"ec2_unused_sg_rules":         {"SOC2-CC6.1", "NIST-CM-2"},
	"ec2_vpc_flow_logs":           {"CIS-3.7", "SOC2-CC7.2", "NIST-AU-2", "PCI-DSS-10.1"},
	"ec2_imdsv1_usage":            {"CIS-5.6", "SOC2-CC6.1", "NIST-AC-3"},

	// IAM Checks
	"iam_unused_access_keys":   {"CIS-1.12", "SOC2-CC6.1", "NIST-AC-2"},
	"iam_access_key_rotation":  {"CIS-1.14", "SOC2-CC6.1", "NIST-IA-5", "PCI-DSS-8.2"},
	"iam_root_usage":           {"CIS-1.7", "SOC2-CC6.1", "NIST-AC-6", "PCI-DSS-8.1"},
	"iam_user_mfa":             {"CIS-1.10", "SOC2-CC6.1", "NIST-IA-2", "PCI-DSS-8.3"},
	"iam_root_mfa":             {"CIS-1.5", "SOC2-CC6.1", "NIST-IA-2", "PCI-DSS-8.3"},
	"iam_overly_permissive":    {"CIS-1.16", "SOC2-CC6.1", "NIST-AC-6", "PCI-DSS-7.1"},
	"iam_privilege_escalation": {"SOC2-CC6.1", "NIST-AC-6"},
	"iam_password_policy":      {"CIS-1.8", "SOC2-CC6.1", "NIST-IA-5", "PCI-DSS-8.2"},
	"iam_unused_users":         {"CIS-1.12", "SOC2-CC6.1", "NIST-AC-2"},
	"iam_inline_policies":      {"CIS-1.16", "SOC2-CC6.1", "NIST-AC-6"},
	"iam_cross_account_trust":  {"SOC2-CC6.1", "NIST-AC-4"},
	"iam_service_role_trust":   {"SOC2-CC6.1", "NIST-AC-6"},
	"iam_admin_access_users":   {"CIS-1.16", "SOC2-CC6.1", "NIST-AC-6", "PCI-DSS-7.1"},
	"iam_not_action":           {"SOC2-CC6.1", "NIST-AC-6"},
	"iam_console_without_mfa":  {"CIS-1.10", "SOC2-CC6.1", "NIST-IA-2", "PCI-DSS-8.3"},

	// Lambda Checks
	"lambda_env_secrets":          {"SOC2-CC6.1", "NIST-SC-28", "PCI-DSS-3.4", "GDPR-32"},
	"lambda_excessive_iam":        {"SOC2-CC6.1", "NIST-AC-6"},
	"lambda_cloudwatch_logs":      {"SOC2-CC7.2", "NIST-AU-2"},
	"lambda_vpc_config":           {"SOC2-CC6.1", "NIST-AC-4"},
	"lambda_dlq":                  {"SOC2-CC7.1", "NIST-SI-2"},
	"lambda_tracing":              {"SOC2-CC7.2", "NIST-AU-6"},
	"lambda_reserved_concurrency": {"SOC2-CC6.1", "NIST-SC-5"},
	"lambda_timeout":              {"SOC2-CC7.1", "NIST-SI-2"},

	// ECS Checks
	"ecs_privileged_container": {"CIS-5.1", "SOC2-CC6.1", "NIST-AC-6"},
	"ecs_public_registry":      {"SOC2-CC6.1", "NIST-SA-12"},
	"ecs_task_iam_role":        {"SOC2-CC6.1", "NIST-AC-6"},
	"ecs_awsvpc_mode":          {"SOC2-CC6.1", "NIST-AC-4"},
	"ecs_secrets_in_env":       {"SOC2-CC6.1", "NIST-SC-28", "PCI-DSS-3.4"},
	"ecs_cloudwatch_logs":      {"SOC2-CC7.2", "NIST-AU-2"},
	"ecs_task_versioning":      {"SOC2-CC8.1", "NIST-CM-3"},
	"ecs_auto_scaling":         {"SOC2-CC7.1", "NIST-CP-10"},

	// DynamoDB Checks
	"dynamodb_encryption":   {"CIS-2.3.1", "SOC2-CC6.1", "NIST-SC-28", "PCI-DSS-3.4", "GDPR-32"},
	"dynamodb_pitr":         {"SOC2-CC6.1", "NIST-CP-9"},
	"dynamodb_backup":       {"SOC2-CC6.1", "NIST-CP-9"},
	"dynamodb_ttl":          {"GDPR-17", "NIST-SI-12"},
	"dynamodb_auto_scaling": {"SOC2-CC7.1", "NIST-CP-10"},
	"dynamodb_vpc_endpoint": {"SOC2-CC6.1", "NIST-AC-4"},
}

// GetCompliance returns the compliance framework codes associated with the given check ID.
// If the check ID is not present in the mappings, it returns an empty slice.
func GetCompliance(checkID string) []string {
	if mappings, exists := CheckMappings[checkID]; exists {
		return mappings
	}
	return []string{}
}