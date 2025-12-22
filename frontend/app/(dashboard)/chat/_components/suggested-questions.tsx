"use client";

import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import {
  ShieldWarningIcon,
  PathIcon,
  DatabaseIcon,
} from "@phosphor-icons/react";

const suggestedQuestions = [
  {
    category: "Resources",
    icon: DatabaseIcon,
    questions: [
      "How many EC2 instances are running?",
      "List all S3 buckets in my account",
      "Which IAM users have console access?",
    ],
  },
  {
    category: "Security",
    icon: ShieldWarningIcon,
    questions: [
      "Find S3 buckets with public access",
      "Which security groups allow SSH from anywhere?",
      "Show me critical security findings",
    ],
  },
  {
    category: "Attack Paths",
    icon: PathIcon,
    questions: [
      "Find attack paths to my database",
      "How can an attacker reach sensitive S3 buckets?",
      "Show privilege escalation paths",
    ],
  },
];

interface SuggestedQuestionsProps {
  onSelect: (question: string) => void;
}

export function SuggestedQuestions({ onSelect }: SuggestedQuestionsProps) {
  return (
    <div className="space-y-4 mt-8">
      <div className="text-center">
        <h3 className="text-lg font-semibold">How can I help?</h3>
        <p className="text-sm text-muted-foreground">
          Here are some things you can ask me about your AWS environment
        </p>
      </div>

      <div className="grid gap-4 md:grid-cols-3">
        {suggestedQuestions.map((category) => (
          <Card key={category.category} className="bg-muted/30">
            <CardHeader className="pb-3">
              <CardTitle className="text-sm flex items-center gap-2">
                <category.icon className="size-4 text-primary" />
                {category.category}
              </CardTitle>
            </CardHeader>
            <CardContent className="space-y-2">
              {category.questions.map((question) => (
                <Button
                  key={question}
                  variant="ghost"
                  className="w-full justify-start text-left h-auto py-2 px-3 font-normal text-sm hover:bg-background whitespace-normal"
                  onClick={() => onSelect(question)}
                >
                  {question}
                </Button>
              ))}
            </CardContent>
          </Card>
        ))}
      </div>
    </div>
  );
}
