// Package iam provides IAM security scanning capabilities.
package iam

import (
	"context"
	"fmt"
	"time"

	"cloudcop/api/internal/scanner"
	"cloudcop/api/internal/scanner/compliance"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/iam/types"
)

// Scanner performs security checks on IAM resources.
type Scanner struct {
	client    *iam.Client
	region    string
	accountID string
}

// NewScanner creates a new IAM scanner for the given region and account ID.
func NewScanner(cfg aws.Config, region, accountID string) scanner.ServiceScanner {
	return &Scanner{
		client:    iam.NewFromConfig(cfg),
		region:    region,
		accountID: accountID,
	}
}

// Service returns the AWS service name.
func (i *Scanner) Service() string {
	return "iam"
}

// Scan executes all IAM security checks.
func (i *Scanner) Scan(ctx context.Context, _ string) ([]scanner.Finding, error) {
	var findings []scanner.Finding

	users, err := i.listUsers(ctx)
	if err != nil {
		return nil, fmt.Errorf("listing users: %w", err)
	}

	for _, user := range users {
		findings = append(findings, i.checkUnusedAccessKeys(ctx, user)...)
		findings = append(findings, i.checkAccessKeyRotation(ctx, user)...)
		findings = append(findings, i.checkUserMFA(ctx, user)...)
		findings = append(findings, i.checkInlinePolicies(ctx, user)...)
		findings = append(findings, i.checkConsoleWithoutMFA(ctx, user)...)
	}

	findings = append(findings, i.checkRootMFA(ctx)...)
	findings = append(findings, i.checkPasswordPolicy(ctx)...)
	findings = append(findings, i.checkOverlyPermissivePolicies(ctx)...)
	findings = append(findings, i.checkCrossAccountTrust(ctx)...)

	return findings, nil
}

func (i *Scanner) listUsers(ctx context.Context) ([]types.User, error) {
	var users []types.User
	paginator := iam.NewListUsersPaginator(i.client, &iam.ListUsersInput{})

	for paginator.HasMorePages() {
		output, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		users = append(users, output.Users...)
	}
	return users, nil
}

func (i *Scanner) createFinding(checkID, resourceID, title, description string, status scanner.FindingStatus, severity scanner.Severity) scanner.Finding {
	return scanner.Finding{
		Service:     i.Service(),
		Region:      "global",
		ResourceID:  resourceID,
		CheckID:     checkID,
		Status:      status,
		Severity:    severity,
		Title:       title,
		Description: description,
		Compliance:  compliance.GetCompliance(checkID),
		Timestamp:   time.Now(),
	}
}
