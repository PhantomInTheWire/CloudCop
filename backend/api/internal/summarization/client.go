// Package summarization provides a gRPC client for the AI summarization service.
package summarization

import (
	"context"
	"fmt"

	pb "cloudcop/api/internal/grpc"
	"cloudcop/api/internal/scanner"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Client wraps the gRPC client for summarization.
type Client struct {
	conn   *grpc.ClientConn
	client pb.SummarizationServiceClient
}

// NewClient creates a new summarization client.
func NewClient(address string) (*Client, error) {
	conn, err := grpc.NewClient(address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to summarization service: %w", err)
	}

	return &Client{
		conn:   conn,
		client: pb.NewSummarizationServiceClient(conn),
	}, nil
}

// Close closes the gRPC connection.
func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// SummarizeFindings sends findings to the AI service for summarization.
func (c *Client) SummarizeFindings(ctx context.Context, scanID, accountID string, findings []scanner.Finding) (*SummaryResult, error) {
	// Convert scanner findings to protobuf format
	pbFindings := make([]*pb.Finding, len(findings))
	for i, f := range findings {
		pbFindings[i] = convertFinding(f)
	}

	req := &pb.SummarizeFindingsRequest{
		ScanId:    scanID,
		AccountId: accountID,
		Findings:  pbFindings,
		Options: &pb.SummarizationOptions{
			IncludeTerraformFixes: true,
			GroupByService:        true,
			GroupBySeverity:       false,
			MaxGroups:             50,
		},
	}

	resp, err := c.client.SummarizeFindings(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("summarization failed: %w", err)
	}

	return convertResponse(resp), nil
}

// SummaryResult contains the summarized findings.
type SummaryResult struct {
	ScanID      string
	Groups      []FindingGroup
	RiskSummary RiskSummary
	Actions     []ActionItem
}

// FindingGroup represents a group of similar findings.
type FindingGroup struct {
	GroupID           string
	Title             string
	Description       string
	Severity          string
	FindingCount      int
	ResourceIDs       []string
	CheckID           string
	Service           string
	Compliance        []string
	RiskScore         int
	RecommendedAction string
}

// RiskSummary contains overall risk metrics.
type RiskSummary struct {
	OverallScore  int
	CriticalCount int
	HighCount     int
	MediumCount   int
	LowCount      int
	PassedCount   int
	RiskLevel     string
	SummaryText   string
}

// ActionItem represents a recommended action.
type ActionItem struct {
	ActionID     string
	ActionType   string
	Severity     string
	Title        string
	Description  string
	GroupID      string
	TerraformFix *TerraformFix
}

// TerraformFix contains generated Terraform code.
type TerraformFix struct {
	ResourceType string
	ResourceName string
	Code         string
	Explanation  string
}

func convertFinding(f scanner.Finding) *pb.Finding {
	return &pb.Finding{
		Service:       f.Service,
		Region:        f.Region,
		ResourceId:    f.ResourceID,
		CheckId:       f.CheckID,
		Status:        convertStatus(f.Status),
		Severity:      convertSeverity(f.Severity),
		Title:         f.Title,
		Description:   f.Description,
		Compliance:    f.Compliance,
		TimestampUnix: f.Timestamp.Unix(),
	}
}

func convertStatus(s scanner.FindingStatus) pb.FindingStatus {
	switch s {
	case scanner.StatusPass:
		return pb.FindingStatus_FINDING_STATUS_PASS
	case scanner.StatusFail:
		return pb.FindingStatus_FINDING_STATUS_FAIL
	default:
		return pb.FindingStatus_FINDING_STATUS_UNSPECIFIED
	}
}

func convertSeverity(s scanner.Severity) pb.Severity {
	switch s {
	case scanner.SeverityLow:
		return pb.Severity_SEVERITY_LOW
	case scanner.SeverityMedium:
		return pb.Severity_SEVERITY_MEDIUM
	case scanner.SeverityHigh:
		return pb.Severity_SEVERITY_HIGH
	case scanner.SeverityCritical:
		return pb.Severity_SEVERITY_CRITICAL
	default:
		return pb.Severity_SEVERITY_UNSPECIFIED
	}
}

func convertResponse(resp *pb.SummarizeFindingsResponse) *SummaryResult {
	groups := make([]FindingGroup, len(resp.Groups))
	for i, g := range resp.Groups {
		groups[i] = FindingGroup{
			GroupID:           g.GroupId,
			Title:             g.Title,
			Description:       g.Description,
			Severity:          severityToString(g.Severity),
			FindingCount:      int(g.FindingCount),
			ResourceIDs:       g.ResourceIds,
			CheckID:           g.CheckId,
			Service:           g.Service,
			Compliance:        g.Compliance,
			RiskScore:         int(g.RiskScore),
			RecommendedAction: actionToString(g.RecommendedAction),
		}
	}

	actions := make([]ActionItem, len(resp.ActionItems))
	for i, a := range resp.ActionItems {
		action := ActionItem{
			ActionID:    a.ActionId,
			ActionType:  actionToString(a.ActionType),
			Severity:    severityToString(a.Severity),
			Title:       a.Title,
			Description: a.Description,
			GroupID:     a.GroupId,
		}
		if a.TerraformFix != nil {
			action.TerraformFix = &TerraformFix{
				ResourceType: a.TerraformFix.ResourceType,
				ResourceName: a.TerraformFix.ResourceName,
				Code:         a.TerraformFix.Code,
				Explanation:  a.TerraformFix.Explanation,
			}
		}
		actions[i] = action
	}

	var riskSummary RiskSummary
	if resp.RiskSummary != nil {
		riskSummary = RiskSummary{
			OverallScore:  int(resp.RiskSummary.OverallScore),
			CriticalCount: int(resp.RiskSummary.CriticalCount),
			HighCount:     int(resp.RiskSummary.HighCount),
			MediumCount:   int(resp.RiskSummary.MediumCount),
			LowCount:      int(resp.RiskSummary.LowCount),
			PassedCount:   int(resp.RiskSummary.PassedCount),
			RiskLevel:     resp.RiskSummary.RiskLevel,
			SummaryText:   resp.RiskSummary.SummaryText,
		}
	}

	return &SummaryResult{
		ScanID:      resp.ScanId,
		Groups:      groups,
		RiskSummary: riskSummary,
		Actions:     actions,
	}
}

func severityToString(s pb.Severity) string {
	switch s {
	case pb.Severity_SEVERITY_LOW:
		return "LOW"
	case pb.Severity_SEVERITY_MEDIUM:
		return "MEDIUM"
	case pb.Severity_SEVERITY_HIGH:
		return "HIGH"
	case pb.Severity_SEVERITY_CRITICAL:
		return "CRITICAL"
	default:
		return "UNKNOWN"
	}
}

func actionToString(a pb.ActionType) string {
	switch a {
	case pb.ActionType_ACTION_TYPE_SUGGEST_FIX:
		return "SUGGEST_FIX"
	case pb.ActionType_ACTION_TYPE_ALERT:
		return "ALERT"
	case pb.ActionType_ACTION_TYPE_ESCALATE:
		return "ESCALATE"
	default:
		return "NONE"
	}
}
