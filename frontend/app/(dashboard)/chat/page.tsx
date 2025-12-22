"use client";

import * as React from "react";
import { PageHeader } from "@/components/page-header";
import { Button } from "@/components/ui/button";
import { Textarea } from "@/components/ui/textarea";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Badge } from "@/components/ui/badge";
import { Avatar, AvatarFallback } from "@/components/ui/avatar";
import {
  PaperPlaneTiltIcon,
  RobotIcon,
  SparkleIcon,
} from "@phosphor-icons/react";
import { ChatMessage } from "./_components/chat-message";
import { SuggestedQuestions } from "./_components/suggested-questions";

interface Message {
  id: string;
  role: "user" | "assistant";
  content: string;
  timestamp: Date;
}

const INITIAL_MESSAGE_CONTENT =
  "Hello! I'm CloudCop AI, your cloud security assistant. I can help you understand your AWS environment, find security issues, and generate remediation commands. What would you like to know?";

export default function ChatPage() {
  const [messages, setMessages] = React.useState<Message[]>([]);
  const [input, setInput] = React.useState("");
  const [isLoading, setIsLoading] = React.useState(false);
  const [isInitialized, setIsInitialized] = React.useState(false);
  const scrollRef = React.useRef<HTMLDivElement>(null);

  // Initialize messages on client side only to avoid hydration mismatch
  React.useEffect(() => {
    if (!isInitialized) {
      setMessages([
        {
          id: "1",
          role: "assistant",
          content: INITIAL_MESSAGE_CONTENT,
          timestamp: new Date(),
        },
      ]);
      setIsInitialized(true);
    }
  }, [isInitialized]);

  const handleSend = async () => {
    if (!input.trim() || isLoading) return;

    const userMessage = {
      id: Date.now().toString(),
      role: "user" as const,
      content: input,
      timestamp: new Date(),
    };

    setMessages((prev) => [...prev, userMessage]);
    setInput("");
    setIsLoading(true);

    // Simulate AI response - will be replaced with real API call
    setTimeout(() => {
      const aiResponse = {
        id: (Date.now() + 1).toString(),
        role: "assistant" as const,
        content: getAIResponse(input),
        timestamp: new Date(),
      };
      setMessages((prev) => [...prev, aiResponse]);
      setIsLoading(false);
    }, 1500);
  };

  const handleSuggestedQuestion = (question: string) => {
    setInput(question);
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === "Enter" && !e.shiftKey) {
      e.preventDefault();
      handleSend();
    }
  };

  React.useEffect(() => {
    if (scrollRef.current) {
      scrollRef.current.scrollTop = scrollRef.current.scrollHeight;
    }
  }, [messages]);

  return (
    <>
      <PageHeader
        title="Chat"
        breadcrumbs={[{ label: "Chat with your Cloud" }]}
        actions={
          <Badge variant="secondary" className="gap-1">
            <SparkleIcon className="size-3" />
            AI Powered
          </Badge>
        }
      />

      <div className="flex flex-1 flex-col h-[calc(100vh-4rem)]">
        <div className="flex-1 overflow-hidden">
          <ScrollArea className="h-full p-6" ref={scrollRef}>
            <div className="max-w-3xl mx-auto space-y-6">
              {messages.map((message) => (
                <ChatMessage key={message.id} message={message} />
              ))}
              {isLoading && (
                <div className="flex items-start gap-4">
                  <Avatar className="size-8">
                    <AvatarFallback className="bg-primary/10 text-primary">
                      <RobotIcon className="size-4" />
                    </AvatarFallback>
                  </Avatar>
                  <div className="flex-1">
                    <div className="flex items-center gap-2 text-sm text-muted-foreground">
                      <span>CloudCop AI</span>
                      <span className="animate-pulse">is thinking...</span>
                    </div>
                    <div className="mt-2 flex gap-1">
                      <span className="size-2 rounded-full bg-primary/50 animate-bounce" />
                      <span
                        className="size-2 rounded-full bg-primary/50 animate-bounce"
                        style={{ animationDelay: "0.1s" }}
                      />
                      <span
                        className="size-2 rounded-full bg-primary/50 animate-bounce"
                        style={{ animationDelay: "0.2s" }}
                      />
                    </div>
                  </div>
                </div>
              )}

              {messages.length === 1 && !isLoading && (
                <SuggestedQuestions onSelect={handleSuggestedQuestion} />
              )}
            </div>
          </ScrollArea>
        </div>

        {/* Input Area */}
        <div className="border-t bg-background p-4">
          <div className="max-w-3xl mx-auto">
            <div className="flex gap-2">
              <Textarea
                placeholder="Ask about your AWS environment..."
                value={input}
                onChange={(e) => setInput(e.target.value)}
                onKeyDown={handleKeyDown}
                className="min-h-[60px] resize-none"
                rows={2}
              />
              <Button
                size="lg"
                className="h-auto px-4"
                onClick={handleSend}
                disabled={!input.trim() || isLoading}
              >
                <PaperPlaneTiltIcon className="size-5" />
              </Button>
            </div>
            <p className="text-xs text-muted-foreground mt-2 text-center">
              CloudCop AI can query your AWS resources, find security issues,
              and generate remediation commands.
            </p>
          </div>
        </div>
      </div>
    </>
  );
}

// Mock AI response generator
function getAIResponse(question: string): string {
  const lowerQuestion = question.toLowerCase();

  if (lowerQuestion.includes("ec2") && lowerQuestion.includes("running")) {
    return `Based on my analysis of your AWS environment, you have **47 running EC2 instances** across your accounts.

**Breakdown by region:**
- us-east-1: 28 instances
- us-west-2: 12 instances
- eu-west-1: 7 instances

**Notable findings:**
- 3 instances have IMDSv1 enabled (security risk)
- 5 instances have public IP addresses
- 2 instances are running outdated AMIs

Would you like me to show details about any specific instances or generate remediation commands?`;
  }

  if (lowerQuestion.includes("s3") && lowerQuestion.includes("public")) {
    return `I found **2 S3 buckets with public access** enabled:

1. **prod-data-bucket** (us-east-1)
   - Public read access enabled
   - Contains sensitive data
   - **CRITICAL** - Immediate action required

2. **staging-assets** (us-west-2)
   - Public list access enabled
   - Contains non-sensitive assets

**Remediation command for prod-data-bucket:**
\`\`\`bash
aws s3api put-public-access-block \\
  --bucket prod-data-bucket \\
  --public-access-block-configuration "BlockPublicAcls=true,IgnorePublicAcls=true,BlockPublicPolicy=true,RestrictPublicBuckets=true"
\`\`\`

Would you like me to generate commands for both buckets?`;
  }

  if (lowerQuestion.includes("attack") && lowerQuestion.includes("path")) {
    return `I've analyzed your AWS environment and found **3 attack paths** to sensitive resources:

**Path 1: Internet -> Database (CRITICAL)**
\`\`\`
Internet (0.0.0.0/0)
  -> Security Group sg-abc123 (SSH open)
  -> EC2 web-server-1
  -> IAM Role WebServerRole
  -> RDS prod-customers
\`\`\`
**Risk:** An attacker with SSH access could exfiltrate customer data.

**Path 2: Lambda -> S3 (HIGH)**
\`\`\`
Lambda data-processor
  -> IAM Role DataRole (overprivileged)
  -> S3 prod-data-bucket (sensitive)
\`\`\`
**Risk:** Lambda has more permissions than needed.

**Path 3: EC2 -> IAM (MEDIUM)**
\`\`\`
EC2 admin-box
  -> IAM Role AdminRole
  -> Full account access
\`\`\`
**Risk:** Compromised instance could control entire account.

Would you like remediation steps for any of these paths?`;
  }

  if (lowerQuestion.includes("critical") || lowerQuestion.includes("urgent")) {
    return `You have **3 critical findings** that require immediate attention:

1. **S3 bucket allows public write access** (prod-data-bucket)
   - Risk: Data tampering or deletion
   - Affected data: Customer PII

2. **Security group allows SSH from 0.0.0.0/0** (sg-0123456789)
   - Risk: Brute force attacks
   - Affected: 5 EC2 instances

3. **IAM user has AdministratorAccess** (legacy-admin)
   - Risk: Account takeover
   - No MFA enabled

I can generate AWS CLI commands to fix all of these. Would you like me to proceed?`;
  }

  return `I understand you're asking about "${question}".

To give you the most accurate information, I'll need to analyze your AWS environment. Here's what I can help with:

- **Resource queries**: Count instances, list resources, find configurations
- **Security analysis**: Find vulnerabilities, attack paths, misconfigurations
- **Compliance checks**: CIS benchmarks, SOC 2, PCI-DSS, GDPR
- **Remediation**: Generate AWS CLI commands to fix issues

Could you be more specific about what you'd like to know? For example:
- "How many EC2 instances are running?"
- "Find S3 buckets with public access"
- "Show attack paths to my database"`;
}
