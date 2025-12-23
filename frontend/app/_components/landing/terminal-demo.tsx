"use client";

import { cn } from "@/lib/utils";
import { useEffect, useState } from "react";

interface Command {
  description: string;
  command: string;
  status: "pending" | "running" | "success";
}

const fixCommands: Command[] = [
  {
    description: "Block public access on S3 bucket",
    command:
      "aws s3api put-public-access-block --bucket prod-data --public-access-block-configuration BlockPublicAcls=true,IgnorePublicAcls=true",
  },
  {
    description: "Revoke SSH access from 0.0.0.0/0",
    command:
      "aws ec2 revoke-security-group-ingress --group-id sg-abc123 --protocol tcp --port 22 --cidr 0.0.0.0/0",
  },
  {
    description: "Enable IMDSv2 on EC2 instance",
    command:
      "aws ec2 modify-instance-metadata-options --instance-id i-1234567890 --http-tokens required --http-endpoint enabled",
  },
  {
    description: "Attach least-privilege policy to Lambda role",
    command:
      "aws iam put-role-policy --role-name data-processor --policy-name MinimalAccess --policy-document file://policy.json",
  },
  {
    description: "Enable encryption on DynamoDB table",
    command:
      "aws dynamodb update-table --table-name user-sessions --sse-specification Enabled=true,SSEType=KMS",
  },
].map((cmd) => ({ ...cmd, status: "pending" as const }));

export function TerminalDemo({ className }: { className?: string }) {
  const [commands, setCommands] = useState(fixCommands);
  const [currentIndex, setCurrentIndex] = useState(0);

  useEffect(() => {
    if (currentIndex >= commands.length) {
      // Reset after all commands complete
      const timeout = setTimeout(() => {
        setCommands(fixCommands.map((cmd) => ({ ...cmd, status: "pending" })));
        setCurrentIndex(0);
      }, 3000);
      return () => clearTimeout(timeout);
    }

    // Run current command
    const runTimeout = setTimeout(() => {
      setCommands((prev) =>
        prev.map((cmd, idx) =>
          idx === currentIndex ? { ...cmd, status: "running" } : cmd,
        ),
      );
    }, 500);

    // Complete current command
    const completeTimeout = setTimeout(() => {
      setCommands((prev) =>
        prev.map((cmd, idx) =>
          idx === currentIndex ? { ...cmd, status: "success" } : cmd,
        ),
      );
      setCurrentIndex((prev) => prev + 1);
    }, 1500);

    return () => {
      clearTimeout(runTimeout);
      clearTimeout(completeTimeout);
    };
  }, [currentIndex, commands.length]);

  return (
    <div className={cn("p-4", className)}>
      {/* Terminal Window */}
      <div className="rounded-lg overflow-hidden border border-zinc-800 bg-zinc-950 shadow-2xl">
        {/* Terminal Header */}
        <div className="flex items-center gap-2 px-4 py-3 bg-zinc-900 border-b border-zinc-800">
          <div className="flex gap-2">
            <div className="w-3 h-3 rounded-full bg-red-500" />
            <div className="w-3 h-3 rounded-full bg-yellow-500" />
            <div className="w-3 h-3 rounded-full bg-green-500" />
          </div>
          <span className="text-xs text-zinc-400 ml-2 font-mono">
            cloudcop-remediation
          </span>
        </div>

        {/* Terminal Content */}
        <div className="p-4 font-mono text-sm space-y-4 max-h-[320px] overflow-hidden">
          {commands.map((cmd, idx) => (
            <div
              key={idx}
              className={cn(
                "transition-all duration-300",
                idx > currentIndex && "opacity-30",
              )}
            >
              {/* Command description */}
              <div className="flex items-center gap-2 mb-1">
                <span className="text-zinc-500 text-xs">#</span>
                <span className="text-zinc-400 text-xs">{cmd.description}</span>
                {cmd.status === "success" && (
                  <span className="text-green-400 text-xs ml-auto">
                    ✓ Fixed
                  </span>
                )}
                {cmd.status === "running" && (
                  <span className="text-yellow-400 text-xs ml-auto animate-pulse">
                    Running...
                  </span>
                )}
              </div>

              {/* Command */}
              <div className="flex items-start gap-2">
                <span className="text-green-400">$</span>
                <span
                  className={cn(
                    "text-zinc-300 break-all",
                    cmd.status === "running" && "text-white",
                    cmd.status === "success" && "text-zinc-500",
                  )}
                >
                  {cmd.command}
                </span>
              </div>

              {/* Output for completed commands */}
              {cmd.status === "success" && (
                <div className="ml-4 mt-1 text-xs text-green-400/70">
                  {idx === 0 &&
                    "PublicAccessBlockConfiguration applied successfully."}
                  {idx === 1 && "Ingress rule removed from security group."}
                  {idx === 2 && "Instance metadata options updated."}
                  {idx === 3 && "Role policy attached successfully."}
                  {idx === 4 && "Server-side encryption enabled."}
                </div>
              )}
            </div>
          ))}

          {/* Cursor */}
          {currentIndex >= commands.length && (
            <div className="flex items-center gap-2 mt-4">
              <span className="text-green-400">$</span>
              <span className="text-zinc-300 animate-pulse">
                All security issues remediated! ✨
              </span>
            </div>
          )}
        </div>
      </div>

      {/* Stats bar */}
      <div className="flex items-center justify-between mt-4 px-2">
        <div className="flex items-center gap-4 text-xs">
          <div className="flex items-center gap-1.5">
            <div className="w-2 h-2 rounded-full bg-green-500" />
            <span className="text-zinc-400">
              {commands.filter((c) => c.status === "success").length} Fixed
            </span>
          </div>
          <div className="flex items-center gap-1.5">
            <div className="w-2 h-2 rounded-full bg-yellow-500 animate-pulse" />
            <span className="text-zinc-400">
              {commands.filter((c) => c.status === "running").length} Running
            </span>
          </div>
          <div className="flex items-center gap-1.5">
            <div className="w-2 h-2 rounded-full bg-zinc-500" />
            <span className="text-zinc-400">
              {commands.filter((c) => c.status === "pending").length} Pending
            </span>
          </div>
        </div>
        <span className="text-xs text-zinc-500">One-click remediation</span>
      </div>
    </div>
  );
}
