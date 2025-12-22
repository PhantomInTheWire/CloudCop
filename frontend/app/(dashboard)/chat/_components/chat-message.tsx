"use client";

import * as React from "react";
import { Avatar, AvatarFallback } from "@/components/ui/avatar";
import { Button } from "@/components/ui/button";
import {
  RobotIcon,
  UserIcon,
  CopyIcon,
  CheckIcon,
} from "@phosphor-icons/react";

interface Message {
  id: string;
  role: "user" | "assistant";
  content: string;
  timestamp: Date;
}

interface ChatMessageProps {
  message: Message;
}

export function ChatMessage({ message }: ChatMessageProps) {
  const [copiedCode, setCopiedCode] = React.useState<string | null>(null);
  const isAssistant = message.role === "assistant";

  const copyToClipboard = (code: string) => {
    navigator.clipboard.writeText(code);
    setCopiedCode(code);
    setTimeout(() => setCopiedCode(null), 2000);
  };

  return (
    <div
      className={`flex items-start gap-4 ${isAssistant ? "" : "flex-row-reverse"}`}
    >
      <Avatar className="size-8">
        <AvatarFallback
          className={
            isAssistant
              ? "bg-primary/10 text-primary"
              : "bg-muted text-muted-foreground"
          }
        >
          {isAssistant ? (
            <RobotIcon className="size-4" />
          ) : (
            <UserIcon className="size-4" />
          )}
        </AvatarFallback>
      </Avatar>

      <div className={`flex-1 ${isAssistant ? "" : "text-right"}`}>
        <div className="flex items-center gap-2 text-sm text-muted-foreground mb-1">
          <span>{isAssistant ? "CloudCop AI" : "You"}</span>
          <span className="text-xs">
            {message.timestamp.toLocaleTimeString([], {
              hour: "2-digit",
              minute: "2-digit",
            })}
          </span>
        </div>

        <div
          className={`prose prose-sm dark:prose-invert max-w-none ${
            isAssistant
              ? "bg-muted/50 rounded-lg p-4"
              : "bg-primary text-primary-foreground rounded-lg p-4 inline-block text-left"
          }`}
        >
          {isAssistant ? (
            <div className="space-y-4">
              {message.content.split("```").map((part, index) => {
                if (index % 2 === 1) {
                  // This is a code block
                  const [lang, ...codeLines] = part.split("\n");
                  const code = codeLines.join("\n").trim();
                  return (
                    <div
                      key={index}
                      className="relative rounded-lg bg-background border overflow-hidden"
                    >
                      <div className="flex items-center justify-between px-3 py-2 border-b bg-muted/50">
                        <span className="text-xs text-muted-foreground font-mono">
                          {lang || "bash"}
                        </span>
                        <Button
                          variant="ghost"
                          size="sm"
                          className="h-6 px-2"
                          onClick={() => copyToClipboard(code)}
                        >
                          {copiedCode === code ? (
                            <>
                              <CheckIcon className="size-3 mr-1" />
                              Copied
                            </>
                          ) : (
                            <>
                              <CopyIcon className="size-3 mr-1" />
                              Copy
                            </>
                          )}
                        </Button>
                      </div>
                      <pre className="p-3 text-xs font-mono overflow-x-auto">
                        <code>{code}</code>
                      </pre>
                    </div>
                  );
                }
                // Regular text - render with markdown-like formatting
                return (
                  <div key={index} className="whitespace-pre-wrap">
                    {part.split("\n").map((line, lineIndex) => {
                      // Handle bold text
                      const formattedLine = line.replace(
                        /\*\*(.*?)\*\*/g,
                        "<strong>$1</strong>",
                      );
                      // Handle inline code
                      const withCode = formattedLine.replace(
                        /`([^`]+)`/g,
                        '<code class="bg-muted px-1 py-0.5 rounded text-sm font-mono">$1</code>',
                      );

                      if (line.startsWith("- ")) {
                        return (
                          <div
                            key={lineIndex}
                            className="flex gap-2 ml-4"
                            dangerouslySetInnerHTML={{
                              __html: `<span class="text-muted-foreground">•</span><span>${withCode.slice(2)}</span>`,
                            }}
                          />
                        );
                      }
                      return (
                        <p
                          key={lineIndex}
                          dangerouslySetInnerHTML={{ __html: withCode }}
                        />
                      );
                    })}
                  </div>
                );
              })}
            </div>
          ) : (
            <p className="m-0">{message.content}</p>
          )}
        </div>
      </div>
    </div>
  );
}
