// Package ec2 provides EC2 security scanning capabilities.
package ec2

import (
	"context"
	"fmt"
	"time"

	"cloudcop/api/internal/scanner"
	"cloudcop/api/internal/scanner/compliance"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

// Scanner performs security checks on EC2 resources.
type Scanner struct {
	client    *ec2.Client
	region    string
	accountID string
}

// NewScanner creates a new EC2 Scanner configured with the provided AWS config, region, and account ID.
// The returned Scanner uses an EC2 client constructed from cfg and is initialized with region and accountID.
func NewScanner(cfg aws.Config, region, accountID string) scanner.ServiceScanner {
	return &Scanner{
		client:    ec2.NewFromConfig(cfg),
		region:    region,
		accountID: accountID,
	}
}

// Service returns the AWS service name.
func (e *Scanner) Service() string {
	return "ec2"
}

// Scan executes all EC2 security checks.
func (e *Scanner) Scan(ctx context.Context, _ string) ([]scanner.Finding, error) {
	var findings []scanner.Finding

	instances, err := e.listInstances(ctx)
	if err != nil {
		return nil, fmt.Errorf("listing instances: %w", err)
	}

	for _, instance := range instances {
		instanceID := aws.ToString(instance.InstanceId)
		findings = append(findings, e.checkPublicIP(ctx, instance)...)
		findings = append(findings, e.checkEBSEncryption(ctx, instance)...)
		findings = append(findings, e.checkSecurityGroups(ctx, instance)...)
		findings = append(findings, e.checkIMDSv2(ctx, instance)...)
		findings = append(findings, e.checkIAMRole(ctx, instance)...)
		findings = append(findings, e.checkCloudWatchMonitoring(ctx, instance)...)
		findings = append(findings, e.checkDetailedMonitoring(ctx, instance)...)
		_ = instanceID
	}

	findings = append(findings, e.checkUnassociatedElasticIPs(ctx)...)
	findings = append(findings, e.checkUnrestrictedSecurityGroups(ctx)...)
	findings = append(findings, e.checkDangerousPorts(ctx)...)

	return findings, nil
}

func (e *Scanner) listInstances(ctx context.Context) ([]types.Instance, error) {
	var instances []types.Instance
	paginator := ec2.NewDescribeInstancesPaginator(e.client, &ec2.DescribeInstancesInput{})

	for paginator.HasMorePages() {
		output, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		for _, reservation := range output.Reservations {
			instances = append(instances, reservation.Instances...)
		}
	}
	return instances, nil
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
