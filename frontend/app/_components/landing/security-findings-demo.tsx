"use client";

import { cn } from "@/lib/utils";
import { AnimatedList } from "@/components/ui/animated-list";

interface Finding {
  title: string;
  description: string;
  severity: "CRITICAL" | "HIGH" | "MEDIUM" | "LOW";
  icon: string;
  color: string;
}

const findings: Finding[] = [
  {
    title: "S3 Bucket Public Access",
    description: "prod-data-bucket allows public read",
    severity: "CRITICAL",
    icon: "🪣",
    color: "#ef4444",
  },
  {
    title: "IAM Overprivileged Role",
    description: "WebServerRole has admin access",
    severity: "HIGH",
    icon: "🔑",
    color: "#f97316",
  },
  {
    title: "EC2 Exposed to Internet",
    description: "SSH open on 0.0.0.0/0",
    severity: "HIGH",
    icon: "🖥️",
    color: "#f97316",
  },
  {
    title: "Lambda Missing VPC",
    description: "data-processor not in VPC",
    severity: "MEDIUM",
    icon: "⚡",
    color: "#eab308",
  },
  {
    title: "DynamoDB No Encryption",
    description: "user-sessions table unencrypted",
    severity: "MEDIUM",
    icon: "📊",
    color: "#eab308",
  },
  {
    title: "Security Group Unused",
    description: "sg-abc123 has no instances",
    severity: "LOW",
    icon: "🛡️",
    color: "#22c55e",
  },
];

// Duplicate for continuous animation
const allFindings = [...findings, ...findings];

const Finding = ({ title, description, severity, icon, color }: Finding) => {
  return (
    <figure
      className={cn(
        "relative mx-auto min-h-fit w-full max-w-[400px] cursor-pointer overflow-hidden rounded-2xl p-4",
        "bg-background [box-shadow:0_0_0_1px_rgba(0,0,0,.03),0_2px_4px_rgba(0,0,0,.05),0_12px_24px_rgba(0,0,0,.05)]",
        "dark:[box-shadow:0_-20px_80px_-20px_#ffffff1f_inset] dark:[border:1px_solid_rgba(255,255,255,.1)]",
        "transform-gpu transition-all duration-200 ease-in-out hover:scale-[103%]",
      )}
    >
      <div className="flex flex-row items-center gap-3">
        <div
          className="flex size-10 items-center justify-center rounded-2xl"
          style={{ backgroundColor: `${color}20` }}
        >
          <span className="text-lg">{icon}</span>
        </div>
        <div className="flex flex-col overflow-hidden">
          <figcaption className="flex flex-row items-center gap-2 whitespace-pre text-lg font-medium dark:text-white">
            <span className="text-sm">{title}</span>
            <span
              className="text-xs font-semibold px-2 py-0.5 rounded-full"
              style={{
                backgroundColor: `${color}20`,
                color: color,
              }}
            >
              {severity}
            </span>
          </figcaption>
          <p className="text-sm font-normal text-muted-foreground truncate">
            {description}
          </p>
        </div>
      </div>
    </figure>
  );
};

export function SecurityFindingsDemo({ className }: { className?: string }) {
  return (
    <div
      className={cn(
        "relative flex h-[300px] w-full flex-col overflow-hidden p-6",
        className,
      )}
    >
      <AnimatedList delay={2000}>
        {allFindings.map((finding, idx) => (
          <Finding {...finding} key={idx} />
        ))}
      </AnimatedList>
    </div>
  );
}
