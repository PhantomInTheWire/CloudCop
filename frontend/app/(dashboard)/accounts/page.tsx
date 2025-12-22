"use client";

import { PageHeader } from "@/components/page-header";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { AccountsList } from "./_components/accounts-list";
import { AddAccountDialog } from "./_components/add-account-dialog";

export default function AccountsPage() {
  return (
    <>
      <PageHeader
        title="AWS Accounts"
        breadcrumbs={[{ label: "AWS Accounts" }]}
        actions={<AddAccountDialog />}
      />

      <div className="flex flex-1 flex-col gap-6 p-6">
        <Card>
          <CardHeader>
            <CardTitle>Connected AWS Accounts</CardTitle>
            <CardDescription>
              Manage your connected AWS accounts for security scanning
            </CardDescription>
          </CardHeader>
          <CardContent>
            <AccountsList />
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>How it Works</CardTitle>
            <CardDescription>
              CloudCop uses AWS IAM roles for secure, read-only access to your
              accounts
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div className="grid gap-6 md:grid-cols-3">
              <div className="space-y-2">
                <div className="flex items-center gap-2">
                  <div className="size-8 rounded-full bg-primary/10 flex items-center justify-center text-primary font-semibold">
                    1
                  </div>
                  <h4 className="font-medium">Deploy CloudFormation</h4>
                </div>
                <p className="text-sm text-muted-foreground pl-10">
                  Deploy our CloudFormation stack that creates a read-only IAM
                  role with security scanning permissions.
                </p>
              </div>
              <div className="space-y-2">
                <div className="flex items-center gap-2">
                  <div className="size-8 rounded-full bg-primary/10 flex items-center justify-center text-primary font-semibold">
                    2
                  </div>
                  <h4 className="font-medium">Add Account Details</h4>
                </div>
                <p className="text-sm text-muted-foreground pl-10">
                  Enter your AWS Account ID and the External ID from the
                  CloudFormation outputs.
                </p>
              </div>
              <div className="space-y-2">
                <div className="flex items-center gap-2">
                  <div className="size-8 rounded-full bg-primary/10 flex items-center justify-center text-primary font-semibold">
                    3
                  </div>
                  <h4 className="font-medium">Verify & Scan</h4>
                </div>
                <p className="text-sm text-muted-foreground pl-10">
                  We verify the connection using STS AssumeRole and you&apos;re
                  ready to run security scans.
                </p>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>
    </>
  );
}
