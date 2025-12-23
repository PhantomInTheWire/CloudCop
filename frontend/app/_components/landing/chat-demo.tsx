"use client";

import { cn } from "@/lib/utils";
import { useEffect, useState } from "react";

const chatMessages = [
  {
    role: "user",
    content: "What EC2 instances can access my production database?",
  },
  {
    role: "assistant",
    content:
      "Found 3 EC2 instances with access:\n• web-server-1 via IAM Role WebServerRole\n• admin-box via AdminRole (CRITICAL)\n• data-pipeline via DataRole (read-only)",
  },
  {
    role: "user",
    content: "Show me attack paths to that database",
  },
  {
    role: "assistant",
    content:
      "Found 2 attack paths:\n\n1. Internet → SSH (0.0.0.0/0) → EC2 → RDS\n2. Internet → ALB → EC2 admin → IAM Role → RDS\n\nRecommendation: Fix Path 1 immediately.",
  },
];

export function ChatDemo({ className }: { className?: string }) {
  const [visibleMessages, setVisibleMessages] = useState(0);

  useEffect(() => {
    const interval = setInterval(() => {
      setVisibleMessages((prev) =>
        prev < chatMessages.length ? prev + 1 : prev,
      );
    }, 1500);

    return () => clearInterval(interval);
  }, []);

  return (
    <div className={cn("flex flex-col gap-3 p-4 overflow-hidden", className)}>
      {chatMessages.slice(0, visibleMessages).map((message, idx) => (
        <div
          key={idx}
          className={cn(
            "max-w-[85%] rounded-2xl px-4 py-2 text-xs",
            message.role === "user"
              ? "self-end bg-primary text-primary-foreground"
              : "self-start bg-muted text-foreground whitespace-pre-wrap",
          )}
        >
          {message.content}
        </div>
      ))}
      {visibleMessages < chatMessages.length && (
        <div className="self-start flex gap-1 px-4 py-2">
          <div className="h-2 w-2 rounded-full bg-muted-foreground/40 animate-bounce" />
          <div className="h-2 w-2 rounded-full bg-muted-foreground/40 animate-bounce delay-100" />
          <div className="h-2 w-2 rounded-full bg-muted-foreground/40 animate-bounce delay-200" />
        </div>
      )}
    </div>
  );
}
