package ec2

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
	
	if got := s.Service(); got != "ec2" {
		t.Errorf("Service() = %v, want ec2", got)
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
			name:        "IMDSv2 pass",
			checkID:     "ec2_imdsv2_required",
			resourceID:  "i-1234567890abcdef0",
			title:       "EC2 instance requires IMDSv2",
			description: "Instance enforces IMDSv2",
			status:      scanner.StatusPass,
			severity:    scanner.SeverityHigh,
		},
		{
			name:        "Public IP fail",
			checkID:     "ec2_public_ip",
			resourceID:  "i-0987654321fedcba0",
			title:       "EC2 instance has public IP",
			description: "Instance has public IP address",
			status:      scanner.StatusFail,
			severity:    scanner.SeverityMedium,
		},
		{
			name:        "EBS encryption critical",
			checkID:     "ec2_ebs_encryption",
			resourceID:  "vol-1234567890",
			title:       "EBS volume not encrypted",
			description: "Volume is unencrypted",
			status:      scanner.StatusFail,
			severity:    scanner.SeverityCritical,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			before := time.Now()
			finding := s.createFinding(tt.checkID, tt.resourceID, tt.title, tt.description, tt.status, tt.severity)
			after := time.Now()
			
			if finding.Service != "ec2" {
				t.Errorf("Service = %v, want ec2", finding.Service)
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
			if finding.Timestamp.Before(before) || finding.Timestamp.After(after) {
				t.Error("Timestamp not in expected range")
			}
		})
	}
}

func TestDangerousPortsMap(t *testing.T) {
	expectedPorts := map[int32]string{
		22:    "SSH",
		3389:  "RDP",
		3306:  "MySQL",
		5432:  "PostgreSQL",
		1433:  "MSSQL",
		27017: "MongoDB",
		6379:  "Redis",
	}
	
	for port, name := range expectedPorts {
		if dangerousPorts[port] != name {
			t.Errorf("dangerousPorts[%d] = %v, want %v", port, dangerousPorts[port], name)
		}
	}
	
	// Verify map has expected count
	if len(dangerousPorts) != len(expectedPorts) {
		t.Errorf("dangerousPorts has %d entries, want %d", len(dangerousPorts), len(expectedPorts))
	}
}

func TestIPv4AnyConstant(t *testing.T) {
	if ipv4Any != "0.0.0.0/0" {
		t.Errorf("ipv4Any = %v, want 0.0.0.0/0", ipv4Any)
	}
}