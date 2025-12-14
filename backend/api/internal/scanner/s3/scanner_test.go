package s3

import (
	"testing"
	"time"

	"cloudcop/api/internal/scanner"

	"github.com/aws/aws-sdk-go-v2/aws"
)

func TestNewScanner(t *testing.T) {
	cfg := aws.Config{Region: "us-east-1"}
	region := "us-east-1"
	accountID := "123456789012"

	s := NewScanner(cfg, region, accountID)

	if s == nil {
		t.Fatal("NewScanner returned nil")
	}

	scanner, ok := s.(*Scanner)
	if !ok {
		t.Fatal("NewScanner did not return *Scanner type")
	}

	if scanner.region != region {
		t.Errorf("region = %v, want %v", scanner.region, region)
	}
	if scanner.accountID != accountID {
		t.Errorf("accountID = %v, want %v", scanner.accountID, accountID)
	}
	if scanner.client == nil {
		t.Error("client not initialized")
	}
}

func TestScanner_Service(t *testing.T) {
	s := &Scanner{}

	if got := s.Service(); got != "s3" {
		t.Errorf("Service() = %v, want s3", got)
	}
}

func TestScanner_createFinding(t *testing.T) {
	s := &Scanner{
		region:    "us-east-1",
		accountID: "123456789012",
	}

	tests := []struct {
		name        string
		checkID     string
		resourceID  string
		title       string
		description string
		status      scanner.FindingStatus
		severity    scanner.Severity
	}{
		{
			name:        "Public access fail",
			checkID:     "s3_bucket_public_access",
			resourceID:  "my-bucket",
			title:       "S3 bucket is publicly accessible",
			description: "Bucket allows public access",
			status:      scanner.StatusFail,
			severity:    scanner.SeverityCritical,
		},
		{
			name:        "Encryption pass",
			checkID:     "s3_bucket_encryption",
			resourceID:  "secure-bucket",
			title:       "S3 bucket has encryption enabled",
			description: "Bucket uses AES256 encryption",
			status:      scanner.StatusPass,
			severity:    scanner.SeverityHigh,
		},
		{
			name:        "Versioning medium",
			checkID:     "s3_bucket_versioning",
			resourceID:  "versioned-bucket",
			title:       "S3 bucket versioning enabled",
			description: "Bucket has versioning configured",
			status:      scanner.StatusPass,
			severity:    scanner.SeverityMedium,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			before := time.Now()
			finding := s.createFinding(tt.checkID, tt.resourceID, tt.title, tt.description, tt.status, tt.severity)
			after := time.Now()

			if finding.Service != "s3" {
				t.Errorf("Service = %v, want s3", finding.Service)
			}
			if finding.Region != s.region {
				t.Errorf("Region = %v, want %v", finding.Region, s.region)
			}
			if finding.ResourceID != tt.resourceID {
				t.Errorf("ResourceID = %v, want %v", finding.ResourceID, tt.resourceID)
			}
			if finding.CheckID != tt.checkID {
				t.Errorf("CheckID = %v, want %v", finding.CheckID, tt.checkID)
			}
			if finding.Status != tt.status {
				t.Errorf("Status = %v, want %v", finding.Status, tt.status)
			}
			if finding.Severity != tt.severity {
				t.Errorf("Severity = %v, want %v", finding.Severity, tt.severity)
			}
			if finding.Timestamp.Before(before) || finding.Timestamp.After(after) {
				t.Error("Timestamp not in expected range")
			}
		})
	}
}
