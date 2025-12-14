package ecs

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
	
	if got := s.Service(); got != "ecs" {
		t.Errorf("Service() = %v, want ecs", got)
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
			name:        "Privileged container critical",
			checkID:     "ecs_privileged_container",
			resourceID:  "arn:aws:ecs:us-east-1:123456789012:task-definition/my-task:1",
			title:       "ECS task has privileged container",
			description: "Container runs in privileged mode",
			status:      scanner.StatusFail,
			severity:    scanner.SeverityCritical,
		},
		{
			name:        "Public registry medium",
			checkID:     "ecs_public_registry",
			resourceID:  "arn:aws:ecs:us-east-1:123456789012:task-definition/my-task:2",
			title:       "Container uses public registry",
			description: "Image from docker.io",
			status:      scanner.StatusFail,
			severity:    scanner.SeverityMedium,
		},
		{
			name:        "Task IAM role pass",
			checkID:     "ecs_task_iam_role",
			resourceID:  "arn:aws:ecs:us-east-1:123456789012:task-definition/secure-task:1",
			title:       "ECS task has IAM role",
			description: "Task has IAM role assigned",
			status:      scanner.StatusPass,
			severity:    scanner.SeverityMedium,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			before := time.Now()
			finding := s.createFinding(tt.checkID, tt.resourceID, tt.title, tt.description, tt.status, tt.severity)
			after := time.Now()
			
			if finding.Service != "ecs" {
				t.Errorf("Service = %v, want ecs", finding.Service)
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

func TestSensitiveEnvPatterns(t *testing.T) {
	// Test that the exported variable exists and has expected patterns
	expectedPatterns := []string{
		"SECRET", "PASSWORD", "KEY", "TOKEN", "CREDENTIAL", "API_KEY", "PRIVATE", "AUTH",
	}
	
	if len(sensitiveEnvPatterns) != len(expectedPatterns) {
		t.Errorf("sensitiveEnvPatterns has %d items, want %d", len(sensitiveEnvPatterns), len(expectedPatterns))
	}
	
	for _, expected := range expectedPatterns {
		found := false
		for _, pattern := range sensitiveEnvPatterns {
			if pattern == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected pattern %s not found in sensitiveEnvPatterns", expected)
		}
	}
}