"use client";

import { Badge } from "@/components/ui/badge";
import { ScrollArea } from "@/components/ui/scroll-area";
import {
  ShieldWarningIcon,
  WarningCircleIcon,
  InfoIcon,
  WarningIcon,
} from "@phosphor-icons/react";

// Mock data - will be replaced with real API calls
const recentFindings = [
  {
    id: "1",
    title: "S3 bucket allows public access",
    service: "S3",
    severity: "CRITICAL",
    resourceId: "prod-data-bucket",
    time: "2 hours ago",
  },
  {
    id: "2",
    title: "EC2 instance has IMDSv1 enabled",
    service: "EC2",
    severity: "HIGH",
    resourceId: "i-0abc123def456",
    time: "2 hours ago",
  },
  {
    id: "3",
    title: "IAM user has inline policies",
    service: "IAM",
    severity: "MEDIUM",
    resourceId: "admin-user",
    time: "2 hours ago",
  },
  {
    id: "4",
    title: "Lambda function not in VPC",
    service: "Lambda",
    severity: "LOW",
    resourceId: "data-processor",
    time: "2 hours ago",
  },
  {
    id: "5",
    title: "Security group allows unrestricted SSH",
    service: "EC2",
    severity: "CRITICAL",
    resourceId: "sg-0123456789",
    time: "2 hours ago",
  },
];

const severityConfig = {
  CRITICAL: {
    icon: ShieldWarningIcon,
    variant: "destructive" as const,
    className: "text-red-500",
  },
  HIGH: {
    icon: WarningIcon,
    variant: "default" as const,
    className: "text-orange-500 bg-orange-500/10",
  },
  MEDIUM: {
    icon: WarningCircleIcon,
    variant: "secondary" as const,
    className: "text-yellow-500 bg-yellow-500/10",
  },
  LOW: {
    icon: InfoIcon,
    variant: "outline" as const,
    className: "text-green-500 bg-green-500/10",
  },
};

export function RecentFindings() {
  return (
    <ScrollArea className="h-[280px]">
      <div className="space-y-4">
        {recentFindings.map((finding) => {
          const config =
            severityConfig[finding.severity as keyof typeof severityConfig];
          const Icon = config.icon;

          return (
            <div
              key={finding.id}
              className="flex items-start gap-4 rounded-lg border p-3 transition-colors hover:bg-muted/50"
            >
              <div className={`mt-0.5 ${config.className}`}>
                <Icon className="size-5" weight="fill" />
              </div>
              <div className="flex-1 space-y-1">
                <div className="flex items-center justify-between gap-2">
                  <p className="text-sm font-medium leading-none">
                    {finding.title}
                  </p>
                  <Badge variant={config.variant} className="text-xs">
                    {finding.severity}
                  </Badge>
                </div>
                <div className="flex items-center gap-2 text-xs text-muted-foreground">
                  <Badge variant="outline" className="text-xs">
                    {finding.service}
                  </Badge>
                  <span className="truncate">{finding.resourceId}</span>
                  <span className="ml-auto">{finding.time}</span>
                </div>
              </div>
            </div>
          );
        })}
      </div>
    </ScrollArea>
  );
}
