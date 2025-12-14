package lambda

import (
	"testing"
	"time"

	"cloudcop/api/internal/scanner"

	"github.com/aws/aws-sdk-go-v2/aws"
)

const testServiceName = "lambda"

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

	if got := s.Service(); got != testServiceName {
		t.Errorf("Service() = %v, want %s", got, testServiceName)
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
			name:        "Environment secrets fail",
			checkID:     "lambda_env_secrets",
			resourceID:  "my-function",
			title:       "Lambda has secrets in env vars",
			description: "Function has API_KEY in environment",
			status:      scanner.StatusFail,
			severity:    scanner.SeverityHigh,
		},
		{
			name:        "VPC config pass",
			checkID:     "lambda_vpc_config",
			resourceID:  "secure-function",
			title:       "Lambda in VPC",
			description: "Function is configured in VPC",
			status:      scanner.StatusPass,
			severity:    scanner.SeverityMedium,
		},
		{
			name:        "CloudWatch logs fail",
			checkID:     "lambda_cloudwatch_logs",
			resourceID:  "unmonitored-function",
			title:       "Lambda has no CloudWatch logs",
			description: "Function does not log to CloudWatch",
			status:      scanner.StatusFail,
			severity:    scanner.SeverityMedium,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			before := time.Now()
			finding := s.createFinding(tt.checkID, tt.resourceID, tt.title, tt.description, tt.status, tt.severity)
			after := time.Now()

			if finding.Service != testServiceName {
				t.Errorf("Service = %v, want %s", finding.Service, testServiceName)
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

func TestSensitiveEnvVarPatterns(t *testing.T) {
	// Test that the exported variable exists and has expected patterns
	expectedPatterns := []string{
		"SECRET", "PASSWORD", "KEY", "TOKEN", "CREDENTIAL", "API_KEY", "PRIVATE", "AUTH",
	}

	// Just verify that expected patterns are present (actual array may have more)
	if len(sensitiveEnvVarPatterns) == 0 {
		t.Error("sensitiveEnvVarPatterns should not be empty")
	}

	for _, expected := range expectedPatterns {
		found := false
		for _, pattern := range sensitiveEnvVarPatterns {
			if pattern == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected pattern %s not found in sensitiveEnvVarPatterns", expected)
		}
	}
}
