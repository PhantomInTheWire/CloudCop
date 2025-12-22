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
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { PlayIcon } from "@phosphor-icons/react";
import { ScanHistoryTable } from "./_components/scan-history-table";
import { NewScanForm } from "./_components/new-scan-form";

export default function ScansPage() {
  return (
    <>
      <PageHeader title="Scans" breadcrumbs={[{ label: "Scans" }]} />

      <div className="flex flex-1 flex-col gap-6 p-6">
        <Tabs defaultValue="history" className="space-y-6">
          <TabsList>
            <TabsTrigger value="history">Scan History</TabsTrigger>
            <TabsTrigger value="new">New Scan</TabsTrigger>
          </TabsList>

          <TabsContent value="history" className="space-y-4">
            <Card>
              <CardHeader>
                <CardTitle>Recent Scans</CardTitle>
                <CardDescription>
                  View all security scans and their results
                </CardDescription>
              </CardHeader>
              <CardContent>
                <ScanHistoryTable />
              </CardContent>
            </Card>
          </TabsContent>

          <TabsContent value="new" className="space-y-4">
            <Card>
              <CardHeader>
                <CardTitle>Start New Scan</CardTitle>
                <CardDescription>
                  Configure and run a security scan on your AWS environment
                </CardDescription>
              </CardHeader>
              <CardContent>
                <NewScanForm />
              </CardContent>
            </Card>
          </TabsContent>
        </Tabs>
      </div>
    </>
  );
}
