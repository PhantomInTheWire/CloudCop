// Package security provides a high-level service for security scanning with AI summarization.
package security

import (
	"context"
	"fmt"
	"log"

	"cloudcop/api/internal/scanner"
	"cloudcop/api/internal/summarization"

	"github.com/aws/aws-sdk-go-v2/aws"
)

// Service orchestrates security scanning and AI summarization.
type Service struct {
	coordinator *scanner.Coordinator
	summClient  *summarization.Client
	summAddress string
	summEnabled bool
}

// Config holds configuration for the security service.
type Config struct {
	// AWSConfig is the AWS configuration for scanning.
	AWSConfig aws.Config
	// AccountID is the AWS account to scan.
	AccountID string
	// SummarizationAddress is the gRPC address for the AI service (e.g., "localhost:50051").
	SummarizationAddress string
	// EnableSummarization controls whether AI summarization is enabled.
	EnableSummarization bool
}

// NewService creates a new security service.
func NewService(cfg Config) (*Service, error) {
	coordinator := scanner.NewCoordinator(cfg.AWSConfig, cfg.AccountID)

	s := &Service{
		coordinator: coordinator,
		summAddress: cfg.SummarizationAddress,
		summEnabled: cfg.EnableSummarization,
	}

	return s, nil
}

// RegisterScanner registers a scanner factory with the coordinator.
func (s *Service) RegisterScanner(service string, factory func(aws.Config, string, string) scanner.ServiceScanner) {
	s.coordinator.RegisterScanner(service, factory)
}

// GetSupportedServices returns the list of registered scanner services.
func (s *Service) GetSupportedServices() []string {
	return s.coordinator.GetSupportedServices()
}

// Scan executes security scans and optionally summarizes findings with AI.
func (s *Service) Scan(ctx context.Context, config scanner.ScanConfig) (*scanner.ScanResultWithSummary, error) {
	// Execute the scan
	result, err := s.coordinator.StartScan(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("scan failed: %w", err)
	}

	// Return early if summarization is disabled or no failed findings
	if !s.summEnabled || result.FailedChecks == 0 {
		return &scanner.ScanResultWithSummary{
			ScanResult: result,
			Summary:    nil,
		}, nil
	}

	// Connect to summarization service
	summClient, err := s.connectSummarization()
	if err != nil {
		log.Printf("Warning: Could not connect to summarization service: %v", err)
		return &scanner.ScanResultWithSummary{
			ScanResult: result,
			Summary:    nil,
		}, nil
	}
	defer func() { _ = summClient.Close() }()

	// Generate scan ID
	scanID := fmt.Sprintf("scan-%d", result.StartedAt.Unix())

	// Call summarization service
	summResult, err := summClient.SummarizeFindings(ctx, scanID, config.AccountID, result.Findings)
	if err != nil {
		log.Printf("Warning: Summarization failed: %v", err)
		return &scanner.ScanResultWithSummary{
			ScanResult: result,
			Summary:    nil,
		}, nil
	}

	// Convert summarization result to ScanSummary
	summary := convertSummaryResult(summResult)

	return &scanner.ScanResultWithSummary{
		ScanResult: result,
		Summary:    summary,
	}, nil
}

// connectSummarization creates a connection to the summarization service.
func (s *Service) connectSummarization() (*summarization.Client, error) {
	if s.summClient != nil {
		return s.summClient, nil
	}

	if s.summAddress == "" {
		return nil, fmt.Errorf("summarization address not configured")
	}

	client, err := summarization.NewClient(s.summAddress)
	if err != nil {
		return nil, err
	}

	return client, nil
}

// Close closes any open connections.
func (s *Service) Close() error {
	if s.summClient != nil {
		return s.summClient.Close()
	}
	return nil
}

// convertSummaryResult converts a summarization.SummaryResult to scanner.ScanSummary.
func convertSummaryResult(r *summarization.SummaryResult) *scanner.ScanSummary {
	groups := make([]scanner.FindingGroupSummary, len(r.Groups))
	for i, g := range r.Groups {
		groups[i] = scanner.FindingGroupSummary{
			GroupID:      g.GroupID,
			Title:        g.Title,
			Service:      g.Service,
			CheckID:      g.CheckID,
			Severity:     g.Severity,
			FindingCount: g.FindingCount,
			ResourceIDs:  g.ResourceIDs,
			Summary:      g.Summary,
			Remedy:       g.Remedy,
		}
	}

	actions := make([]scanner.ActionItemSummary, len(r.Actions))
	for i, a := range r.Actions {
		actions[i] = scanner.ActionItemSummary{
			ActionID:    a.ActionID,
			Title:       a.Title,
			Description: a.Description,
			Severity:    a.Severity,
			Commands:    a.Commands,
			GroupID:     a.GroupID,
		}
	}

	return &scanner.ScanSummary{
		Groups:      groups,
		RiskLevel:   r.RiskSummary.RiskLevel,
		RiskScore:   r.RiskSummary.OverallScore,
		SummaryText: r.RiskSummary.SummaryText,
		Actions:     actions,
	}
}
