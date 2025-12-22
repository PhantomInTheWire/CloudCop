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
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import {
  CheckCircleIcon,
  WarningCircleIcon,
  DotsThreeIcon,
  PlayIcon,
  TrashIcon,
  ArrowClockwiseIcon,
  CloudIcon,
} from "@phosphor-icons/react";
import Link from "next/link";

// Mock data - will be replaced with real API calls
const accounts = [
  {
    id: "1",
    accountId: "123456789012",
    name: "Production",
    roleArn: "arn:aws:iam::123456789012:role/CloudCopSecurityScanRole",
    externalId: "cc-ext-abc123",
    verified: true,
    lastVerifiedAt: "2024-12-22T10:00:00Z",
    lastScanAt: "2024-12-22T10:30:00Z",
    findingsCount: 32,
  },
  {
    id: "2",
    accountId: "987654321098",
    name: "Development",
    roleArn: "arn:aws:iam::987654321098:role/CloudCopSecurityScanRole",
    externalId: "cc-ext-def456",
    verified: true,
    lastVerifiedAt: "2024-12-21T14:00:00Z",
    lastScanAt: "2024-12-21T15:00:00Z",
    findingsCount: 15,
  },
  {
    id: "3",
    accountId: "456789123456",
    name: "Staging",
    roleArn: "arn:aws:iam::456789123456:role/CloudCopSecurityScanRole",
    externalId: "cc-ext-ghi789",
    verified: false,
    lastVerifiedAt: null,
    lastScanAt: null,
    findingsCount: 0,
  },
];

function formatDate(dateString: string | null) {
  if (!dateString) return "Never";
  return new Date(dateString).toLocaleDateString("en-US", {
    month: "short",
    day: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  });
}

export function AccountsList() {
  if (accounts.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center py-12 text-center">
        <CloudIcon className="size-12 text-muted-foreground mb-4" />
        <h3 className="text-lg font-medium">No AWS accounts connected</h3>
        <p className="text-sm text-muted-foreground mt-1 mb-4">
          Connect your first AWS account to start scanning for security issues.
        </p>
      </div>
    );
  }

  return (
    <Table>
      <TableHeader>
        <TableRow>
          <TableHead>Account</TableHead>
          <TableHead>Account ID</TableHead>
          <TableHead>Status</TableHead>
          <TableHead>Last Scan</TableHead>
          <TableHead>Findings</TableHead>
          <TableHead className="text-right">Actions</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        {accounts.map((account) => (
          <TableRow key={account.id}>
            <TableCell>
              <div className="flex items-center gap-3">
                <div className="size-10 rounded-lg bg-primary/10 flex items-center justify-center">
                  <CloudIcon className="size-5 text-primary" />
                </div>
                <div>
                  <div className="font-medium">{account.name}</div>
                  <div className="text-xs text-muted-foreground truncate max-w-[200px]">
                    {account.roleArn}
                  </div>
                </div>
              </div>
            </TableCell>
            <TableCell>
              <code className="text-sm bg-muted px-2 py-1 rounded">
                {account.accountId}
              </code>
            </TableCell>
            <TableCell>
              {account.verified ? (
                <Badge
                  variant="default"
                  className="bg-green-500/10 text-green-500 hover:bg-green-500/20"
                >
                  <CheckCircleIcon className="mr-1 size-3" weight="fill" />
                  Verified
                </Badge>
              ) : (
                <Badge variant="secondary" className="text-yellow-500">
                  <WarningCircleIcon className="mr-1 size-3" weight="fill" />
                  Pending
                </Badge>
              )}
            </TableCell>
            <TableCell>
              <span className="text-sm text-muted-foreground">
                {formatDate(account.lastScanAt)}
              </span>
            </TableCell>
            <TableCell>
              {account.findingsCount > 0 ? (
                <Badge variant="secondary">{account.findingsCount}</Badge>
              ) : (
                <span className="text-sm text-muted-foreground">-</span>
              )}
            </TableCell>
            <TableCell className="text-right">
              <DropdownMenu>
                <DropdownMenuTrigger asChild>
                  <Button variant="ghost" size="icon">
                    <DotsThreeIcon className="size-4" weight="bold" />
                  </Button>
                </DropdownMenuTrigger>
                <DropdownMenuContent align="end">
                  {account.verified ? (
                    <>
                      <DropdownMenuItem asChild>
                        <Link href={`/scans?accountId=${account.accountId}`}>
                          <PlayIcon className="mr-2 size-4" />
                          Run Scan
                        </Link>
                      </DropdownMenuItem>
                      <DropdownMenuItem>
                        <ArrowClockwiseIcon className="mr-2 size-4" />
                        Re-verify Connection
                      </DropdownMenuItem>
                    </>
                  ) : (
                    <DropdownMenuItem>
                      <ArrowClockwiseIcon className="mr-2 size-4" />
                      Verify Connection
                    </DropdownMenuItem>
                  )}
                  <DropdownMenuSeparator />
                  <DropdownMenuItem className="text-destructive">
                    <TrashIcon className="mr-2 size-4" />
                    Remove Account
                  </DropdownMenuItem>
                </DropdownMenuContent>
              </DropdownMenu>
            </TableCell>
          </TableRow>
        ))}
      </TableBody>
    </Table>
  );
}
