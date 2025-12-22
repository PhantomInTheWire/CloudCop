"use client";

import { Button } from "@/components/ui/button";
import Link from "next/link";
import {
  PlayIcon,
  ChatCircleDotsIcon,
  CloudIcon,
  LightningIcon,
  ArrowRightIcon,
} from "@phosphor-icons/react";

const actions = [
  {
    title: "Run Security Scan",
    description: "Scan your AWS environment for vulnerabilities",
    icon: PlayIcon,
    href: "/scans",
    variant: "default" as const,
  },
  {
    title: "Chat with Cloud",
    description: "Ask questions about your infrastructure",
    icon: ChatCircleDotsIcon,
    href: "/chat",
    variant: "outline" as const,
  },
  {
    title: "Connect AWS Account",
    description: "Add a new AWS account to monitor",
    icon: CloudIcon,
    href: "/accounts",
    variant: "outline" as const,
  },
  {
    title: "View Critical Findings",
    description: "3 issues require immediate attention",
    icon: LightningIcon,
    href: "/findings?severity=CRITICAL",
    variant: "outline" as const,
  },
];

export function QuickActions() {
  return (
    <div className="space-y-3">
      {actions.map((action) => (
        <Button
          key={action.title}
          variant={action.variant}
          className="w-full justify-start h-auto py-3"
          asChild
        >
          <Link href={action.href}>
            <action.icon className="size-5 mr-3" />
            <div className="flex-1 text-left">
              <div className="font-medium">{action.title}</div>
              <div className="text-xs text-muted-foreground font-normal">
                {action.description}
              </div>
            </div>
            <ArrowRightIcon className="size-4 ml-2" />
          </Link>
        </Button>
      ))}
    </div>
  );
}
