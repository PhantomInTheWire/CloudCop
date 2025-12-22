"use client";

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
import { Progress } from "@/components/ui/progress";
import {
  CheckCircleIcon,
  CircleNotchIcon,
  XCircleIcon,
  ClockIcon,
  ArrowRightIcon,
  EyeIcon,
} from "@phosphor-icons/react";
import Link from "next/link";

// Mock data - will be replaced with real API calls
const scans = [
  {
    id: "scan-001",
    status: "completed",
    services: ["S3", "EC2", "IAM", "Lambda"],
    regions: ["us-east-1", "us-west-2"],
    score: 72,
    findings: 47,
    startedAt: "2024-12-22T10:30:00Z",
    completedAt: "2024-12-22T10:35:00Z",
    duration: "5m 23s",
  },
  {
    id: "scan-002",
    status: "completed",
    services: ["S3", "EC2"],
    regions: ["us-east-1"],
    score: 85,
    findings: 12,
    startedAt: "2024-12-21T15:00:00Z",
    completedAt: "2024-12-21T15:03:00Z",
    duration: "3m 12s",
  },
  {
    id: "scan-003",
    status: "running",
    services: ["S3", "EC2", "IAM", "Lambda", "ECS", "DynamoDB"],
    regions: ["us-east-1", "us-west-2", "eu-west-1"],
    score: null,
    findings: null,
    startedAt: "2024-12-22T12:00:00Z",
    completedAt: null,
    duration: null,
    progress: 65,
  },
  {
    id: "scan-004",
    status: "failed",
    services: ["S3"],
    regions: ["us-east-1"],
    score: null,
    findings: null,
    startedAt: "2024-12-20T09:00:00Z",
    completedAt: "2024-12-20T09:01:00Z",
    duration: "1m 05s",
    error: "AWS credentials expired",
  },
  {
    id: "scan-005",
    status: "completed",
    services: ["IAM"],
    regions: ["global"],
    score: 68,
    findings: 23,
    startedAt: "2024-12-19T14:30:00Z",
    completedAt: "2024-12-19T14:32:00Z",
    duration: "2m 45s",
  },
];

const statusConfig = {
  completed: {
    icon: CheckCircleIcon,
    label: "Completed",
    variant: "default" as const,
    className: "text-green-500",
  },
  running: {
    icon: CircleNotchIcon,
    label: "Running",
    variant: "secondary" as const,
    className: "text-blue-500 animate-spin",
  },
  failed: {
    icon: XCircleIcon,
    label: "Failed",
    variant: "destructive" as const,
    className: "text-destructive",
  },
  pending: {
    icon: ClockIcon,
    label: "Pending",
    variant: "outline" as const,
    className: "text-muted-foreground",
  },
};

function formatDate(dateString: string) {
  return new Date(dateString).toLocaleString("en-US", {
    month: "short",
    day: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  });
}

export function ScanHistoryTable() {
  return (
    <Table>
      <TableHeader>
        <TableRow>
          <TableHead>Status</TableHead>
          <TableHead>Services</TableHead>
          <TableHead>Regions</TableHead>
          <TableHead>Score</TableHead>
          <TableHead>Findings</TableHead>
          <TableHead>Started</TableHead>
          <TableHead>Duration</TableHead>
          <TableHead className="text-right">Actions</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        {scans.map((scan) => {
          const config = statusConfig[scan.status as keyof typeof statusConfig];
          const StatusIcon = config.icon;

          return (
            <TableRow key={scan.id}>
              <TableCell>
                <div className="flex items-center gap-2">
                  <StatusIcon
                    className={`size-4 ${config.className}`}
                    weight="fill"
                  />
                  <Badge variant={config.variant}>{config.label}</Badge>
                </div>
              </TableCell>
              <TableCell>
                <div className="flex flex-wrap gap-1">
                  {scan.services.slice(0, 3).map((service) => (
                    <Badge key={service} variant="outline" className="text-xs">
                      {service}
                    </Badge>
                  ))}
                  {scan.services.length > 3 && (
                    <Badge variant="outline" className="text-xs">
                      +{scan.services.length - 3}
                    </Badge>
                  )}
                </div>
              </TableCell>
              <TableCell>
                <span className="text-sm text-muted-foreground">
                  {scan.regions.length} region
                  {scan.regions.length !== 1 ? "s" : ""}
                </span>
              </TableCell>
              <TableCell>
                {scan.status === "running" ? (
                  <div className="w-20">
                    <Progress value={scan.progress} className="h-2" />
                    <span className="text-xs text-muted-foreground">
                      {scan.progress}%
                    </span>
                  </div>
                ) : scan.score !== null ? (
                  <span
                    className={`font-medium ${
                      scan.score >= 80
                        ? "text-green-500"
                        : scan.score >= 60
                          ? "text-yellow-500"
                          : "text-destructive"
                    }`}
                  >
                    {scan.score}%
                  </span>
                ) : (
                  <span className="text-muted-foreground">-</span>
                )}
              </TableCell>
              <TableCell>
                {scan.findings !== null ? (
                  <span className="font-medium">{scan.findings}</span>
                ) : (
                  <span className="text-muted-foreground">-</span>
                )}
              </TableCell>
              <TableCell>
                <span className="text-sm">{formatDate(scan.startedAt)}</span>
              </TableCell>
              <TableCell>
                <span className="text-sm text-muted-foreground">
                  {scan.duration || "-"}
                </span>
              </TableCell>
              <TableCell className="text-right">
                {scan.status === "completed" && (
                  <Button variant="ghost" size="sm" asChild>
                    <Link href={`/findings?scanId=${scan.id}`}>
                      <EyeIcon className="mr-2 size-4" />
                      View
                    </Link>
                  </Button>
                )}
              </TableCell>
            </TableRow>
          );
        })}
      </TableBody>
    </Table>
  );
}
