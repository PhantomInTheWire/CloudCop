// Package ecs provides ECS security scanning capabilities.
package ecs

import (
	"context"
	"fmt"
	"time"

	"cloudcop/api/internal/scanner"
	"cloudcop/api/internal/scanner/compliance"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
)

// Scanner performs security checks on ECS resources.
type Scanner struct {
	client    *ecs.Client
	region    string
	accountID string
}

// NewScanner creates and returns a Scanner that implements scanner.ServiceScanner for ECS security scanning.
// cfg is the AWS SDK configuration used to initialize the ECS client; region and accountID are stored as scanner metadata.
func NewScanner(cfg aws.Config, region, accountID string) scanner.ServiceScanner {
	return &Scanner{
		client:    ecs.NewFromConfig(cfg),
		region:    region,
		accountID: accountID,
	}
}

// Service returns the AWS service name.
func (e *Scanner) Service() string {
	return "ecs"
}

// Scan executes all ECS security checks.
func (e *Scanner) Scan(ctx context.Context, _ string) ([]scanner.Finding, error) {
	var findings []scanner.Finding

	taskDefs, err := e.listTaskDefinitions(ctx)
	if err != nil {
		return nil, fmt.Errorf("listing task definitions: %w", err)
	}

	for _, taskDefArn := range taskDefs {
		taskDef, err := e.describeTaskDefinition(ctx, taskDefArn)
		if err != nil {
			continue
		}
		findings = append(findings, e.checkPrivilegedContainers(ctx, taskDef)...)
		findings = append(findings, e.checkPublicRegistry(ctx, taskDef)...)
		findings = append(findings, e.checkTaskIAMRole(ctx, taskDef)...)
		findings = append(findings, e.checkNetworkMode(ctx, taskDef)...)
		findings = append(findings, e.checkSecretsInEnv(ctx, taskDef)...)
		findings = append(findings, e.checkCloudWatchLogs(ctx, taskDef)...)
	}

	return findings, nil
}

func (e *Scanner) listTaskDefinitions(ctx context.Context) ([]string, error) {
	var taskDefs []string
	paginator := ecs.NewListTaskDefinitionsPaginator(e.client, &ecs.ListTaskDefinitionsInput{})

	for paginator.HasMorePages() {
		output, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		taskDefs = append(taskDefs, output.TaskDefinitionArns...)
	}
	return taskDefs, nil
}

func (e *Scanner) describeTaskDefinition(ctx context.Context, arn string) (*types.TaskDefinition, error) {
	output, err := e.client.DescribeTaskDefinition(ctx, &ecs.DescribeTaskDefinitionInput{
		TaskDefinition: aws.String(arn),
	})
	if err != nil {
		return nil, err
	}
	return output.TaskDefinition, nil
}

func (e *Scanner) createFinding(checkID, resourceID, title, description string, status scanner.FindingStatus, severity scanner.Severity) scanner.Finding {
	return scanner.Finding{
		Service:     e.Service(),
		Region:      e.region,
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