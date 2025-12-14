// Package ec2 provides EC2 security scanning capabilities.
package ec2

import (
	"context"
	"fmt"

	"cloudcop/api/internal/scanner"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

const ipv4Any = "0.0.0.0/0"

var dangerousPorts = map[int32]string{
	22:    "SSH",
	3389:  "RDP",
	3306:  "MySQL",
	5432:  "PostgreSQL",
	1433:  "MSSQL",
	27017: "MongoDB",
	6379:  "Redis",
}

func (e *Scanner) checkPublicIP(_ context.Context, instance types.Instance) []scanner.Finding {
	instanceID := aws.ToString(instance.InstanceId)
	if instance.PublicIpAddress != nil {
		return []scanner.Finding{e.createFinding(
			"ec2_public_ip",
			instanceID,
			"EC2 instance has public IP address",
			fmt.Sprintf("Instance %s has public IP %s", instanceID, aws.ToString(instance.PublicIpAddress)),
			scanner.StatusFail,
			scanner.SeverityMedium,
		)}
	}
	return []scanner.Finding{e.createFinding(
		"ec2_public_ip",
		instanceID,
		"EC2 instance has no public IP",
		fmt.Sprintf("Instance %s has no public IP address", instanceID),
		scanner.StatusPass,
		scanner.SeverityMedium,
	)}
}

func (e *Scanner) checkEBSEncryption(ctx context.Context, instance types.Instance) []scanner.Finding {
	instanceID := aws.ToString(instance.InstanceId)
	var findings []scanner.Finding

	for _, bdm := range instance.BlockDeviceMappings {
		if bdm.Ebs == nil {
			continue
		}
		volumeID := aws.ToString(bdm.Ebs.VolumeId)
		volumes, err := e.client.DescribeVolumes(ctx, &ec2.DescribeVolumesInput{
			VolumeIds: []string{volumeID},
		})
		if err != nil {
			continue
		}
		for _, vol := range volumes.Volumes {
			if !aws.ToBool(vol.Encrypted) {
				findings = append(findings, e.createFinding(
					"ec2_ebs_encryption",
					volumeID,
					"EBS volume is not encrypted",
					fmt.Sprintf("Volume %s attached to %s is not encrypted", volumeID, instanceID),
					scanner.StatusFail,
					scanner.SeverityMedium,
				))
			} else {
				findings = append(findings, e.createFinding(
					"ec2_ebs_encryption",
					volumeID,
					"EBS volume is encrypted",
					fmt.Sprintf("Volume %s attached to %s is encrypted", volumeID, instanceID),
					scanner.StatusPass,
					scanner.SeverityMedium,
				))
			}
		}
	}
	return findings
}

func (e *Scanner) checkSecurityGroups(ctx context.Context, instance types.Instance) []scanner.Finding {
	instanceID := aws.ToString(instance.InstanceId)
	var findings []scanner.Finding

	for _, sg := range instance.SecurityGroups {
		sgID := aws.ToString(sg.GroupId)
		sgDetails, err := e.client.DescribeSecurityGroups(ctx, &ec2.DescribeSecurityGroupsInput{
			GroupIds: []string{sgID},
		})
		if err != nil {
			continue
		}
		for _, group := range sgDetails.SecurityGroups {
			for _, perm := range group.IpPermissions {
				for _, ipRange := range perm.IpRanges {
					if aws.ToString(ipRange.CidrIp) == ipv4Any {
						port := aws.ToInt32(perm.FromPort)
						findings = append(findings, e.createFinding(
							"ec2_sg_unrestricted_ingress",
							sgID,
							"Security group allows unrestricted ingress",
							fmt.Sprintf("SG %s on instance %s allows 0.0.0.0/0 on port %d", sgID, instanceID, port),
							scanner.StatusFail,
							scanner.SeverityHigh,
						))
					}
				}
			}
		}
	}
	return findings
}

func (e *Scanner) checkIMDSv2(_ context.Context, instance types.Instance) []scanner.Finding {
	instanceID := aws.ToString(instance.InstanceId)
	if instance.MetadataOptions != nil && instance.MetadataOptions.HttpTokens == types.HttpTokensStateRequired {
		return []scanner.Finding{e.createFinding(
			"ec2_imdsv2_required",
			instanceID,
			"EC2 instance requires IMDSv2",
			fmt.Sprintf("Instance %s enforces IMDSv2", instanceID),
			scanner.StatusPass,
			scanner.SeverityHigh,
		)}
	}
	return []scanner.Finding{e.createFinding(
		"ec2_imdsv2_required",
		instanceID,
		"EC2 instance does not require IMDSv2",
		fmt.Sprintf("Instance %s allows IMDSv1 (vulnerable to SSRF)", instanceID),
		scanner.StatusFail,
		scanner.SeverityHigh,
	)}
}

func (e *Scanner) checkIAMRole(_ context.Context, instance types.Instance) []scanner.Finding {
	instanceID := aws.ToString(instance.InstanceId)
	if instance.IamInstanceProfile != nil {
		return []scanner.Finding{e.createFinding(
			"ec2_iam_role",
			instanceID,
			"EC2 instance has IAM role attached",
			fmt.Sprintf("Instance %s has IAM instance profile", instanceID),
			scanner.StatusPass,
			scanner.SeverityMedium,
		)}
	}
	return []scanner.Finding{e.createFinding(
		"ec2_iam_role",
		instanceID,
		"EC2 instance has no IAM role",
		fmt.Sprintf("Instance %s has no IAM instance profile attached", instanceID),
		scanner.StatusFail,
		scanner.SeverityMedium,
	)}
}

func (e *Scanner) checkDetailedMonitoring(_ context.Context, instance types.Instance) []scanner.Finding {
	instanceID := aws.ToString(instance.InstanceId)
	if instance.Monitoring != nil && instance.Monitoring.State == types.MonitoringStateEnabled {
		return []scanner.Finding{e.createFinding(
			"ec2_detailed_monitoring",
			instanceID,
			"EC2 detailed monitoring is enabled",
			fmt.Sprintf("Instance %s has detailed monitoring (1-minute interval)", instanceID),
			scanner.StatusPass,
			scanner.SeverityLow,
		)}
	}
	return []scanner.Finding{e.createFinding(
		"ec2_detailed_monitoring",
		instanceID,
		"EC2 detailed monitoring is disabled",
		fmt.Sprintf("Instance %s uses basic monitoring (5-minute interval)", instanceID),
		scanner.StatusFail,
		scanner.SeverityLow,
	)}
}
func (e *Scanner) checkUnassociatedElasticIPs(ctx context.Context) []scanner.Finding {
	var findings []scanner.Finding
	addresses, err := e.client.DescribeAddresses(ctx, &ec2.DescribeAddressesInput{})
	if err != nil {
		return nil
	}
	for _, addr := range addresses.Addresses {
		allocID := aws.ToString(addr.AllocationId)
		if addr.AssociationId == nil {
			findings = append(findings, e.createFinding(
				"ec2_unassociated_eip",
				allocID,
				"Elastic IP is not associated",
				fmt.Sprintf("EIP %s is allocated but not associated with any resource", aws.ToString(addr.PublicIp)),
				scanner.StatusFail,
				scanner.SeverityLow,
			))
		}
	}
	return findings
}

func (e *Scanner) checkUnrestrictedSecurityGroups(ctx context.Context) []scanner.Finding {
	var findings []scanner.Finding
	sgs, err := e.client.DescribeSecurityGroups(ctx, &ec2.DescribeSecurityGroupsInput{})
	if err != nil {
		return nil
	}
	for _, sg := range sgs.SecurityGroups {
		sgID := aws.ToString(sg.GroupId)
		for _, perm := range sg.IpPermissions {
			for _, ipRange := range perm.IpRanges {
				if aws.ToString(ipRange.CidrIp) == ipv4Any {
					findings = append(findings, e.createFinding(
						"ec2_sg_unrestricted_ingress",
						sgID,
						"Security group allows unrestricted ingress from 0.0.0.0/0",
						fmt.Sprintf("SG %s allows traffic from any IP on port %d", sgID, aws.ToInt32(perm.FromPort)),
						scanner.StatusFail,
						scanner.SeverityHigh,
					))
				}
			}
		}
	}
	return findings
}

func (e *Scanner) checkDangerousPorts(ctx context.Context) []scanner.Finding {
	var findings []scanner.Finding
	sgs, err := e.client.DescribeSecurityGroups(ctx, &ec2.DescribeSecurityGroupsInput{})
	if err != nil {
		return nil
	}
	for _, sg := range sgs.SecurityGroups {
		sgID := aws.ToString(sg.GroupId)
		for _, perm := range sg.IpPermissions {
			for _, ipRange := range perm.IpRanges {
				if aws.ToString(ipRange.CidrIp) == ipv4Any {
					port := aws.ToInt32(perm.FromPort)
					if serviceName, dangerous := dangerousPorts[port]; dangerous {
						findings = append(findings, e.createFinding(
							"ec2_sg_dangerous_ports",
							sgID,
							fmt.Sprintf("Security group exposes %s port to internet", serviceName),
							fmt.Sprintf("SG %s allows 0.0.0.0/0 access to %s (port %d)", sgID, serviceName, port),
							scanner.StatusFail,
							scanner.SeverityCritical,
						))
					}
				}
			}
		}
	}
	return findings
}
