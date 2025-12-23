"use client";

import { BentoCard, BentoGrid } from "@/components/ui/bento-grid";
import { Marquee } from "@/components/ui/marquee";
import { cn } from "@/lib/utils";
import {
  ShieldCheck,
  Graph,
  TerminalWindow,
  Chats,
} from "@phosphor-icons/react/dist/ssr";
import { AttackPathDemo } from "./attack-path-demo";
import { SecurityFindingsDemo } from "./security-findings-demo";
import { ChatDemo } from "./chat-demo";
import { TerminalDemo } from "./terminal-demo";

// AWS Services for marquee
const awsServices = [
  {
    name: "S3 Buckets",
    body: "Public access, encryption, versioning, logging checks",
  },
  {
    name: "EC2 Instances",
    body: "Security groups, IMDSv2, encryption, public IPs",
  },
  {
    name: "IAM Roles",
    body: "Unused credentials, overly permissive policies, MFA",
  },
  {
    name: "Lambda",
    body: "Environment variables, VPC config, IAM roles",
  },
  {
    name: "ECS Tasks",
    body: "Task definitions, container IAM, network security",
  },
  {
    name: "DynamoDB",
    body: "Encryption, PITR, auto-scaling, TTL settings",
  },
];

const features = [
  {
    Icon: ShieldCheck,
    name: "60+ Security Rules",
    description:
      "Comprehensive checks for S3, EC2, IAM, Lambda, ECS, and DynamoDB with AI-powered summarization.",
    href: "/dashboard",
    cta: "Start Scanning",
    className: "col-span-3 lg:col-span-1",
    background: (
      <Marquee
        pauseOnHover
        className="absolute top-10 [mask-image:linear-gradient(to_top,transparent_40%,#000_100%)] [--duration:20s]"
      >
        {awsServices.map((service, idx) => (
          <figure
            key={idx}
            className={cn(
              "relative w-36 cursor-pointer overflow-hidden rounded-xl border p-4",
              "border-border bg-muted/50 hover:bg-muted",
              "transform-gpu blur-[1px] transition-all duration-300 ease-out hover:blur-none",
            )}
          >
            <div className="flex flex-col">
              <figcaption className="text-sm font-medium text-foreground">
                {service.name}
              </figcaption>
              <blockquote className="mt-2 text-xs text-muted-foreground">
                {service.body}
              </blockquote>
            </div>
          </figure>
        ))}
      </Marquee>
    ),
  },
  {
    Icon: Graph,
    name: "Security Findings",
    description:
      "AI-powered summarization reduces alert fatigue by intelligently grouping findings.",
    href: "/findings",
    cta: "View Findings",
    className: "col-span-3 lg:col-span-2",
    background: (
      <SecurityFindingsDemo className="absolute top-4 right-2 h-[300px] w-full scale-90 border-none [mask-image:linear-gradient(to_top,transparent_10%,#000_100%)] transition-all duration-300 ease-out group-hover:scale-95" />
    ),
  },
  {
    Icon: Graph,
    name: "Attack Path Discovery",
    description:
      "Neo4j-powered graph analysis discovers how attackers could chain vulnerabilities.",
    href: "/dashboard",
    cta: "Explore Paths",
    className: "col-span-3 lg:col-span-2",
    background: (
      <AttackPathDemo className="absolute top-4 right-2 h-[300px] border-none [mask-image:linear-gradient(to_top,transparent_10%,#000_100%)] transition-all duration-300 ease-out group-hover:scale-105" />
    ),
  },
  {
    Icon: Chats,
    name: "Chat with Your Cloud",
    description:
      'Ask natural language questions like "What can access my database?"',
    className: "col-span-3 lg:col-span-1",
    href: "/chat",
    cta: "Start Chatting",
    background: (
      <ChatDemo className="absolute top-10 right-0 left-0 h-[280px] origin-top [mask-image:linear-gradient(to_top,transparent_20%,#000_100%)] transition-all duration-300 ease-out group-hover:scale-105" />
    ),
  },
  {
    Icon: TerminalWindow,
    name: "Fix Commands",
    description:
      "Get ready-to-run AWS CLI commands to remediate issues instantly. One-click fixes for common security misconfigurations.",
    className: "col-span-3 lg:col-span-3",
    href: "/dashboard",
    cta: "Generate Fixes",
    background: (
      <TerminalDemo className="absolute inset-0 [mask-image:linear-gradient(to_top,transparent_30%,#000_100%)] transition-all duration-300 ease-out" />
    ),
  },
];

export function Features() {
  return (
    <section id="features" className="w-full py-20 md:py-32 px-4 bg-background">
      <div className="max-w-7xl mx-auto">
        {/* Section Header */}
        <div className="text-center mb-16">
          <h2 className="text-3xl md:text-4xl font-bold mb-4">
            Four Core Features
          </h2>
          <p className="text-lg text-muted-foreground max-w-2xl mx-auto">
            Everything you need to secure your cloud infrastructure, powered by
            AI and graph technology.
          </p>
        </div>

        {/* Bento Grid */}
        <BentoGrid>
          {features.map((feature, idx) => (
            <BentoCard key={idx} {...feature} />
          ))}
        </BentoGrid>
      </div>
    </section>
  );
}
