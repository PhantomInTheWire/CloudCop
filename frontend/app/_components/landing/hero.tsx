"use client";

import Link from "next/link";
import { Spotlight } from "@/components/ui/spotlight-new";
import { TextGenerateEffect } from "@/components/ui/text-generate-effect";
import { Button } from "@/components/ui/moving-border";

export function Hero() {
  return (
    <section className="relative min-h-screen w-full flex flex-col items-center justify-center overflow-hidden bg-background">
      <Spotlight />

      <div className="relative z-10 flex flex-col items-center justify-center px-4 text-center">
        {/* Logo/Brand */}
        <div className="mb-6">
          <h1 className="text-5xl md:text-7xl font-bold tracking-tight">
            <span className="bg-gradient-to-r from-primary via-chart-3 to-chart-5 bg-clip-text text-transparent">
              CloudCop
            </span>
          </h1>
        </div>

        {/* Animated Tagline */}
        <div className="max-w-3xl">
          <TextGenerateEffect
            words="AI-Powered Cloud Security Platform"
            className="text-3xl md:text-5xl font-bold"
            duration={0.4}
          />
        </div>

        {/* Subheadline */}
        <p className="mt-6 max-w-2xl text-lg md:text-xl text-muted-foreground">
          Scan AWS environments. Discover attack paths. Chat with your cloud.
          Generate fixes.
        </p>

        {/* CTA Buttons */}
        <div className="mt-10 flex flex-col sm:flex-row items-center gap-4">
          <Button
            as={Link}
            href="/dashboard"
            containerClassName="h-14 w-48"
            borderClassName="bg-[radial-gradient(var(--primary)_40%,transparent_60%)]"
            className="border-primary/20 bg-primary/10 text-foreground font-semibold hover:bg-primary/20"
          >
            Go to Dashboard
          </Button>

          <Link
            href="#features"
            className="px-6 py-3 text-muted-foreground hover:text-foreground transition-colors font-medium"
          >
            Learn More
          </Link>
        </div>
      </div>

      {/* Scroll indicator */}
      <div className="absolute bottom-8 left-1/2 -translate-x-1/2 animate-bounce">
        <svg
          className="w-6 h-6 text-muted-foreground"
          fill="none"
          strokeLinecap="round"
          strokeLinejoin="round"
          strokeWidth="2"
          viewBox="0 0 24 24"
          stroke="currentColor"
        >
          <path d="M19 14l-7 7m0 0l-7-7m7 7V3"></path>
        </svg>
      </div>
    </section>
  );
}
