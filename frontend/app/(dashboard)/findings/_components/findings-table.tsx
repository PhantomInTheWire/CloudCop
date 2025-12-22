"use client";

import * as React from "react";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { ScrollArea } from "@/components/ui/scroll-area";
import {
  ShieldWarningIcon,
  WarningIcon,
  WarningCircleIcon,
  InfoIcon,
  CopyIcon,
  CheckIcon,
  CodeIcon,
  ArrowRightIcon,
} from "@phosphor-icons/react";

// Mock data - will be replaced with real API calls
const findings = [
  {
    id: "f-001",
    service: "S3",
    region: "us-east-1",
    resourceId: "prod-data-bucket",
    checkId: "s3-public-access",
    severity: "CRITICAL",
    title: "S3 bucket allows public access",
    description:
      "The S3 bucket has public access enabled, which could expose sensitive data to the internet. This is a critical security risk that should be addressed immediately.",
    compliance: ["CIS", "SOC2", "PCI-DSS"],
    remediation: `aws s3api put-public-access-block \\
  --bucket prod-data-bucket \\
  --public-access-block-configuration "BlockPublicAcls=true,IgnorePublicAcls=true,BlockPublicPolicy=true,RestrictPublicBuckets=true"`,
  },
  {
    id: "f-002",
    service: "EC2",
    region: "us-east-1",
    resourceId: "sg-0123456789abcdef0",
    checkId: "ec2-unrestricted-ssh",
    severity: "CRITICAL",
    title: "Security group allows unrestricted SSH",
    description:
      "The security group allows SSH access from 0.0.0.0/0, which exposes the instances to brute force attacks from anywhere on the internet.",
    compliance: ["CIS", "NIST"],
    remediation: `aws ec2 revoke-security-group-ingress \\
  --group-id sg-0123456789abcdef0 \\
  --protocol tcp --port 22 --cidr 0.0.0.0/0

aws ec2 authorize-security-group-ingress \\
  --group-id sg-0123456789abcdef0 \\
  --protocol tcp --port 22 --cidr 10.0.0.0/8`,
  },
  {
    id: "f-003",
    service: "EC2",
    region: "us-west-2",
    resourceId: "i-0abc123def456789a",
    checkId: "ec2-imdsv1",
    severity: "HIGH",
    title: "EC2 instance has IMDSv1 enabled",
    description:
      "Instance Metadata Service Version 1 (IMDSv1) is enabled, which is vulnerable to SSRF attacks. IMDSv2 should be enforced.",
    compliance: ["CIS", "SOC2"],
    remediation: `aws ec2 modify-instance-metadata-options \\
  --instance-id i-0abc123def456789a \\
  --http-tokens required \\
  --http-endpoint enabled`,
  },
  {
    id: "f-004",
    service: "IAM",
    region: "global",
    resourceId: "admin-user",
    checkId: "iam-inline-policy",
    severity: "MEDIUM",
    title: "IAM user has inline policies",
    description:
      "The IAM user has inline policies attached instead of managed policies. This makes it harder to audit and manage permissions.",
    compliance: ["CIS"],
    remediation: `# Create a managed policy from the inline policy
aws iam create-policy \\
  --policy-name admin-user-policy \\
  --policy-document file://policy.json

# Attach the managed policy
aws iam attach-user-policy \\
  --user-name admin-user \\
  --policy-arn arn:aws:iam::ACCOUNT_ID:policy/admin-user-policy

# Delete the inline policy
aws iam delete-user-policy \\
  --user-name admin-user \\
  --policy-name InlinePolicy`,
  },
  {
    id: "f-005",
    service: "Lambda",
    region: "us-east-1",
    resourceId: "data-processor",
    checkId: "lambda-vpc",
    severity: "LOW",
    title: "Lambda function not in VPC",
    description:
      "The Lambda function is not configured to run inside a VPC. While not always required, VPC integration provides additional network isolation.",
    compliance: ["SOC2"],
    remediation: `aws lambda update-function-configuration \\
  --function-name data-processor \\
  --vpc-config SubnetIds=subnet-xxx,SecurityGroupIds=sg-xxx`,
  },
  {
    id: "f-006",
    service: "S3",
    region: "us-east-1",
    resourceId: "logs-bucket",
    checkId: "s3-versioning",
    severity: "MEDIUM",
    title: "S3 bucket versioning not enabled",
    description:
      "Object versioning is not enabled on this bucket. Enabling versioning helps protect against accidental deletions and overwrites.",
    compliance: ["CIS", "SOC2"],
    remediation: `aws s3api put-bucket-versioning \\
  --bucket logs-bucket \\
  --versioning-configuration Status=Enabled`,
  },
];

const severityConfig = {
  CRITICAL: {
    icon: ShieldWarningIcon,
    variant: "destructive" as const,
    className: "text-red-500",
    bgClass: "bg-red-500/10",
  },
  HIGH: {
    icon: WarningIcon,
    variant: "default" as const,
    className: "text-orange-500",
    bgClass: "bg-orange-500/10",
  },
  MEDIUM: {
    icon: WarningCircleIcon,
    variant: "secondary" as const,
    className: "text-yellow-500",
    bgClass: "bg-yellow-500/10",
  },
  LOW: {
    icon: InfoIcon,
    variant: "outline" as const,
    className: "text-green-500",
    bgClass: "bg-green-500/10",
  },
};

function FindingDetailDialog({ finding }: { finding: (typeof findings)[0] }) {
  const [copied, setCopied] = React.useState(false);
  const config =
    severityConfig[finding.severity as keyof typeof severityConfig];
  const SeverityIcon = config.icon;

  const copyRemediation = () => {
    navigator.clipboard.writeText(finding.remediation);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  return (
    <DialogContent className="max-w-2xl">
      <DialogHeader>
        <div className="flex items-start gap-3">
          <div className={`p-2 rounded-lg ${config.bgClass}`}>
            <SeverityIcon
              className={`size-5 ${config.className}`}
              weight="fill"
            />
          </div>
          <div className="flex-1">
            <DialogTitle className="text-lg">{finding.title}</DialogTitle>
            <DialogDescription className="mt-1">
              {finding.service} / {finding.region} / {finding.resourceId}
            </DialogDescription>
          </div>
          <Badge variant={config.variant}>{finding.severity}</Badge>
        </div>
      </DialogHeader>

      <div className="space-y-6 mt-4">
        <div>
          <h4 className="text-sm font-medium mb-2">Description</h4>
          <p className="text-sm text-muted-foreground">{finding.description}</p>
        </div>

        <div>
          <h4 className="text-sm font-medium mb-2">Compliance Frameworks</h4>
          <div className="flex flex-wrap gap-2">
            {finding.compliance.map((framework) => (
              <Badge key={framework} variant="outline">
                {framework}
              </Badge>
            ))}
          </div>
        </div>

        <div>
          <div className="flex items-center justify-between mb-2">
            <h4 className="text-sm font-medium flex items-center gap-2">
              <CodeIcon className="size-4" />
              Remediation Commands
            </h4>
            <Button variant="ghost" size="sm" onClick={copyRemediation}>
              {copied ? (
                <>
                  <CheckIcon className="mr-2 size-4" />
                  Copied!
                </>
              ) : (
                <>
                  <CopyIcon className="mr-2 size-4" />
                  Copy
                </>
              )}
            </Button>
          </div>
          <ScrollArea className="h-[200px] w-full rounded-lg border bg-muted/50 p-4">
            <pre className="text-xs font-mono whitespace-pre-wrap">
              {finding.remediation}
            </pre>
          </ScrollArea>
        </div>
      </div>
    </DialogContent>
  );
}

export function FindingsTable() {
  return (
    <Table>
      <TableHeader>
        <TableRow>
          <TableHead>Severity</TableHead>
          <TableHead>Title</TableHead>
          <TableHead>Service</TableHead>
          <TableHead>Resource</TableHead>
          <TableHead>Region</TableHead>
          <TableHead>Compliance</TableHead>
          <TableHead className="text-right">Actions</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        {findings.map((finding) => {
          const config =
            severityConfig[finding.severity as keyof typeof severityConfig];
          const SeverityIcon = config.icon;

          return (
            <TableRow key={finding.id}>
              <TableCell>
                <div className="flex items-center gap-2">
                  <SeverityIcon
                    className={`size-4 ${config.className}`}
                    weight="fill"
                  />
                  <Badge variant={config.variant}>{finding.severity}</Badge>
                </div>
              </TableCell>
              <TableCell>
                <span className="font-medium">{finding.title}</span>
              </TableCell>
              <TableCell>
                <Badge variant="outline">{finding.service}</Badge>
              </TableCell>
              <TableCell>
                <code className="text-xs bg-muted px-1.5 py-0.5 rounded">
                  {finding.resourceId}
                </code>
              </TableCell>
              <TableCell>
                <span className="text-sm text-muted-foreground">
                  {finding.region}
                </span>
              </TableCell>
              <TableCell>
                <div className="flex gap-1">
                  {finding.compliance.slice(0, 2).map((framework) => (
                    <Badge
                      key={framework}
                      variant="secondary"
                      className="text-xs"
                    >
                      {framework}
                    </Badge>
                  ))}
                  {finding.compliance.length > 2 && (
                    <Badge variant="secondary" className="text-xs">
                      +{finding.compliance.length - 2}
                    </Badge>
                  )}
                </div>
              </TableCell>
              <TableCell className="text-right">
                <Dialog>
                  <DialogTrigger asChild>
                    <Button variant="ghost" size="sm">
                      View Details
                      <ArrowRightIcon className="ml-2 size-4" />
                    </Button>
                  </DialogTrigger>
                  <FindingDetailDialog finding={finding} />
                </Dialog>
              </TableCell>
            </TableRow>
          );
        })}
      </TableBody>
    </Table>
  );
}
