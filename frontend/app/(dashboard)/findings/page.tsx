"use client";

import { PageHeader } from "@/components/page-header";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { FindingsTable } from "./_components/findings-table";
import { FindingsFilters } from "./_components/findings-filters";
import {
  ShieldWarningIcon,
  WarningIcon,
  WarningCircleIcon,
  InfoIcon,
} from "@phosphor-icons/react";

// Mock stats - will be replaced with real API calls
const stats = {
  critical: 3,
  high: 12,
  medium: 18,
  low: 14,
};

export default function FindingsPage() {
  return (
    <>
      <PageHeader title="Findings" breadcrumbs={[{ label: "Findings" }]} />

      <div className="flex flex-1 flex-col gap-6 p-6">
        {/* Severity Summary Cards */}
        <div className="grid gap-4 md:grid-cols-4">
          <Card className="border-l-4 border-l-red-500">
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">Critical</CardTitle>
              <ShieldWarningIcon
                className="size-4 text-red-500"
                weight="fill"
              />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold text-red-500">
                {stats.critical}
              </div>
              <p className="text-xs text-muted-foreground">
                Immediate action required
              </p>
            </CardContent>
          </Card>

          <Card className="border-l-4 border-l-orange-500">
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">High</CardTitle>
              <WarningIcon className="size-4 text-orange-500" weight="fill" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold text-orange-500">
                {stats.high}
              </div>
              <p className="text-xs text-muted-foreground">
                Should be addressed soon
              </p>
            </CardContent>
          </Card>

          <Card className="border-l-4 border-l-yellow-500">
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">Medium</CardTitle>
              <WarningCircleIcon
                className="size-4 text-yellow-500"
                weight="fill"
              />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold text-yellow-500">
                {stats.medium}
              </div>
              <p className="text-xs text-muted-foreground">Plan to remediate</p>
            </CardContent>
          </Card>

          <Card className="border-l-4 border-l-green-500">
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">Low</CardTitle>
              <InfoIcon className="size-4 text-green-500" weight="fill" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold text-green-500">
                {stats.low}
              </div>
              <p className="text-xs text-muted-foreground">
                Best practice improvements
              </p>
            </CardContent>
          </Card>
        </div>

        {/* Findings Table */}
        <Card>
          <CardHeader>
            <div className="flex items-center justify-between">
              <div>
                <CardTitle>Security Findings</CardTitle>
                <CardDescription>
                  All security issues discovered during scans
                </CardDescription>
              </div>
              <Badge variant="secondary">
                {stats.critical + stats.high + stats.medium + stats.low} Total
              </Badge>
            </div>
          </CardHeader>
          <CardContent>
            <FindingsFilters />
            <FindingsTable />
          </CardContent>
        </Card>
      </div>
    </>
  );
}
