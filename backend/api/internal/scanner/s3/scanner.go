// Package s3 provides S3 security scanning capabilities.
package s3

import (
	"context"
	"fmt"
	"time"

	"cloudcop/api/internal/scanner"
	"cloudcop/api/internal/scanner/compliance"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// Scanner performs security checks on S3 buckets.
type Scanner struct {
	client    *s3.Client
	region    string
	accountID string
}

// NewScanner creates a new S3 scanner.
func NewScanner(cfg aws.Config, region, accountID string) scanner.ServiceScanner {
	return &Scanner{
		client:    s3.NewFromConfig(cfg),
		region:    region,
		accountID: accountID,
	}
}

// Service returns the AWS service name.
func (s *Scanner) Service() string {
	return "s3"
}

// Scan executes all S3 security checks.
func (s *Scanner) Scan(ctx context.Context, _ string) ([]scanner.Finding, error) {
	buckets, err := s.listBucketsInRegion(ctx)
	if err != nil {
		return nil, fmt.Errorf("listing buckets: %w", err)
	}

	var findings []scanner.Finding

	for _, bucket := range buckets {
		bucketName := aws.ToString(bucket.Name)

		// Execute all S3 checks
		findings = append(findings, s.checkPublicAccess(ctx, bucketName)...)
		findings = append(findings, s.checkBucketPolicy(ctx, bucketName)...)
		findings = append(findings, s.checkEncryption(ctx, bucketName)...)
		findings = append(findings, s.checkVersioning(ctx, bucketName)...)
		findings = append(findings, s.checkLogging(ctx, bucketName)...)
		findings = append(findings, s.checkBlockPublicAccess(ctx, bucketName)...)
		findings = append(findings, s.checkMFADelete(ctx, bucketName)...)
		findings = append(findings, s.checkLifecyclePolicy(ctx, bucketName)...)
		findings = append(findings, s.checkSSLOnly(ctx, bucketName)...)
		findings = append(findings, s.checkObjectLock(ctx, bucketName)...)
	}

	return findings, nil
}

func (s *Scanner) listBucketsInRegion(ctx context.Context) ([]types.Bucket, error) {
	result, err := s.client.ListBuckets(ctx, &s3.ListBucketsInput{})
	if err != nil {
		return nil, err
	}

	var bucketsInRegion []types.Bucket
	for _, bucket := range result.Buckets {
		bucketName := aws.ToString(bucket.Name)
		location, err := s.client.GetBucketLocation(ctx, &s3.GetBucketLocationInput{
			Bucket: aws.String(bucketName),
		})
		if err != nil {
			continue // Skip buckets we can't access
		}

		// AWS returns empty string for us-east-1
		bucketRegion := string(location.LocationConstraint)
		if bucketRegion == "" {
			bucketRegion = "us-east-1"
		}

		if bucketRegion == s.region {
			bucketsInRegion = append(bucketsInRegion, bucket)
		}
	}

	return bucketsInRegion, nil
}

func (s *Scanner) createFinding(checkID, resourceID, title, description string, status scanner.FindingStatus, severity scanner.Severity) scanner.Finding {
	return scanner.Finding{
		Service:     s.Service(),
		Region:      s.region,
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
