// Package dynamodb provides DynamoDB security scanning capabilities.
package dynamodb

import (
	"context"
	"fmt"

	"cloudcop/api/internal/scanner"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

func (d *Scanner) checkEncryption(ctx context.Context, tableName string) []scanner.Finding {
	table, err := d.client.DescribeTable(ctx, &dynamodb.DescribeTableInput{
		TableName: aws.String(tableName),
	})
	if err != nil {
		return nil
	}

	if table.Table.SSEDescription != nil && table.Table.SSEDescription.Status == types.SSEStatusEnabled {
		return []scanner.Finding{d.createFinding(
			"dynamodb_encryption",
			tableName,
			"DynamoDB table has encryption enabled",
			fmt.Sprintf("Table %s has server-side encryption enabled", tableName),
			scanner.StatusPass,
			scanner.SeverityHigh,
		)}
	}
	return []scanner.Finding{d.createFinding(
		"dynamodb_encryption",
		tableName,
		"DynamoDB table does not have encryption enabled",
		fmt.Sprintf("Table %s does not have server-side encryption enabled", tableName),
		scanner.StatusFail,
		scanner.SeverityHigh,
	)}
}

func (d *Scanner) checkPITR(ctx context.Context, tableName string) []scanner.Finding {
	pitr, err := d.client.DescribeContinuousBackups(ctx, &dynamodb.DescribeContinuousBackupsInput{
		TableName: aws.String(tableName),
	})
	if err != nil {
		return nil
	}

	if pitr.ContinuousBackupsDescription != nil &&
		pitr.ContinuousBackupsDescription.PointInTimeRecoveryDescription != nil &&
		pitr.ContinuousBackupsDescription.PointInTimeRecoveryDescription.PointInTimeRecoveryStatus == types.PointInTimeRecoveryStatusEnabled {
		return []scanner.Finding{d.createFinding(
			"dynamodb_pitr",
			tableName,
			"DynamoDB table has point-in-time recovery enabled",
			fmt.Sprintf("Table %s has PITR enabled", tableName),
			scanner.StatusPass,
			scanner.SeverityMedium,
		)}
	}
	return []scanner.Finding{d.createFinding(
		"dynamodb_pitr",
		tableName,
		"DynamoDB table does not have point-in-time recovery",
		fmt.Sprintf("Table %s does not have PITR enabled", tableName),
		scanner.StatusFail,
		scanner.SeverityMedium,
	)}
}

func (d *Scanner) checkTTL(ctx context.Context, tableName string) []scanner.Finding {
	ttl, err := d.client.DescribeTimeToLive(ctx, &dynamodb.DescribeTimeToLiveInput{
		TableName: aws.String(tableName),
	})
	if err != nil {
		return nil
	}

	if ttl.TimeToLiveDescription != nil && ttl.TimeToLiveDescription.TimeToLiveStatus == types.TimeToLiveStatusEnabled {
		return []scanner.Finding{d.createFinding(
			"dynamodb_ttl",
			tableName,
			"DynamoDB table has TTL configured",
			fmt.Sprintf("Table %s has TTL enabled on attribute %s", tableName, aws.ToString(ttl.TimeToLiveDescription.AttributeName)),
			scanner.StatusPass,
			scanner.SeverityLow,
		)}
	}
	return []scanner.Finding{d.createFinding(
		"dynamodb_ttl",
		tableName,
		"DynamoDB table has no TTL configured",
		fmt.Sprintf("Table %s does not have TTL configured", tableName),
		scanner.StatusFail,
		scanner.SeverityLow,
	)}
}

func (d *Scanner) checkAutoScaling(ctx context.Context, tableName string) []scanner.Finding {
	table, err := d.client.DescribeTable(ctx, &dynamodb.DescribeTableInput{
		TableName: aws.String(tableName),
	})
	if err != nil {
		return nil
	}

	if table.Table.BillingModeSummary != nil && table.Table.BillingModeSummary.BillingMode == types.BillingModePayPerRequest {
		return []scanner.Finding{d.createFinding(
			"dynamodb_auto_scaling",
			tableName,
			"DynamoDB table uses on-demand capacity",
			fmt.Sprintf("Table %s uses PAY_PER_REQUEST (auto-scales automatically)", tableName),
			scanner.StatusPass,
			scanner.SeverityLow,
		)}
	}

	// For provisioned capacity, we'd need to check Application Auto Scaling
	// For simplicity, we flag provisioned tables as needing review
	return []scanner.Finding{d.createFinding(
		"dynamodb_auto_scaling",
		tableName,
		"DynamoDB table uses provisioned capacity",
		fmt.Sprintf("Table %s uses provisioned capacity (verify auto-scaling is configured)", tableName),
		scanner.StatusFail,
		scanner.SeverityLow,
	)}
}
