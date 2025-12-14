package dynamodb

import (
	"testing"
	"time"

	"cloudcop/api/internal/scanner"

	"github.com/aws/aws-sdk-go-v2/aws"
)

const testServiceName = "dynamodb"

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
	if scanner.client == nil {
		t.Error("client not initialized")
	}
}

func TestScanner_Service(t *testing.T) {
	s := &Scanner{}

	if got := s.Service(); got != testServiceName {
		t.Errorf("Service() = %v, want %s", got, testServiceName)
	}
}

func TestScanner_createFinding(t *testing.T) {
	s := &Scanner{
		region: "us-east-1",
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
			name:        "Pass finding",
			checkID:     "dynamodb_encryption",
			resourceID:  "test-table",
			title:       "Test title",
			description: "Test description",
			status:      scanner.StatusPass,
			severity:    scanner.SeverityHigh,
		},
		{
			name:        "Fail finding",
			checkID:     "dynamodb_pitr",
			resourceID:  "another-table",
			title:       "Another title",
			description: "Another description",
			status:      scanner.StatusFail,
			severity:    scanner.SeverityMedium,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			before := time.Now()
			finding := s.createFinding(tt.checkID, tt.resourceID, tt.title, tt.description, tt.status, tt.severity)
			after := time.Now()

			if finding.Service != "dynamodb" {
				t.Errorf("Service = %v, want dynamodb", finding.Service)
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
			if finding.Title != tt.title {
				t.Errorf("Title = %v, want %v", finding.Title, tt.title)
			}
			if finding.Description != tt.description {
				t.Errorf("Description = %v, want %v", finding.Description, tt.description)
			}
			if finding.Timestamp.Before(before) || finding.Timestamp.After(after) {
				t.Error("Timestamp not in expected range")
			}
			if finding.Compliance == nil {
				t.Error("Compliance should be initialized (may be empty)")
			}
		})
	}
}

func TestScanner_createFinding_ComplianceMappings(t *testing.T) {
	s := &Scanner{region: "us-east-1"}

	// Test that compliance mappings are included
	finding := s.createFinding(
		"dynamodb_encryption",
		"test-table",
		"Test",
		"Test description",
		scanner.StatusFail,
		scanner.SeverityHigh,
	)

	// dynamodb_encryption should have compliance mappings
	if len(finding.Compliance) == 0 {
		t.Error("Expected compliance mappings for dynamodb_encryption check")
	}
}
