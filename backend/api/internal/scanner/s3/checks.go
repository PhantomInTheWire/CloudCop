package s3

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"cloudcop/api/internal/scanner"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/smithy-go"
)

const versioningEnabled = "Enabled"

func (s *Scanner) checkPublicAccess(ctx context.Context, bucketName string) []scanner.Finding {
	acl, err := s.client.GetBucketAcl(ctx, &s3.GetBucketAclInput{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		return nil
	}

	for _, grant := range acl.Grants {
		if grant.Grantee != nil && grant.Grantee.URI != nil {
			uri := aws.ToString(grant.Grantee.URI)
			if strings.Contains(uri, "AllUsers") || strings.Contains(uri, "AuthenticatedUsers") {
				return []scanner.Finding{s.createFinding(
					"s3_bucket_public_access",
					bucketName,
					"S3 bucket has public access via ACL",
					fmt.Sprintf("Bucket %s grants access to %s via ACL", bucketName, uri),
					scanner.StatusFail,
					scanner.SeverityCritical,
				)}
			}
		}
	}

	return []scanner.Finding{s.createFinding(
		"s3_bucket_public_access",
		bucketName,
		"S3 bucket ACL does not allow public access",
		fmt.Sprintf("Bucket %s has no public ACL grants", bucketName),
		scanner.StatusPass,
		scanner.SeverityCritical,
	)}
}

func (s *Scanner) checkBucketPolicy(ctx context.Context, bucketName string) []scanner.Finding {
	policyStatus, err := s.client.GetBucketPolicyStatus(ctx, &s3.GetBucketPolicyStatusInput{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		var apiErr smithy.APIError
		if ok := errors.As(err, &apiErr); ok && apiErr.ErrorCode() == "NoSuchBucketPolicy" {
			return []scanner.Finding{s.createFinding(
				"s3_bucket_policy_public",
				bucketName,
				"S3 bucket has no bucket policy",
				fmt.Sprintf("Bucket %s has no bucket policy configured", bucketName),
				scanner.StatusPass,
				scanner.SeverityCritical,
			)}
		}
		return nil
	}

	if policyStatus.PolicyStatus != nil && aws.ToBool(policyStatus.PolicyStatus.IsPublic) {
		return []scanner.Finding{s.createFinding(
			"s3_bucket_policy_public",
			bucketName,
			"S3 bucket policy allows public access",
			fmt.Sprintf("Bucket %s has a public bucket policy", bucketName),
			scanner.StatusFail,
			scanner.SeverityCritical,
		)}
	}

	return []scanner.Finding{s.createFinding(
		"s3_bucket_policy_public",
		bucketName,
		"S3 bucket policy does not allow public access",
		fmt.Sprintf("Bucket %s policy is not public", bucketName),
		scanner.StatusPass,
		scanner.SeverityCritical,
	)}
}

func (s *Scanner) checkEncryption(ctx context.Context, bucketName string) []scanner.Finding {
	encryption, err := s.client.GetBucketEncryption(ctx, &s3.GetBucketEncryptionInput{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		var apiErr smithy.APIError
		if ok := errors.As(err, &apiErr); ok && apiErr.ErrorCode() == "ServerSideEncryptionConfigurationNotFoundError" {
			return []scanner.Finding{s.createFinding(
				"s3_bucket_encryption",
				bucketName,
				"S3 bucket encryption is not enabled",
				fmt.Sprintf("Bucket %s does not have server-side encryption configured", bucketName),
				scanner.StatusFail,
				scanner.SeverityHigh,
			)}
		}
		return nil
	}

	if encryption.ServerSideEncryptionConfiguration != nil && len(encryption.ServerSideEncryptionConfiguration.Rules) > 0 {
		return []scanner.Finding{s.createFinding(
			"s3_bucket_encryption",
			bucketName,
			"S3 bucket encryption is enabled",
			fmt.Sprintf("Bucket %s has server-side encryption configured", bucketName),
			scanner.StatusPass,
			scanner.SeverityHigh,
		)}
	}

	return []scanner.Finding{s.createFinding(
		"s3_bucket_encryption",
		bucketName,
		"S3 bucket encryption is not enabled",
		fmt.Sprintf("Bucket %s does not have encryption rules configured", bucketName),
		scanner.StatusFail,
		scanner.SeverityHigh,
	)}
}

func (s *Scanner) checkVersioning(ctx context.Context, bucketName string) []scanner.Finding {
	versioning, err := s.client.GetBucketVersioning(ctx, &s3.GetBucketVersioningInput{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		return nil
	}

	if versioning.Status == versioningEnabled {
		return []scanner.Finding{s.createFinding(
			"s3_bucket_versioning",
			bucketName,
			"S3 bucket versioning is enabled",
			fmt.Sprintf("Bucket %s has versioning enabled", bucketName),
			scanner.StatusPass,
			scanner.SeverityMedium,
		)}
	}

	return []scanner.Finding{s.createFinding(
		"s3_bucket_versioning",
		bucketName,
		"S3 bucket versioning is not enabled",
		fmt.Sprintf("Bucket %s does not have versioning enabled", bucketName),
		scanner.StatusFail,
		scanner.SeverityMedium,
	)}
}

func (s *Scanner) checkLogging(ctx context.Context, bucketName string) []scanner.Finding {
	logging, err := s.client.GetBucketLogging(ctx, &s3.GetBucketLoggingInput{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		return nil
	}

	if logging.LoggingEnabled != nil {
		return []scanner.Finding{s.createFinding(
			"s3_bucket_logging",
			bucketName,
			"S3 bucket logging is enabled",
			fmt.Sprintf("Bucket %s has access logging enabled", bucketName),
			scanner.StatusPass,
			scanner.SeverityMedium,
		)}
	}

	return []scanner.Finding{s.createFinding(
		"s3_bucket_logging",
		bucketName,
		"S3 bucket logging is not enabled",
		fmt.Sprintf("Bucket %s does not have access logging enabled", bucketName),
		scanner.StatusFail,
		scanner.SeverityMedium,
	)}
}

func (s *Scanner) checkBlockPublicAccess(ctx context.Context, bucketName string) []scanner.Finding {
	publicAccess, err := s.client.GetPublicAccessBlock(ctx, &s3.GetPublicAccessBlockInput{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		return []scanner.Finding{s.createFinding(
			"s3_block_public_access",
			bucketName,
			"S3 bucket Block Public Access is not configured",
			fmt.Sprintf("Bucket %s does not have Block Public Access configured", bucketName),
			scanner.StatusFail,
			scanner.SeverityHigh,
		)}
	}

	config := publicAccess.PublicAccessBlockConfiguration
	if config != nil &&
		aws.ToBool(config.BlockPublicAcls) &&
		aws.ToBool(config.BlockPublicPolicy) &&
		aws.ToBool(config.IgnorePublicAcls) &&
		aws.ToBool(config.RestrictPublicBuckets) {
		return []scanner.Finding{s.createFinding(
			"s3_block_public_access",
			bucketName,
			"S3 bucket Block Public Access is fully enabled",
			fmt.Sprintf("Bucket %s has all Block Public Access settings enabled", bucketName),
			scanner.StatusPass,
			scanner.SeverityHigh,
		)}
	}

	return []scanner.Finding{s.createFinding(
		"s3_block_public_access",
		bucketName,
		"S3 bucket Block Public Access is not fully enabled",
		fmt.Sprintf("Bucket %s has some Block Public Access settings disabled", bucketName),
		scanner.StatusFail,
		scanner.SeverityHigh,
	)}
}

func (s *Scanner) checkMFADelete(ctx context.Context, bucketName string) []scanner.Finding {
	versioning, err := s.client.GetBucketVersioning(ctx, &s3.GetBucketVersioningInput{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		return nil
	}

	if versioning.MFADelete == versioningEnabled {
		return []scanner.Finding{s.createFinding(
			"s3_mfa_delete",
			bucketName,
			"S3 bucket MFA Delete is enabled",
			fmt.Sprintf("Bucket %s has MFA Delete enabled", bucketName),
			scanner.StatusPass,
			scanner.SeverityHigh,
		)}
	}

	return []scanner.Finding{s.createFinding(
		"s3_mfa_delete",
		bucketName,
		"S3 bucket MFA Delete is not enabled",
		fmt.Sprintf("Bucket %s does not have MFA Delete enabled", bucketName),
		scanner.StatusFail,
		scanner.SeverityHigh,
	)}
}

func (s *Scanner) checkLifecyclePolicy(ctx context.Context, bucketName string) []scanner.Finding {
	lifecycle, err := s.client.GetBucketLifecycleConfiguration(ctx, &s3.GetBucketLifecycleConfigurationInput{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		return []scanner.Finding{s.createFinding(
			"s3_lifecycle_policy",
			bucketName,
			"S3 bucket has no lifecycle policy",
			fmt.Sprintf("Bucket %s does not have a lifecycle policy configured", bucketName),
			scanner.StatusFail,
			scanner.SeverityLow,
		)}
	}

	if lifecycle != nil && len(lifecycle.Rules) > 0 {
		return []scanner.Finding{s.createFinding(
			"s3_lifecycle_policy",
			bucketName,
			"S3 bucket has lifecycle policy configured",
			fmt.Sprintf("Bucket %s has %d lifecycle rules", bucketName, len(lifecycle.Rules)),
			scanner.StatusPass,
			scanner.SeverityLow,
		)}
	}

	return []scanner.Finding{s.createFinding(
		"s3_lifecycle_policy",
		bucketName,
		"S3 bucket has no lifecycle rules",
		fmt.Sprintf("Bucket %s lifecycle configuration has no rules", bucketName),
		scanner.StatusFail,
		scanner.SeverityLow,
	)}
}

func (s *Scanner) checkSSLOnly(ctx context.Context, bucketName string) []scanner.Finding {
	policy, err := s.client.GetBucketPolicy(ctx, &s3.GetBucketPolicyInput{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		return []scanner.Finding{s.createFinding(
			"s3_ssl_only",
			bucketName,
			"S3 bucket has no policy to enforce SSL",
			fmt.Sprintf("Bucket %s has no bucket policy to enforce HTTPS", bucketName),
			scanner.StatusFail,
			scanner.SeverityHigh,
		)}
	}

	// Parse policy to check for aws:SecureTransport condition
	var policyDoc map[string]interface{}
	if err := json.Unmarshal([]byte(aws.ToString(policy.Policy)), &policyDoc); err != nil {
		return nil
	}

	statements, ok := policyDoc["Statement"].([]interface{})
	if !ok {
		return nil
	}

	for _, stmt := range statements {
		stmtMap, ok := stmt.(map[string]interface{})
		if !ok {
			continue
		}
		if effect, ok := stmtMap["Effect"].(string); ok && effect == "Deny" {
			if condition, ok := stmtMap["Condition"].(map[string]interface{}); ok {
				if boolCond, ok := condition["Bool"].(map[string]interface{}); ok {
					if secureTransport, ok := boolCond["aws:SecureTransport"].(string); ok && secureTransport == "false" {
						return []scanner.Finding{s.createFinding(
							"s3_ssl_only",
							bucketName,
							"S3 bucket enforces SSL/HTTPS connections",
							fmt.Sprintf("Bucket %s policy denies non-HTTPS requests", bucketName),
							scanner.StatusPass,
							scanner.SeverityHigh,
						)}
					}
				}
			}
		}
	}

	return []scanner.Finding{s.createFinding(
		"s3_ssl_only",
		bucketName,
		"S3 bucket does not enforce SSL/HTTPS connections",
		fmt.Sprintf("Bucket %s policy does not deny non-HTTPS requests", bucketName),
		scanner.StatusFail,
		scanner.SeverityHigh,
	)}
}

func (s *Scanner) checkObjectLock(ctx context.Context, bucketName string) []scanner.Finding {
	objectLock, err := s.client.GetObjectLockConfiguration(ctx, &s3.GetObjectLockConfigurationInput{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		return []scanner.Finding{s.createFinding(
			"s3_object_lock",
			bucketName,
			"S3 bucket Object Lock is not configured",
			fmt.Sprintf("Bucket %s does not have Object Lock enabled", bucketName),
			scanner.StatusFail,
			scanner.SeverityMedium,
		)}
	}

	if objectLock.ObjectLockConfiguration != nil && objectLock.ObjectLockConfiguration.ObjectLockEnabled == "Enabled" {
		return []scanner.Finding{s.createFinding(
			"s3_object_lock",
			bucketName,
			"S3 bucket Object Lock is enabled",
			fmt.Sprintf("Bucket %s has Object Lock enabled", bucketName),
			scanner.StatusPass,
			scanner.SeverityMedium,
		)}
	}

	return []scanner.Finding{s.createFinding(
		"s3_object_lock",
		bucketName,
		"S3 bucket Object Lock is not enabled",
		fmt.Sprintf("Bucket %s Object Lock is disabled", bucketName),
		scanner.StatusFail,
		scanner.SeverityMedium,
	)}
}
