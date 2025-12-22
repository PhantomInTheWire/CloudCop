// Package scanner provides AWS security scanning infrastructure for CloudCop.
package scanner

import (
	"context"
	"time"
)

// FindingStatus represents the status of a security finding.
type FindingStatus string

const (
	// StatusPass indicates the resource passed the security check.
	StatusPass FindingStatus = "PASS"
	// StatusFail indicates the resource failed the security check.
	StatusFail FindingStatus = "FAIL"
)

// Severity represents the severity level of a security finding.
type Severity string

const (
	// SeverityLow indicates a low-priority finding.
	SeverityLow Severity = "LOW"
	// SeverityMedium indicates a medium-priority finding.
	SeverityMedium Severity = "MEDIUM"
	// SeverityHigh indicates a high-priority finding.
	SeverityHigh Severity = "HIGH"
	// SeverityCritical indicates a critical finding requiring immediate attention.
	SeverityCritical Severity = "CRITICAL"
)

// Finding represents a security finding from a scan.
type Finding struct {
	// Service is the AWS service name (e.g., "s3", "ec2").
	Service string `json:"service"`
	// Region is the AWS region where the finding was detected.
	Region string `json:"region"`
	// ResourceID is the AWS resource identifier (ARN or ID).
	ResourceID string `json:"resource_id"`
	// CheckID is the unique identifier for the security check.
	CheckID string `json:"check_id"`
	// Status indicates whether the check passed or failed.
	Status FindingStatus `json:"status"`
	// Severity indicates the importance of the finding.
	Severity Severity `json:"severity"`
	// Title is a short description of the finding.
	Title string `json:"title"`
	// Description provides detailed information about the finding.
	Description string `json:"description"`
	// Compliance lists the compliance frameworks this check maps to.
	Compliance []string `json:"compliance"`
	// Timestamp is when the finding was detected.
	Timestamp time.Time `json:"timestamp"`
}

// ServiceScanner defines the interface for service-specific scanners.
type ServiceScanner interface {
	// Scan executes all security checks for the service in the specified region.
	Scan(ctx context.Context, region string) ([]Finding, error)
	// Service returns the AWS service name (e.g., "s3", "ec2").
	Service() string
}

// ScanConfig holds configuration for a security scan.
type ScanConfig struct {
	// AccountID is the AWS account being scanned.
	AccountID string
	// Regions is the list of AWS regions to scan.
	Regions []string
	// Services is the list of AWS services to scan.
	Services []string
}

// ScanResult holds the aggregated results of a security scan.
type ScanResult struct {
	// AccountID is the AWS account that was scanned.
	AccountID string `json:"account_id"`
	// Regions is the list of regions that were scanned.
	Regions []string `json:"regions"`
	// Services is the list of services that were scanned.
	Services []string `json:"services"`
	// Findings is the list of all security findings.
	Findings []Finding `json:"findings"`
	// StartedAt is when the scan started.
	StartedAt time.Time `json:"started_at"`
	// CompletedAt is when the scan completed.
	CompletedAt time.Time `json:"completed_at"`
	// TotalChecks is the total number of checks executed.
	TotalChecks int `json:"total_checks"`
	// PassedChecks is the number of checks that passed.
	PassedChecks int `json:"passed_checks"`
	// FailedChecks is the number of checks that failed.
	FailedChecks int `json:"failed_checks"`
}

// ScanItem represents a scan result for a specific service/region combination.
type ScanItem struct {
	// Service is the AWS service name.
	Service string `json:"service"`
	// Region is the AWS region.
	Region string `json:"region"`
	// Findings is the list of findings for this service/region.
	Findings []Finding `json:"findings"`
}

// ScanSummary contains AI-generated summaries and remediation commands.
type ScanSummary struct {
	// Groups contains grouped findings with AI summaries.
	Groups []FindingGroupSummary `json:"groups"`
	// RiskLevel is the overall risk level (LOW, MEDIUM, HIGH, CRITICAL).
	RiskLevel string `json:"risk_level"`
	// RiskScore is the overall risk score (0-100).
	RiskScore int `json:"risk_score"`
	// SummaryText is an AI-generated summary of the scan results.
	SummaryText string `json:"summary_text"`
	// Actions contains recommended actions with CLI commands.
	Actions []ActionItemSummary `json:"actions"`
}

// FindingGroupSummary contains AI-generated summary for a group of findings.
type FindingGroupSummary struct {
	// GroupID is the unique identifier for this group.
	GroupID string `json:"group_id"`
	// Title is a short description of the group.
	Title string `json:"title"`
	// Service is the AWS service.
	Service string `json:"service"`
	// CheckID is the security check identifier.
	CheckID string `json:"check_id"`
	// Severity is the highest severity in the group.
	Severity string `json:"severity"`
	// FindingCount is the number of findings in the group.
	FindingCount int `json:"finding_count"`
	// ResourceIDs are the affected resources.
	ResourceIDs []string `json:"resource_ids"`
	// Summary is an AI-generated summary of the issue.
	Summary string `json:"summary"`
	// Remedy is an AI-generated remediation description.
	Remedy string `json:"remedy"`
}

// ActionItemSummary contains a recommended action with CLI commands.
type ActionItemSummary struct {
	// ActionID is the unique identifier for this action.
	ActionID string `json:"action_id"`
	// Title is a short description of the action.
	Title string `json:"title"`
	// Description provides details about the action.
	Description string `json:"description"`
	// Severity indicates the priority.
	Severity string `json:"severity"`
	// Commands are AWS CLI commands for remediation.
	Commands []string `json:"commands"`
	// GroupID is the related finding group.
	GroupID string `json:"group_id"`
}

// ScanResultWithSummary combines scan results with AI-generated summaries.
type ScanResultWithSummary struct {
	// ScanResult contains the raw scan results.
	*ScanResult
	// Summary contains AI-generated summaries (nil if summarization was skipped).
	Summary *ScanSummary `json:"summary,omitempty"`
}
