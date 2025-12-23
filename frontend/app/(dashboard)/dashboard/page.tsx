"use client";

import { PageHeader } from "@/components/page-header";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Progress } from "@/components/ui/progress";
import {
  ShieldWarningIcon,
  ShieldCheckIcon,
  WarningCircleIcon,
  ArrowRightIcon,
  PlayIcon,
  CloudIcon,
} from "@phosphor-icons/react";
import Link from "next/link";
import { SecurityScoreChart } from "./_components/security-score-chart";
import { FindingsBySeverityChart } from "./_components/findings-by-severity-chart";
import { RecentFindings } from "./_components/recent-findings";
import { QuickActions } from "./_components/quick-actions";

// Mock data - will be replaced with real API calls
const stats = {
  totalFindings: 47,
  criticalFindings: 3,
  highFindings: 12,
  mediumFindings: 18,
  lowFindings: 14,
  securityScore: 72,
  scannedAccounts: 2,
  lastScanTime: "2 hours ago",
};

export default function DashboardPage() {
  return (
    <>
      <PageHeader
        title="Dashboard"
        actions={
          <Button asChild>
            <Link href="/scans">
              <PlayIcon className="mr-2 size-4" />
              New Scan
            </Link>
          </Button>
        }
      />

      <div className="flex flex-1 flex-col gap-6 p-6">
        {/* Stats Overview */}
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">
                Security Score
              </CardTitle>
              <ShieldCheckIcon className="size-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{stats.securityScore}%</div>
              <Progress value={stats.securityScore} className="mt-2" />
              <p className="text-xs text-muted-foreground mt-2">
                +5% from last scan
              </p>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">
                Critical Issues
              </CardTitle>
              <ShieldWarningIcon className="size-4 text-destructive" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold text-destructive">
                {stats.criticalFindings}
              </div>
              <p className="text-xs text-muted-foreground mt-2">
                Require immediate attention
              </p>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">
                Total Findings
              </CardTitle>
              <WarningCircleIcon className="size-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{stats.totalFindings}</div>
              <p className="text-xs text-muted-foreground mt-2">
                Across {stats.scannedAccounts} AWS accounts
              </p>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">Last Scan</CardTitle>
              <CloudIcon className="size-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{stats.lastScanTime}</div>
              <p className="text-xs text-muted-foreground mt-2">
                All services scanned
              </p>
            </CardContent>
          </Card>
        </div>

        {/* Charts Row */}
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-7">
          <Card className="lg:col-span-3">
            <CardHeader>
              <CardTitle>Security Score</CardTitle>
              <CardDescription>
                Overall security posture based on findings
              </CardDescription>
            </CardHeader>
            <CardContent>
              <SecurityScoreChart score={stats.securityScore} />
            </CardContent>
          </Card>

          <Card className="lg:col-span-4">
            <CardHeader>
              <CardTitle>Findings by Severity</CardTitle>
              <CardDescription>
                Distribution of security findings by severity level
              </CardDescription>
            </CardHeader>
            <CardContent>
              <FindingsBySeverityChart
                critical={stats.criticalFindings}
                high={stats.highFindings}
                medium={stats.mediumFindings}
                low={stats.lowFindings}
              />
            </CardContent>
          </Card>
        </div>

        {/* Bottom Row */}
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-7">
          <Card className="lg:col-span-4">
            <CardHeader className="flex flex-row items-center justify-between">
              <div>
                <CardTitle>Recent Findings</CardTitle>
                <CardDescription>
                  Latest security issues discovered
                </CardDescription>
              </div>
              <Button variant="ghost" size="sm" asChild>
                <Link href="/findings">
                  View all
                  <ArrowRightIcon className="ml-2 size-4" />
                </Link>
              </Button>
            </CardHeader>
            <CardContent>
              <RecentFindings />
            </CardContent>
          </Card>

          <Card className="lg:col-span-3">
            <CardHeader>
              <CardTitle>Quick Actions</CardTitle>
              <CardDescription>
                Common tasks and recommendations
              </CardDescription>
            </CardHeader>
            <CardContent>
              <QuickActions />
            </CardContent>
          </Card>
        </div>
      </div>
    </>
  );
}
