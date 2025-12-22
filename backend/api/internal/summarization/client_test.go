package summarization

import (
	"testing"
	"time"

	pb "cloudcop/api/internal/grpc"
	"cloudcop/api/internal/scanner"
)

func TestConvertFinding(t *testing.T) {
	finding := scanner.Finding{
		Service:     "s3",
		Region:      "us-east-1",
		ResourceID:  "my-bucket",
		CheckID:     "s3_bucket_encryption",
		Status:      scanner.StatusFail,
		Severity:    scanner.SeverityHigh,
		Title:       "S3 bucket encryption is not enabled",
		Description: "Bucket my-bucket does not have encryption",
		Compliance:  []string{"CIS", "SOC2"},
		Timestamp:   time.Now(),
	}

	pbFinding := convertFinding(finding)

	if pbFinding.Service != "s3" {
		t.Errorf("Service = %v, want s3", pbFinding.Service)
	}
	if pbFinding.ResourceId != "my-bucket" {
		t.Errorf("ResourceId = %v, want my-bucket", pbFinding.ResourceId)
	}
	if pbFinding.Status != pb.FindingStatus_FINDING_STATUS_FAIL {
		t.Errorf("Status = %v, want FAIL", pbFinding.Status)
	}
	if pbFinding.Severity != pb.Severity_SEVERITY_HIGH {
		t.Errorf("Severity = %v, want HIGH", pbFinding.Severity)
	}
}

func TestConvertStatus(t *testing.T) {
	tests := []struct {
		input    scanner.FindingStatus
		expected pb.FindingStatus
	}{
		{scanner.StatusPass, pb.FindingStatus_FINDING_STATUS_PASS},
		{scanner.StatusFail, pb.FindingStatus_FINDING_STATUS_FAIL},
		{"unknown", pb.FindingStatus_FINDING_STATUS_UNSPECIFIED},
	}

	for _, tt := range tests {
		result := convertStatus(tt.input)
		if result != tt.expected {
			t.Errorf("convertStatus(%v) = %v, want %v", tt.input, result, tt.expected)
		}
	}
}

func TestConvertSeverity(t *testing.T) {
	tests := []struct {
		input    scanner.Severity
		expected pb.Severity
	}{
		{scanner.SeverityLow, pb.Severity_SEVERITY_LOW},
		{scanner.SeverityMedium, pb.Severity_SEVERITY_MEDIUM},
		{scanner.SeverityHigh, pb.Severity_SEVERITY_HIGH},
		{scanner.SeverityCritical, pb.Severity_SEVERITY_CRITICAL},
		{"unknown", pb.Severity_SEVERITY_UNSPECIFIED},
	}

	for _, tt := range tests {
		result := convertSeverity(tt.input)
		if result != tt.expected {
			t.Errorf("convertSeverity(%v) = %v, want %v", tt.input, result, tt.expected)
		}
	}
}

func TestSeverityToString(t *testing.T) {
	tests := []struct {
		input    pb.Severity
		expected string
	}{
		{pb.Severity_SEVERITY_LOW, "LOW"},
		{pb.Severity_SEVERITY_MEDIUM, "MEDIUM"},
		{pb.Severity_SEVERITY_HIGH, "HIGH"},
		{pb.Severity_SEVERITY_CRITICAL, "CRITICAL"},
		{pb.Severity_SEVERITY_UNSPECIFIED, "UNKNOWN"},
	}

	for _, tt := range tests {
		result := severityToString(tt.input)
		if result != tt.expected {
			t.Errorf("severityToString(%v) = %v, want %v", tt.input, result, tt.expected)
		}
	}
}

func TestActionToString(t *testing.T) {
	tests := []struct {
		input    pb.ActionType
		expected string
	}{
		{pb.ActionType_ACTION_TYPE_SUGGEST_FIX, "SUGGEST_FIX"},
		{pb.ActionType_ACTION_TYPE_ALERT, "ALERT"},
		{pb.ActionType_ACTION_TYPE_ESCALATE, "ESCALATE"},
		{pb.ActionType_ACTION_TYPE_UNSPECIFIED, "NONE"},
	}

	for _, tt := range tests {
		result := actionToString(tt.input)
		if result != tt.expected {
			t.Errorf("actionToString(%v) = %v, want %v", tt.input, result, tt.expected)
		}
	}
}

func TestConvertResponse(t *testing.T) {
	resp := &pb.SummarizeFindingsResponse{
		ScanId: "scan-123",
		Groups: []*pb.FindingGroup{
			{
				GroupId:           "s3:s3_bucket_encryption",
				Title:             "5 S3 resources failed s3_bucket_encryption",
				Description:       "Buckets without encryption",
				Severity:          pb.Severity_SEVERITY_HIGH,
				FindingCount:      5,
				ResourceIds:       []string{"bucket-1", "bucket-2"},
				CheckId:           "s3_bucket_encryption",
				Service:           "s3",
				Compliance:        []string{"CIS"},
				RiskScore:         75,
				RecommendedAction: pb.ActionType_ACTION_TYPE_ALERT,
				Summary:           "5 S3 buckets lack encryption",
				Remedy:            "Enable SSE-S3 or SSE-KMS encryption",
			},
		},
		RiskSummary: &pb.RiskSummary{
			OverallScore:  75,
			CriticalCount: 0,
			HighCount:     5,
			MediumCount:   2,
			LowCount:      1,
			PassedCount:   10,
			RiskLevel:     "HIGH",
			SummaryText:   "Found 8 issues",
		},
		ActionItems: []*pb.ActionItem{
			{
				ActionId:    "action_1",
				ActionType:  pb.ActionType_ACTION_TYPE_ALERT,
				Severity:    pb.Severity_SEVERITY_HIGH,
				Title:       "Fix encryption",
				Description: "Enable encryption",
				GroupId:     "s3:s3_bucket_encryption",
				Commands:    []string{"aws s3api put-bucket-encryption --bucket bucket-1 ..."},
			},
		},
	}

	result := convertResponse(resp)

	if result.ScanID != "scan-123" {
		t.Errorf("ScanID = %v, want scan-123", result.ScanID)
	}
	if len(result.Groups) != 1 {
		t.Fatalf("Groups count = %d, want 1", len(result.Groups))
	}
	if result.Groups[0].Title != "5 S3 resources failed s3_bucket_encryption" {
		t.Errorf("Group title = %v, want expected", result.Groups[0].Title)
	}
	if result.Groups[0].Summary != "5 S3 buckets lack encryption" {
		t.Errorf("Group summary = %v, want expected", result.Groups[0].Summary)
	}
	if result.Groups[0].Remedy != "Enable SSE-S3 or SSE-KMS encryption" {
		t.Errorf("Group remedy = %v, want expected", result.Groups[0].Remedy)
	}
	if result.RiskSummary.OverallScore != 75 {
		t.Errorf("OverallScore = %d, want 75", result.RiskSummary.OverallScore)
	}
	if len(result.Actions) != 1 {
		t.Fatalf("Actions count = %d, want 1", len(result.Actions))
	}
	if len(result.Actions[0].Commands) != 1 {
		t.Errorf("Commands count = %d, want 1", len(result.Actions[0].Commands))
	}
}
