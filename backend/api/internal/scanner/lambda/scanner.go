// Package lambda provides Lambda security scanning capabilities.
package lambda

import (
	"context"
	"fmt"
	"time"

	"cloudcop/api/internal/scanner"
	"cloudcop/api/internal/scanner/compliance"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
)

// Scanner performs security checks on Lambda functions.
type Scanner struct {
	client    *lambda.Client
	region    string
	accountID string
}

// NewScanner creates a new Lambda scanner.
func NewScanner(cfg aws.Config, region, accountID string) scanner.ServiceScanner {
	return &Scanner{
		client:    lambda.NewFromConfig(cfg),
		region:    region,
		accountID: accountID,
	}
}

// Service returns the AWS service name.
func (l *Scanner) Service() string {
	return "lambda"
}

// Scan executes all Lambda security checks.
func (l *Scanner) Scan(ctx context.Context, _ string) ([]scanner.Finding, error) {
	var findings []scanner.Finding

	functions, err := l.listFunctions(ctx)
	if err != nil {
		return nil, fmt.Errorf("listing functions: %w", err)
	}

	for _, fn := range functions {
		findings = append(findings, l.checkEnvSecrets(ctx, fn)...)
		findings = append(findings, l.checkCloudWatchLogs(ctx, fn)...)
		findings = append(findings, l.checkVPCConfig(ctx, fn)...)
		findings = append(findings, l.checkDLQ(ctx, fn)...)
		findings = append(findings, l.checkTracing(ctx, fn)...)
		findings = append(findings, l.checkTimeout(ctx, fn)...)
		findings = append(findings, l.checkReservedConcurrency(ctx, fn)...)
	}

	return findings, nil
}

func (l *Scanner) listFunctions(ctx context.Context) ([]types.FunctionConfiguration, error) {
	var functions []types.FunctionConfiguration
	paginator := lambda.NewListFunctionsPaginator(l.client, &lambda.ListFunctionsInput{})

	for paginator.HasMorePages() {
		output, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		functions = append(functions, output.Functions...)
	}
	return functions, nil
}

func (l *Scanner) createFinding(checkID, resourceID, title, description string, status scanner.FindingStatus, severity scanner.Severity) scanner.Finding {
	return scanner.Finding{
		Service:     l.Service(),
		Region:      l.region,
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
