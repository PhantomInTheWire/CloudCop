package iam

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"

	"cloudcop/api/internal/scanner"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/iam/types"
)

const accessKeyMaxAgeDays = 90

func (i *Scanner) checkUnusedAccessKeys(ctx context.Context, user types.User) []scanner.Finding {
	var findings []scanner.Finding
	userName := aws.ToString(user.UserName)

	keys, err := i.client.ListAccessKeys(ctx, &iam.ListAccessKeysInput{UserName: user.UserName})
	if err != nil {
		return nil
	}

	for _, key := range keys.AccessKeyMetadata {
		keyID := aws.ToString(key.AccessKeyId)
		lastUsed, err := i.client.GetAccessKeyLastUsed(ctx, &iam.GetAccessKeyLastUsedInput{AccessKeyId: key.AccessKeyId})
		if err != nil {
			continue
		}

		if lastUsed.AccessKeyLastUsed.LastUsedDate == nil {
			findings = append(findings, i.createFinding(
				"iam_unused_access_keys",
				keyID,
				"IAM access key has never been used",
				fmt.Sprintf("Access key %s for user %s has never been used", keyID, userName),
				scanner.StatusFail,
				scanner.SeverityMedium,
			))
			continue
		}

		daysSinceUse := int(time.Since(*lastUsed.AccessKeyLastUsed.LastUsedDate).Hours() / 24)
		if daysSinceUse > accessKeyMaxAgeDays {
			findings = append(findings, i.createFinding(
				"iam_unused_access_keys",
				keyID,
				"IAM access key unused for over 90 days",
				fmt.Sprintf("Access key %s for user %s unused for %d days", keyID, userName, daysSinceUse),
				scanner.StatusFail,
				scanner.SeverityMedium,
			))
		}
	}
	return findings
}

func (i *Scanner) checkAccessKeyRotation(ctx context.Context, user types.User) []scanner.Finding {
	var findings []scanner.Finding
	userName := aws.ToString(user.UserName)

	keys, err := i.client.ListAccessKeys(ctx, &iam.ListAccessKeysInput{UserName: user.UserName})
	if err != nil {
		return nil
	}

	for _, key := range keys.AccessKeyMetadata {
		keyID := aws.ToString(key.AccessKeyId)
		if key.CreateDate == nil {
			continue
		}
		daysSinceCreation := int(time.Since(*key.CreateDate).Hours() / 24)
		if daysSinceCreation > accessKeyMaxAgeDays {
			findings = append(findings, i.createFinding(
				"iam_access_key_rotation",
				keyID,
				"IAM access key not rotated in over 90 days",
				fmt.Sprintf("Access key %s for user %s is %d days old", keyID, userName, daysSinceCreation),
				scanner.StatusFail,
				scanner.SeverityMedium,
			))
		}
	}
	return findings
}

func (i *Scanner) checkUserMFA(ctx context.Context, user types.User) []scanner.Finding {
	userName := aws.ToString(user.UserName)
	mfaDevices, err := i.client.ListMFADevices(ctx, &iam.ListMFADevicesInput{UserName: user.UserName})
	if err != nil {
		return nil
	}

	if len(mfaDevices.MFADevices) == 0 {
		return []scanner.Finding{i.createFinding(
			"iam_user_mfa",
			userName,
			"IAM user does not have MFA enabled",
			fmt.Sprintf("User %s has no MFA device configured", userName),
			scanner.StatusFail,
			scanner.SeverityHigh,
		)}
	}
	return []scanner.Finding{i.createFinding(
		"iam_user_mfa",
		userName,
		"IAM user has MFA enabled",
		fmt.Sprintf("User %s has MFA configured", userName),
		scanner.StatusPass,
		scanner.SeverityHigh,
	)}
}

func (i *Scanner) checkRootMFA(ctx context.Context) []scanner.Finding {
	summary, err := i.client.GetAccountSummary(ctx, &iam.GetAccountSummaryInput{})
	if err != nil {
		return nil
	}

	// Check if key exists and get value
	mfaEnabled, exists := summary.SummaryMap["AccountMFAEnabled"]
	if !exists {
		// Key missing - treat as pass/skip since we can't determine
		return []scanner.Finding{i.createFinding(
			"iam_root_mfa",
			"root",
			"Root account MFA status unknown",
			"Could not determine root account MFA status from account summary",
			scanner.StatusPass,
			scanner.SeverityCritical,
		)}
	}

	// Explicitly check for zero value
	if mfaEnabled == 0 {
		return []scanner.Finding{i.createFinding(
			"iam_root_mfa",
			"root",
			"Root account MFA is not enabled",
			"The AWS root account does not have MFA enabled",
			scanner.StatusFail,
			scanner.SeverityCritical,
		)}
	}
	return []scanner.Finding{i.createFinding(
		"iam_root_mfa",
		"root",
		"Root account MFA is enabled",
		"The AWS root account has MFA enabled",
		scanner.StatusPass,
		scanner.SeverityCritical,
	)}
}

func (i *Scanner) checkPasswordPolicy(ctx context.Context) []scanner.Finding {
	policy, err := i.client.GetAccountPasswordPolicy(ctx, &iam.GetAccountPasswordPolicyInput{})
	if err != nil {
		return []scanner.Finding{i.createFinding(
			"iam_password_policy",
			"account",
			"No password policy configured",
			"The AWS account has no custom password policy",
			scanner.StatusFail,
			scanner.SeverityMedium,
		)}
	}

	pp := policy.PasswordPolicy
	issues := []string{}
	if pp.MinimumPasswordLength != nil && *pp.MinimumPasswordLength < 14 {
		issues = append(issues, "minimum length < 14")
	}
	if !pp.RequireUppercaseCharacters {
		issues = append(issues, "uppercase not required")
	}
	if !pp.RequireLowercaseCharacters {
		issues = append(issues, "lowercase not required")
	}
	if !pp.RequireNumbers {
		issues = append(issues, "numbers not required")
	}
	if !pp.RequireSymbols {
		issues = append(issues, "symbols not required")
	}

	if len(issues) > 0 {
		return []scanner.Finding{i.createFinding(
			"iam_password_policy",
			"account",
			"Password policy does not meet best practices",
			fmt.Sprintf("Password policy issues: %v", issues),
			scanner.StatusFail,
			scanner.SeverityMedium,
		)}
	}
	return []scanner.Finding{i.createFinding(
		"iam_password_policy",
		"account",
		"Password policy meets best practices",
		"The password policy meets security requirements",
		scanner.StatusPass,
		scanner.SeverityMedium,
	)}
}

func (i *Scanner) checkInlinePolicies(ctx context.Context, user types.User) []scanner.Finding {
	userName := aws.ToString(user.UserName)
	policies, err := i.client.ListUserPolicies(ctx, &iam.ListUserPoliciesInput{UserName: user.UserName})
	if err != nil {
		return nil
	}

	if len(policies.PolicyNames) > 0 {
		return []scanner.Finding{i.createFinding(
			"iam_inline_policies",
			userName,
			"IAM user has inline policies",
			fmt.Sprintf("User %s has %d inline policies (use managed policies instead)", userName, len(policies.PolicyNames)),
			scanner.StatusFail,
			scanner.SeverityLow,
		)}
	}
	return nil
}

func (i *Scanner) checkOverlyPermissivePolicies(ctx context.Context) []scanner.Finding {
	var findings []scanner.Finding
	paginator := iam.NewListPoliciesPaginator(i.client, &iam.ListPoliciesInput{Scope: types.PolicyScopeTypeLocal})

	for paginator.HasMorePages() {
		output, err := paginator.NextPage(ctx)
		if err != nil {
			break
		}
		for _, policy := range output.Policies {
			policyArn := aws.ToString(policy.Arn)
			version, err := i.client.GetPolicyVersion(ctx, &iam.GetPolicyVersionInput{
				PolicyArn: policy.Arn,
				VersionId: policy.DefaultVersionId,
			})
			if err != nil {
				continue
			}
			doc, err := url.QueryUnescape(aws.ToString(version.PolicyVersion.Document))
			if err != nil {
				continue
			}
			var policyDoc struct {
				Statement []struct {
					Effect   string      `json:"Effect"`
					Action   interface{} `json:"Action"`
					Resource interface{} `json:"Resource"`
				} `json:"Statement"`
			}
			if err := json.Unmarshal([]byte(doc), &policyDoc); err != nil {
				continue
			}
			for _, stmt := range policyDoc.Statement {
				if stmt.Effect == "Allow" && isWildcard(stmt.Action) && isWildcard(stmt.Resource) {
					findings = append(findings, i.createFinding(
						"iam_overly_permissive",
						policyArn,
						"IAM policy is overly permissive",
						fmt.Sprintf("Policy %s allows Action:* on Resource:*", aws.ToString(policy.PolicyName)),
						scanner.StatusFail,
						scanner.SeverityCritical,
					))
				}
			}
		}
	}
	return findings
}

func (i *Scanner) checkCrossAccountTrust(ctx context.Context) []scanner.Finding {
	var findings []scanner.Finding
	paginator := iam.NewListRolesPaginator(i.client, &iam.ListRolesInput{})

	for paginator.HasMorePages() {
		output, err := paginator.NextPage(ctx)
		if err != nil {
			break
		}
		for _, role := range output.Roles {
			roleName := aws.ToString(role.RoleName)
			doc, err := url.QueryUnescape(aws.ToString(role.AssumeRolePolicyDocument))
			if err != nil {
				continue
			}
			var trustPolicy struct {
				Statement []struct {
					Principal interface{} `json:"Principal"`
				} `json:"Statement"`
			}
			if err := json.Unmarshal([]byte(doc), &trustPolicy); err != nil {
				continue
			}
			for _, stmt := range trustPolicy.Statement {
				if hasCrossAccountPrincipal(stmt.Principal, i.accountID) {
					findings = append(findings, i.createFinding(
						"iam_cross_account_trust",
						roleName,
						"IAM role has cross-account trust",
						fmt.Sprintf("Role %s trusts external AWS accounts", roleName),
						scanner.StatusFail,
						scanner.SeverityHigh,
					))
				}
			}
		}
	}
	return findings
}

func (i *Scanner) checkConsoleWithoutMFA(ctx context.Context, user types.User) []scanner.Finding {
	userName := aws.ToString(user.UserName)

	_, err := i.client.GetLoginProfile(ctx, &iam.GetLoginProfileInput{UserName: user.UserName})
	if err != nil {
		return nil // User has no console access
	}

	mfaDevices, err := i.client.ListMFADevices(ctx, &iam.ListMFADevicesInput{UserName: user.UserName})
	if err != nil {
		return nil
	}

	if len(mfaDevices.MFADevices) == 0 {
		return []scanner.Finding{i.createFinding(
			"iam_console_without_mfa",
			userName,
			"IAM user has console access without MFA",
			fmt.Sprintf("User %s can access console without MFA", userName),
			scanner.StatusFail,
			scanner.SeverityHigh,
		)}
	}
	return nil
}

// isWildcard reports whether v represents a wildcard ("*").
// It returns true if v is the string "*" or a slice containing the string "*" as any element; otherwise it returns false.
func isWildcard(v interface{}) bool {
	switch val := v.(type) {
	case string:
		return val == "*"
	case []interface{}:
		for _, item := range val {
			if s, ok := item.(string); ok && s == "*" {
				return true
			}
		}
	}
	return false
}

// hasCrossAccountPrincipal reports whether the given principal represents cross-account access
// relative to the provided account ID.
// It returns true if the principal is a wildcard (`"*"`) or contains an `AWS` principal value
// that does not include the provided account ID, false otherwise.
func hasCrossAccountPrincipal(principal interface{}, accountID string) bool {
	switch p := principal.(type) {
	case string:
		return p == "*"
	case map[string]interface{}:
		if aws, ok := p["AWS"]; ok {
			switch v := aws.(type) {
			case string:
				return !containsAccountID(v, accountID)
			case []interface{}:
				for _, item := range v {
					if s, ok := item.(string); ok && !containsAccountID(s, accountID) {
						return true
					}
				}
			}
		}
	}
	return false
}

// containsAccountID reports whether arn contains the provided accountID.
// It returns true only when both arn and accountID are non-empty and either
// arn equals accountID or arn contains accountID according to the contains helper.
func containsAccountID(arn, accountID string) bool {
	return len(accountID) > 0 && len(arn) > 0 && (arn == accountID || strings.Contains(arn, accountID))
}
