// Package scanner provides AWS security scanning infrastructure for CloudCop.
package scanner

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
)

// Coordinator orchestrates parallel scanning across regions and services.
type Coordinator struct {
	cfg       aws.Config
	accountID string
	scanners  map[string]func(aws.Config, string, string) ServiceScanner
}

// NewCoordinator creates a new scan coordinator with an initialized scanner factory registry.
func NewCoordinator(cfg aws.Config, accountID string) *Coordinator {
	return &Coordinator{
		cfg:       cfg,
		accountID: accountID,
		scanners:  make(map[string]func(aws.Config, string, string) ServiceScanner),
	}
}

// RegisterScanner registers a scanner factory for a service.
func (c *Coordinator) RegisterScanner(service string, factory func(aws.Config, string, string) ServiceScanner) {
	c.scanners[service] = factory
}

// ScanTask represents a single scan task for a service/region combination.
type ScanTask struct {
	Service string
	Region  string
}

// ScanTaskResult holds the result of a single scan task.
type ScanTaskResult struct {
	Task     ScanTask
	Findings []Finding
	Error    error
}

// StartScan executes security scans across the specified regions and services.
func (c *Coordinator) StartScan(ctx context.Context, config ScanConfig) (*ScanResult, error) {
	startedAt := time.Now()

	// Build list of scan tasks
	var tasks []ScanTask
	for _, region := range config.Regions {
		for _, service := range config.Services {
			if _, exists := c.scanners[service]; exists {
				tasks = append(tasks, ScanTask{Service: service, Region: region})
			} else {
				log.Printf("Warning: No scanner registered for service %s", service)
			}
		}
	}

	if len(tasks) == 0 {
		return nil, fmt.Errorf("no valid scan tasks: check that services have registered scanners")
	}

	// Execute tasks in parallel
	results := c.executeParallel(ctx, tasks)

	// Aggregate results
	var allFindings []Finding
	var scanErrors []error

	for _, result := range results {
		if result.Error != nil {
			scanErrors = append(scanErrors, fmt.Errorf("%s/%s: %w", result.Task.Service, result.Task.Region, result.Error))
			continue
		}
		allFindings = append(allFindings, result.Findings...)
	}

	// Count passed and failed checks
	passedChecks := 0
	failedChecks := 0
	for _, f := range allFindings {
		if f.Status == StatusPass {
			passedChecks++
		} else {
			failedChecks++
		}
	}

	// Log any errors (but don't fail the entire scan)
	for _, err := range scanErrors {
		log.Printf("Scan error: %v", err)
	}

	return &ScanResult{
		AccountID:    config.AccountID,
		Regions:      config.Regions,
		Services:     config.Services,
		Findings:     allFindings,
		StartedAt:    startedAt,
		CompletedAt:  time.Now(),
		TotalChecks:  len(allFindings),
		PassedChecks: passedChecks,
		FailedChecks: failedChecks,
	}, nil
}

// executeParallel runs scan tasks concurrently using a bounded worker pool.
func (c *Coordinator) executeParallel(ctx context.Context, tasks []ScanTask) []ScanTaskResult {
	const maxWorkers = 10 // Limit concurrent scans to prevent overwhelming APIs

	var wg sync.WaitGroup
	resultsChan := make(chan ScanTaskResult, len(tasks))
	tasksChan := make(chan ScanTask, len(tasks))

	// Start worker pool
	for i := 0; i < maxWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for task := range tasksChan {
				// Check for context cancellation before processing
				select {
				case <-ctx.Done():
					resultsChan <- ScanTaskResult{
						Task:  task,
						Error: ctx.Err(),
					}
					continue
				default:
				}

				result := ScanTaskResult{Task: task}

				// Create scanner for this service/region
				factory, exists := c.scanners[task.Service]
				if !exists {
					result.Error = fmt.Errorf("no scanner registered for service %s", task.Service)
					resultsChan <- result
					continue
				}

				// Create regional config
				regionalCfg := c.cfg.Copy()
				regionalCfg.Region = task.Region

				scanner := factory(regionalCfg, task.Region, c.accountID)

				// Execute scan with context
				findings, err := scanner.Scan(ctx, task.Region)
				if err != nil {
					result.Error = err
					resultsChan <- result
					continue
				}

				result.Findings = findings
				resultsChan <- result
			}
		}()
	}

	// Send tasks to workers
	go func() {
		for _, task := range tasks {
			tasksChan <- task
		}
		close(tasksChan)
	}()

	// Wait for all workers to complete
	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	// Collect results
	var results []ScanTaskResult
	for result := range resultsChan {
		results = append(results, result)
	}

	return results
}

// GetSupportedServices returns the list of services that have registered scanners.
func (c *Coordinator) GetSupportedServices() []string {
	services := make([]string, 0, len(c.scanners))
	for service := range c.scanners {
		services = append(services, service)
	}
	return services
}

// GetDefaultRegions returns the default AWS regions to scan.
func GetDefaultRegions() []string {
	return []string{
		"us-east-1",
		"us-east-2",
		"us-west-1",
		"us-west-2",
		"eu-west-1",
		"eu-west-2",
		"eu-central-1",
		"ap-southeast-1",
		"ap-southeast-2",
		"ap-northeast-1",
	}
}

var (
	cachedRegions   []string
	cachedRegionsMu sync.RWMutex
	fallbackRegions = []string{
		"us-east-1", "us-east-2", "us-west-1", "us-west-2",
		"af-south-1",
		"ap-east-1", "ap-south-1", "ap-south-2", "ap-southeast-1", "ap-southeast-2",
		"ap-southeast-3", "ap-northeast-1", "ap-northeast-2", "ap-northeast-3",
		"ca-central-1",
		"eu-central-1", "eu-central-2", "eu-west-1", "eu-west-2", "eu-west-3",
		"eu-south-1", "eu-south-2", "eu-north-1",
		"me-south-1", "me-central-1",
		"sa-east-1",
	}
)

// GetAllRegions returns all AWS regions dynamically via EC2 DescribeRegions API.
// Results are cached after the first successful call. Falls back to a hardcoded
// list if the API call fails.
func GetAllRegions(ctx context.Context, cfg aws.Config) []string {
	cachedRegionsMu.RLock()
	if len(cachedRegions) > 0 {
		defer cachedRegionsMu.RUnlock()
		return cachedRegions
	}
	cachedRegionsMu.RUnlock()

	regions, err := fetchRegionsFromEC2(ctx, cfg)
	if err != nil {
		log.Printf("Failed to fetch regions from EC2 API, using fallback: %v", err)
		return fallbackRegions
	}

	cachedRegionsMu.Lock()
	cachedRegions = regions
	cachedRegionsMu.Unlock()

	return regions
}

// fetchRegionsFromEC2 calls EC2 DescribeRegions API to get all available regions.
func fetchRegionsFromEC2(ctx context.Context, cfg aws.Config) ([]string, error) {
	client := ec2.NewFromConfig(cfg)

	output, err := client.DescribeRegions(ctx, &ec2.DescribeRegionsInput{
		AllRegions: aws.Bool(true),
	})
	if err != nil {
		return nil, fmt.Errorf("DescribeRegions failed: %w", err)
	}

	regions := make([]string, 0, len(output.Regions))
	for _, region := range output.Regions {
		if region.RegionName != nil {
			regions = append(regions, *region.RegionName)
		}
	}

	if len(regions) == 0 {
		return nil, fmt.Errorf("DescribeRegions returned no regions")
	}

	return regions, nil
}
