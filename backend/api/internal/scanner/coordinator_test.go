package scanner

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
)

// mockScanner implements ServiceScanner for testing
type mockScanner struct {
	service  string
	findings []Finding
	err      error
	delay    time.Duration
}

func (m *mockScanner) Service() string {
	return m.service
}

func (m *mockScanner) Scan(ctx context.Context, _ string) ([]Finding, error) {
	if m.delay > 0 {
		select {
		case <-time.After(m.delay):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
	return m.findings, m.err
}

func TestNewCoordinator(t *testing.T) {
	cfg := aws.Config{Region: "us-east-1"}
	accountID := "123456789012"

	coord := NewCoordinator(cfg, accountID)

	if coord == nil {
		t.Fatal("NewCoordinator returned nil")
	}
	if coord.accountID != accountID {
		t.Errorf("accountID = %v, want %v", coord.accountID, accountID)
	}
	if coord.scanners == nil {
		t.Error("scanners map not initialized")
	}
}

func TestCoordinator_RegisterScanner(t *testing.T) {
	coord := NewCoordinator(aws.Config{}, "123456789012")

	factory := func(_ aws.Config, _, _ string) ServiceScanner {
		return &mockScanner{service: "test"}
	}

	coord.RegisterScanner("test-service", factory)

	if len(coord.scanners) != 1 {
		t.Errorf("Expected 1 registered scanner, got %d", len(coord.scanners))
	}

	if _, exists := coord.scanners["test-service"]; !exists {
		t.Error("Scanner not registered with correct service name")
	}
}

func TestCoordinator_GetSupportedServices(t *testing.T) {
	coord := NewCoordinator(aws.Config{}, "123456789012")

	services := []string{"s3", "ec2", "iam"}
	for _, svc := range services {
		coord.RegisterScanner(svc, func(_ aws.Config, _, _ string) ServiceScanner {
			return &mockScanner{service: svc}
		})
	}

	supported := coord.GetSupportedServices()

	if len(supported) != len(services) {
		t.Errorf("GetSupportedServices() returned %d services, want %d", len(supported), len(services))
	}

	// Verify all services are present
	serviceMap := make(map[string]bool)
	for _, s := range supported {
		serviceMap[s] = true
	}
	for _, want := range services {
		if !serviceMap[want] {
			t.Errorf("GetSupportedServices() missing service %s", want)
		}
	}
}

func TestCoordinator_StartScan_Success(t *testing.T) {
	coord := NewCoordinator(aws.Config{}, "123456789012")

	// Register mock scanners
	coord.RegisterScanner("s3", func(_ aws.Config, _, _ string) ServiceScanner {
		return &mockScanner{
			service: "s3",
			findings: []Finding{
				{CheckID: "s3_test", Status: StatusPass, Severity: SeverityLow},
			},
		}
	})

	coord.RegisterScanner("ec2", func(_ aws.Config, _, _ string) ServiceScanner {
		return &mockScanner{
			service: "ec2",
			findings: []Finding{
				{CheckID: "ec2_test", Status: StatusFail, Severity: SeverityHigh},
			},
		}
	})

	config := ScanConfig{
		AccountID: "123456789012",
		Regions:   []string{"us-east-1"},
		Services:  []string{"s3", "ec2"},
	}

	result, err := coord.StartScan(context.Background(), config)

	if err != nil {
		t.Fatalf("StartScan() error = %v", err)
	}
	if result == nil {
		t.Fatal("StartScan() returned nil result")
	}
	if len(result.Findings) != 2 {
		t.Errorf("Expected 2 findings, got %d", len(result.Findings))
	}
	if result.PassedChecks != 1 {
		t.Errorf("PassedChecks = %d, want 1", result.PassedChecks)
	}
	if result.FailedChecks != 1 {
		t.Errorf("FailedChecks = %d, want 1", result.FailedChecks)
	}
	if result.TotalChecks != 2 {
		t.Errorf("TotalChecks = %d, want 2", result.TotalChecks)
	}
}

func TestCoordinator_StartScan_MultipleRegions(t *testing.T) {
	coord := NewCoordinator(aws.Config{}, "123456789012")

	coord.RegisterScanner("s3", func(_ aws.Config, region, _ string) ServiceScanner {
		return &mockScanner{
			service: "s3",
			findings: []Finding{
				{CheckID: "s3_test", Region: region, Status: StatusPass},
			},
		}
	})

	config := ScanConfig{
		AccountID: "123456789012",
		Regions:   []string{"us-east-1", "us-west-2", "eu-west-1"},
		Services:  []string{"s3"},
	}

	result, err := coord.StartScan(context.Background(), config)

	if err != nil {
		t.Fatalf("StartScan() error = %v", err)
	}
	if len(result.Findings) != 3 {
		t.Errorf("Expected 3 findings (1 per region), got %d", len(result.Findings))
	}
}

func TestCoordinator_StartScan_WithErrors(t *testing.T) {
	coord := NewCoordinator(aws.Config{}, "123456789012")

	testErr := errors.New("scanner error")
	coord.RegisterScanner("s3", func(_ aws.Config, _, _ string) ServiceScanner {
		return &mockScanner{
			service: "s3",
			err:     testErr,
		}
	})

	coord.RegisterScanner("ec2", func(_ aws.Config, _, _ string) ServiceScanner {
		return &mockScanner{
			service: "ec2",
			findings: []Finding{
				{CheckID: "ec2_test", Status: StatusPass},
			},
		}
	})

	config := ScanConfig{
		AccountID: "123456789012",
		Regions:   []string{"us-east-1"},
		Services:  []string{"s3", "ec2"},
	}

	result, err := coord.StartScan(context.Background(), config)

	// Should not return error, but should log it
	if err != nil {
		t.Fatalf("StartScan() should not return error for partial failures, got %v", err)
	}
	// Should still have findings from successful scanner
	if len(result.Findings) != 1 {
		t.Errorf("Expected 1 finding from successful scanner, got %d", len(result.Findings))
	}
}

func TestCoordinator_StartScan_NoValidTasks(t *testing.T) {
	coord := NewCoordinator(aws.Config{}, "123456789012")

	config := ScanConfig{
		AccountID: "123456789012",
		Regions:   []string{"us-east-1"},
		Services:  []string{"non-existent-service"},
	}

	_, err := coord.StartScan(context.Background(), config)

	if err == nil {
		t.Error("StartScan() should return error when no valid tasks")
	}
}

func TestCoordinator_StartScan_Parallel(t *testing.T) {
	coord := NewCoordinator(aws.Config{}, "123456789012")

	// Register scanner with delay to test parallelism
	coord.RegisterScanner("s3", func(_ aws.Config, _, _ string) ServiceScanner {
		return &mockScanner{
			service: "s3",
			delay:   100 * time.Millisecond,
			findings: []Finding{
				{CheckID: "s3_test", Status: StatusPass},
			},
		}
	})

	config := ScanConfig{
		AccountID: "123456789012",
		Regions:   []string{"us-east-1", "us-west-2", "eu-west-1"},
		Services:  []string{"s3"},
	}

	start := time.Now()
	result, err := coord.StartScan(context.Background(), config)
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("StartScan() error = %v", err)
	}

	// If sequential, would take 300ms+. If parallel, should be closer to 100ms
	if elapsed > 250*time.Millisecond {
		t.Errorf("Scan took %v, expected parallel execution to be faster", elapsed)
	}

	if len(result.Findings) != 3 {
		t.Errorf("Expected 3 findings, got %d", len(result.Findings))
	}
}

func TestCoordinator_StartScan_ResultMetadata(t *testing.T) {
	coord := NewCoordinator(aws.Config{}, "123456789012")

	coord.RegisterScanner("s3", func(_ aws.Config, _, _ string) ServiceScanner {
		return &mockScanner{
			service:  "s3",
			findings: []Finding{{CheckID: "test", Status: StatusPass}},
		}
	})

	config := ScanConfig{
		AccountID: "123456789012",
		Regions:   []string{"us-east-1"},
		Services:  []string{"s3"},
	}

	beforeScan := time.Now()
	result, err := coord.StartScan(context.Background(), config)
	afterScan := time.Now()

	if err != nil {
		t.Fatalf("StartScan() error = %v", err)
	}

	// Verify metadata
	if result.AccountID != config.AccountID {
		t.Errorf("AccountID = %v, want %v", result.AccountID, config.AccountID)
	}
	if len(result.Regions) != 1 || result.Regions[0] != "us-east-1" {
		t.Errorf("Regions = %v, want [us-east-1]", result.Regions)
	}
	if len(result.Services) != 1 || result.Services[0] != "s3" {
		t.Errorf("Services = %v, want [s3]", result.Services)
	}
	if result.StartedAt.Before(beforeScan) || result.StartedAt.After(afterScan) {
		t.Error("StartedAt timestamp out of expected range")
	}
	if result.CompletedAt.Before(result.StartedAt) {
		t.Error("CompletedAt should be after StartedAt")
	}
}

func TestGetDefaultRegions(t *testing.T) {
	regions := GetDefaultRegions()

	if len(regions) == 0 {
		t.Error("GetDefaultRegions() returned empty list")
	}

	// Check for some expected regions
	expectedRegions := []string{"us-east-1", "us-west-2", "eu-west-1"}
	regionMap := make(map[string]bool)
	for _, r := range regions {
		regionMap[r] = true
	}

	for _, expected := range expectedRegions {
		if !regionMap[expected] {
			t.Errorf("GetDefaultRegions() missing expected region %s", expected)
		}
	}
}

func TestGetAllRegions(t *testing.T) {
	cfg := aws.Config{Region: "us-east-1"}
	allRegions := GetAllRegions(context.Background(), cfg)
	defaultRegions := GetDefaultRegions()

	if len(allRegions) <= len(defaultRegions) {
		t.Error("GetAllRegions() should return more regions than GetDefaultRegions()")
	}

	// Verify all default regions are in all regions
	allMap := make(map[string]bool)
	for _, r := range allRegions {
		allMap[r] = true
	}

	for _, defaultRegion := range defaultRegions {
		if !allMap[defaultRegion] {
			t.Errorf("GetAllRegions() missing default region %s", defaultRegion)
		}
	}

	// Check for some expected regions
	expectedRegions := []string{"us-east-1", "ap-southeast-1", "eu-central-1", "sa-east-1"}
	for _, expected := range expectedRegions {
		if !allMap[expected] {
			t.Errorf("GetAllRegions() missing expected region %s", expected)
		}
	}
}

func TestCoordinator_ContextCancellation(t *testing.T) {
	coord := NewCoordinator(aws.Config{}, "123456789012")

	coord.RegisterScanner("s3", func(_ aws.Config, _, _ string) ServiceScanner {
		return &mockScanner{
			service: "s3",
			delay:   5 * time.Second,
			findings: []Finding{
				{CheckID: "s3_test", Status: StatusPass},
			},
		}
	})

	config := ScanConfig{
		AccountID: "123456789012",
		Regions:   []string{"us-east-1"},
		Services:  []string{"s3"},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	start := time.Now()
	result, err := coord.StartScan(ctx, config)
	elapsed := time.Since(start)

	// Context cancellation should stop the scan quickly
	if elapsed > 2*time.Second {
		t.Errorf("Scan took %v with cancelled context, expected faster completion", elapsed)
	}

	_ = err
	_ = result
}

func TestScanTaskResult(t *testing.T) {
	task := ScanTask{
		Service: "s3",
		Region:  "us-east-1",
	}

	result := ScanTaskResult{
		Task: task,
		Findings: []Finding{
			{CheckID: "test1", Status: StatusPass},
			{CheckID: "test2", Status: StatusFail},
		},
		Error: nil,
	}

	if result.Task.Service != "s3" {
		t.Errorf("Task.Service = %v, want s3", result.Task.Service)
	}
	if len(result.Findings) != 2 {
		t.Errorf("len(Findings) = %d, want 2", len(result.Findings))
	}
	if result.Error != nil {
		t.Errorf("Error should be nil, got %v", result.Error)
	}
}
