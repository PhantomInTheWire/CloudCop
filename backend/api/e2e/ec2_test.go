package e2e

import (
	"context"
	"testing"
	"time"

	"cloudcop/api/internal/scanner"
	"cloudcop/api/internal/scanner/ec2"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsec2 "github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

// TestEC2Scanner_E2E tests the EC2 scanner against LocalStack
func TestEC2Scanner_E2E(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	if !IsLocalStackRunning(ctx) {
		t.Skip("LocalStack is not running. Start it with: docker compose -f e2e/docker-compose.yml up -d")
	}

	cfg := NewDefaultConfig()
	awsCfg, err := cfg.GetAWSConfig(ctx)
	if err != nil {
		t.Fatalf("Failed to get AWS config: %v", err)
	}

	ec2Client, err := cfg.NewEC2Client(ctx)
	if err != nil {
		t.Fatalf("Failed to create EC2 client: %v", err)
	}

	// Create a VPC first (required for EC2 instances in LocalStack)
	vpcOutput, err := ec2Client.CreateVpc(ctx, &awsec2.CreateVpcInput{
		CidrBlock: aws.String("10.0.0.0/16"),
	})
	if err != nil {
		t.Fatalf("Failed to create VPC: %v", err)
	}
	vpcID := aws.ToString(vpcOutput.Vpc.VpcId)
	defer func() {
		_, _ = ec2Client.DeleteVpc(ctx, &awsec2.DeleteVpcInput{VpcId: aws.String(vpcID)})
	}()

	// Create a subnet
	subnetOutput, err := ec2Client.CreateSubnet(ctx, &awsec2.CreateSubnetInput{
		VpcId:     aws.String(vpcID),
		CidrBlock: aws.String("10.0.1.0/24"),
	})
	if err != nil {
		t.Fatalf("Failed to create subnet: %v", err)
	}
	subnetID := aws.ToString(subnetOutput.Subnet.SubnetId)
	defer func() {
		_, _ = ec2Client.DeleteSubnet(ctx, &awsec2.DeleteSubnetInput{SubnetId: aws.String(subnetID)})
	}()

	tests := []struct {
		name           string
		setup          func(t *testing.T) (instanceID string, cleanup func())
		expectedChecks map[string]scanner.FindingStatus
	}{
		{
			name: "instance_without_public_ip",
			setup: func(t *testing.T) (string, func()) {
				// Create a security group
				sgOutput, err := ec2Client.CreateSecurityGroup(ctx, &awsec2.CreateSecurityGroupInput{
					GroupName:   aws.String("test-sg-private-" + time.Now().Format("150405")),
					Description: aws.String("Test security group"),
					VpcId:       aws.String(vpcID),
				})
				if err != nil {
					t.Fatalf("Failed to create security group: %v", err)
				}
				sgID := aws.ToString(sgOutput.GroupId)

				// Get a valid AMI ID (LocalStack provides mock AMIs)
				amiID := "ami-12345678" // LocalStack accepts any AMI ID

				// Run instance without public IP
				runOutput, err := ec2Client.RunInstances(ctx, &awsec2.RunInstancesInput{
					ImageId:          aws.String(amiID),
					InstanceType:     types.InstanceTypeT2Micro,
					MinCount:         aws.Int32(1),
					MaxCount:         aws.Int32(1),
					SubnetId:         aws.String(subnetID),
					SecurityGroupIds: []string{sgID},
					NetworkInterfaces: []types.InstanceNetworkInterfaceSpecification{
						{
							DeviceIndex:              aws.Int32(0),
							SubnetId:                 aws.String(subnetID),
							AssociatePublicIpAddress: aws.Bool(false),
							Groups:                   []string{sgID},
						},
					},
				})
				if err != nil {
					t.Fatalf("Failed to run instance: %v", err)
				}
				instanceID := aws.ToString(runOutput.Instances[0].InstanceId)

				return instanceID, func() {
					_, _ = ec2Client.TerminateInstances(ctx, &awsec2.TerminateInstancesInput{
						InstanceIds: []string{instanceID},
					})
					// Wait for termination before deleting SG
					time.Sleep(2 * time.Second)
					_, _ = ec2Client.DeleteSecurityGroup(ctx, &awsec2.DeleteSecurityGroupInput{
						GroupId: aws.String(sgID),
					})
				}
			},
			expectedChecks: map[string]scanner.FindingStatus{
				"ec2_public_ip": scanner.StatusPass,
			},
		},
		{
			name: "instance_with_public_ip",
			setup: func(t *testing.T) (string, func()) {
				// Create a security group
				sgOutput, err := ec2Client.CreateSecurityGroup(ctx, &awsec2.CreateSecurityGroupInput{
					GroupName:   aws.String("test-sg-public-" + time.Now().Format("150405")),
					Description: aws.String("Test security group for public instance"),
					VpcId:       aws.String(vpcID),
				})
				if err != nil {
					t.Fatalf("Failed to create security group: %v", err)
				}
				sgID := aws.ToString(sgOutput.GroupId)

				amiID := "ami-12345678"

				// Run instance with public IP
				runOutput, err := ec2Client.RunInstances(ctx, &awsec2.RunInstancesInput{
					ImageId:      aws.String(amiID),
					InstanceType: types.InstanceTypeT2Micro,
					MinCount:     aws.Int32(1),
					MaxCount:     aws.Int32(1),
					NetworkInterfaces: []types.InstanceNetworkInterfaceSpecification{
						{
							DeviceIndex:              aws.Int32(0),
							SubnetId:                 aws.String(subnetID),
							AssociatePublicIpAddress: aws.Bool(true),
							Groups:                   []string{sgID},
						},
					},
				})
				if err != nil {
					t.Fatalf("Failed to run instance: %v", err)
				}
				instanceID := aws.ToString(runOutput.Instances[0].InstanceId)

				return instanceID, func() {
					_, _ = ec2Client.TerminateInstances(ctx, &awsec2.TerminateInstancesInput{
						InstanceIds: []string{instanceID},
					})
					time.Sleep(2 * time.Second)
					_, _ = ec2Client.DeleteSecurityGroup(ctx, &awsec2.DeleteSecurityGroupInput{
						GroupId: aws.String(sgID),
					})
				}
			},
			expectedChecks: map[string]scanner.FindingStatus{
				"ec2_public_ip": scanner.StatusFail,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			instanceID, cleanup := tt.setup(t)
			defer cleanup()

			// Give LocalStack a moment to fully create the instance
			time.Sleep(1 * time.Second)

			// Run the scanner
			ec2Scanner := ec2.NewScanner(awsCfg, DefaultRegion, TestAccountID)
			findings, err := ec2Scanner.Scan(ctx, DefaultRegion)
			if err != nil {
				t.Fatalf("Scan failed: %v", err)
			}

			// Filter findings for our instance
			instanceFindings := filterFindingsByResource(findings, instanceID)

			t.Logf("Found %d findings for instance %s", len(instanceFindings), instanceID)
			for _, f := range instanceFindings {
				t.Logf("  %s: %s (%s)", f.CheckID, f.Status, f.Title)
			}

			// Verify expected checks
			for checkID, expectedStatus := range tt.expectedChecks {
				finding := findFindingByCheckID(instanceFindings, checkID)
				if finding == nil {
					t.Errorf("Expected finding for check %s, but not found", checkID)
					continue
				}
				if finding.Status != expectedStatus {
					t.Errorf("Check %s: got status %s, want %s", checkID, finding.Status, expectedStatus)
				}
			}
		})
	}
}

// TestEC2Scanner_SecurityGroups tests security group checks
func TestEC2Scanner_SecurityGroups(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	if !IsLocalStackRunning(ctx) {
		t.Skip("LocalStack is not running")
	}

	cfg := NewDefaultConfig()
	awsCfg, err := cfg.GetAWSConfig(ctx)
	if err != nil {
		t.Fatalf("Failed to get AWS config: %v", err)
	}

	ec2Client, err := cfg.NewEC2Client(ctx)
	if err != nil {
		t.Fatalf("Failed to create EC2 client: %v", err)
	}

	// Create VPC
	vpcOutput, err := ec2Client.CreateVpc(ctx, &awsec2.CreateVpcInput{
		CidrBlock: aws.String("10.0.0.0/16"),
	})
	if err != nil {
		t.Fatalf("Failed to create VPC: %v", err)
	}
	vpcID := aws.ToString(vpcOutput.Vpc.VpcId)
	defer func() {
		_, _ = ec2Client.DeleteVpc(ctx, &awsec2.DeleteVpcInput{VpcId: aws.String(vpcID)})
	}()

	tests := []struct {
		name         string
		ingressRules []types.IpPermission
		expectFail   bool
	}{
		{
			name: "sg_with_unrestricted_ssh",
			ingressRules: []types.IpPermission{
				{
					IpProtocol: aws.String("tcp"),
					FromPort:   aws.Int32(22),
					ToPort:     aws.Int32(22),
					IpRanges: []types.IpRange{
						{CidrIp: aws.String("0.0.0.0/0")},
					},
				},
			},
			expectFail: true,
		},
		{
			name: "sg_with_restricted_ssh",
			ingressRules: []types.IpPermission{
				{
					IpProtocol: aws.String("tcp"),
					FromPort:   aws.Int32(22),
					ToPort:     aws.Int32(22),
					IpRanges: []types.IpRange{
						{CidrIp: aws.String("10.0.0.0/8")},
					},
				},
			},
			expectFail: false,
		},
		{
			name: "sg_with_unrestricted_rdp",
			ingressRules: []types.IpPermission{
				{
					IpProtocol: aws.String("tcp"),
					FromPort:   aws.Int32(3389),
					ToPort:     aws.Int32(3389),
					IpRanges: []types.IpRange{
						{CidrIp: aws.String("0.0.0.0/0")},
					},
				},
			},
			expectFail: true,
		},
		{
			name: "sg_with_all_traffic",
			ingressRules: []types.IpPermission{
				{
					IpProtocol: aws.String("-1"),
					IpRanges: []types.IpRange{
						{CidrIp: aws.String("0.0.0.0/0")},
					},
				},
			},
			expectFail: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create security group
			sgName := tt.name + "-" + time.Now().Format("150405")
			sgOutput, err := ec2Client.CreateSecurityGroup(ctx, &awsec2.CreateSecurityGroupInput{
				GroupName:   aws.String(sgName),
				Description: aws.String("Test SG: " + tt.name),
				VpcId:       aws.String(vpcID),
			})
			if err != nil {
				t.Fatalf("Failed to create security group: %v", err)
			}
			sgID := aws.ToString(sgOutput.GroupId)
			defer func() {
				_, _ = ec2Client.DeleteSecurityGroup(ctx, &awsec2.DeleteSecurityGroupInput{
					GroupId: aws.String(sgID),
				})
			}()

			// Add ingress rules
			if len(tt.ingressRules) > 0 {
				_, err = ec2Client.AuthorizeSecurityGroupIngress(ctx, &awsec2.AuthorizeSecurityGroupIngressInput{
					GroupId:       aws.String(sgID),
					IpPermissions: tt.ingressRules,
				})
				if err != nil {
					t.Fatalf("Failed to add ingress rules: %v", err)
				}
			}

			// Run scanner
			ec2Scanner := ec2.NewScanner(awsCfg, DefaultRegion, TestAccountID)
			findings, err := ec2Scanner.Scan(ctx, DefaultRegion)
			if err != nil {
				t.Fatalf("Scan failed: %v", err)
			}

			// Look for security group related findings
			sgFindings := filterFindingsByResource(findings, sgID)

			t.Logf("Found %d findings for security group %s", len(sgFindings), sgID)
			for _, f := range sgFindings {
				t.Logf("  %s: %s (%s)", f.CheckID, f.Status, f.Title)
			}

			// Check if we found the expected result
			hasFail := false
			for _, f := range sgFindings {
				if f.Status == scanner.StatusFail {
					hasFail = true
					break
				}
			}

			if tt.expectFail && !hasFail {
				t.Errorf("Expected FAIL finding for security group with risky rules")
			}
		})
	}
}

// TestEC2Scanner_ElasticIPs tests Elastic IP checks
func TestEC2Scanner_ElasticIPs(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	if !IsLocalStackRunning(ctx) {
		t.Skip("LocalStack is not running")
	}

	cfg := NewDefaultConfig()
	awsCfg, err := cfg.GetAWSConfig(ctx)
	if err != nil {
		t.Fatalf("Failed to get AWS config: %v", err)
	}

	ec2Client, err := cfg.NewEC2Client(ctx)
	if err != nil {
		t.Fatalf("Failed to create EC2 client: %v", err)
	}

	// Allocate an unassociated Elastic IP
	eipOutput, err := ec2Client.AllocateAddress(ctx, &awsec2.AllocateAddressInput{
		Domain: types.DomainTypeVpc,
	})
	if err != nil {
		t.Fatalf("Failed to allocate EIP: %v", err)
	}
	allocationID := aws.ToString(eipOutput.AllocationId)
	defer func() {
		_, _ = ec2Client.ReleaseAddress(ctx, &awsec2.ReleaseAddressInput{
			AllocationId: aws.String(allocationID),
		})
	}()

	t.Logf("Created unassociated EIP: %s", allocationID)

	// Run scanner
	ec2Scanner := ec2.NewScanner(awsCfg, DefaultRegion, TestAccountID)
	findings, err := ec2Scanner.Scan(ctx, DefaultRegion)
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	// Look for unassociated EIP finding
	found := false
	for _, f := range findings {
		if f.CheckID == "ec2_unassociated_eip" && f.Status == scanner.StatusFail {
			found = true
			t.Logf("Found unassociated EIP finding: %s", f.Description)
			break
		}
	}

	if !found {
		t.Logf("All findings:")
		for _, f := range findings {
			if f.Service == "ec2" {
				t.Logf("  %s: %s (%s)", f.CheckID, f.Status, f.Title)
			}
		}
		t.Errorf("Expected to find unassociated EIP check")
	}
}
