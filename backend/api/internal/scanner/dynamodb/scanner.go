// Package dynamodb provides DynamoDB security scanning capabilities.
package dynamodb

import (
	"context"
	"fmt"
	"time"

	"cloudcop/api/internal/scanner"
	"cloudcop/api/internal/scanner/compliance"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

// Scanner performs security checks on DynamoDB tables.
type Scanner struct {
	client    *dynamodb.Client
	region    string
	accountID string
}

// NewScanner constructs a DynamoDB scanner configured for the given AWS region and account.
// It initializes an AWS DynamoDB client from cfg and returns the scanner as a scanner.ServiceScanner.
func NewScanner(cfg aws.Config, region, accountID string) scanner.ServiceScanner {
	return &Scanner{
		client:    dynamodb.NewFromConfig(cfg),
		region:    region,
		accountID: accountID,
	}
}

// Service returns the AWS service name.
func (d *Scanner) Service() string {
	return "dynamodb"
}

// Scan executes all DynamoDB security checks.
func (d *Scanner) Scan(ctx context.Context, _ string) ([]scanner.Finding, error) {
	var findings []scanner.Finding

	tables, err := d.listTables(ctx)
	if err != nil {
		return nil, fmt.Errorf("listing tables: %w", err)
	}

	for _, tableName := range tables {
		findings = append(findings, d.checkEncryption(ctx, tableName)...)
		findings = append(findings, d.checkPITR(ctx, tableName)...)
		findings = append(findings, d.checkTTL(ctx, tableName)...)
		findings = append(findings, d.checkAutoScaling(ctx, tableName)...)
	}

	return findings, nil
}

func (d *Scanner) listTables(ctx context.Context) ([]string, error) {
	var tables []string
	paginator := dynamodb.NewListTablesPaginator(d.client, &dynamodb.ListTablesInput{})

	for paginator.HasMorePages() {
		output, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		tables = append(tables, output.TableNames...)
	}
	return tables, nil
}

func (d *Scanner) createFinding(checkID, resourceID, title, description string, status scanner.FindingStatus, severity scanner.Severity) scanner.Finding {
	return scanner.Finding{
		Service:     d.Service(),
		Region:      d.region,
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