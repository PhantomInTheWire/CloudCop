package iam

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
	
	if got := s.Service(); got != "iam" {
		t.Errorf("Service() = %v, want iam", got)
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
			name:        "Root MFA fail",
			checkID:     "iam_root_mfa",
			resourceID:  "root",
			title:       "Root account has no MFA",
			description: "Root account does not have MFA enabled",
			status:      scanner.StatusFail,
			severity:    scanner.SeverityCritical,
		},
		{
			name:        "User MFA pass",
			checkID:     "iam_user_mfa",
			resourceID:  "test-user",
			title:       "User has MFA enabled",
			description: "User test-user has MFA configured",
			status:      scanner.StatusPass,
			severity:    scanner.SeverityHigh,
		},
		{
			name:        "Access key rotation fail",
			checkID:     "iam_access_key_rotation",
			resourceID:  "AKIAIOSFODNN7EXAMPLE",
			title:       "Access key not rotated",
			description: "Access key is 120 days old",
			status:      scanner.StatusFail,
			severity:    scanner.SeverityMedium,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			before := time.Now()
			finding := s.createFinding(tt.checkID, tt.resourceID, tt.title, tt.description, tt.status, tt.severity)
			after := time.Now()
			
			if finding.Service != "iam" {
				t.Errorf("Service = %v, want iam", finding.Service)
			}
			if finding.Region != "global" {
				t.Errorf("Region = %v, want global", finding.Region)
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

func TestAccessKeyMaxAgeDays(t *testing.T) {
	if accessKeyMaxAgeDays != 90 {
		t.Errorf("accessKeyMaxAgeDays = %d, want 90", accessKeyMaxAgeDays)
	}
}