"use client";

import Link from "next/link";
import { Particles } from "@/components/ui/particles";
import { RetroGrid } from "@/components/ui/retro-grid";
import { OrbitingCircles } from "@/components/ui/orbiting-circles";
import { BorderBeam } from "@/components/ui/border-beam";
import { WordRotate } from "@/components/ui/word-rotate";
import { ShimmerButton } from "@/components/ui/shimmer-button";
import { useTheme } from "next-themes";
import { useEffect, useState } from "react";

// AWS Service Icons
const S3Icon = () => (
  <div className="flex items-center justify-center size-10 rounded-xl bg-green-500/20 border border-green-500/30 shadow-lg shadow-green-500/20">
    <svg viewBox="0 0 24 24" className="size-5 text-green-400">
      <path
        fill="currentColor"
        d="M12 2L2 7l10 5 10-5-10-5zM2 17l10 5 10-5M2 12l10 5 10-5"
      />
    </svg>
  </div>
);

const EC2Icon = () => (
  <div className="flex items-center justify-center size-10 rounded-xl bg-orange-500/20 border border-orange-500/30 shadow-lg shadow-orange-500/20">
    <svg viewBox="0 0 24 24" className="size-5 text-orange-400">
      <rect x="4" y="4" width="16" height="16" rx="2" fill="currentColor" />
      <rect
        x="7"
        y="7"
        width="4"
        height="4"
        fill="currentColor"
        opacity="0.5"
      />
      <rect
        x="13"
        y="7"
        width="4"
        height="4"
        fill="currentColor"
        opacity="0.5"
      />
      <rect
        x="7"
        y="13"
        width="4"
        height="4"
        fill="currentColor"
        opacity="0.5"
      />
      <rect
        x="13"
        y="13"
        width="4"
        height="4"
        fill="currentColor"
        opacity="0.5"
      />
    </svg>
  </div>
);

const IAMIcon = () => (
  <div className="flex items-center justify-center size-10 rounded-xl bg-yellow-500/20 border border-yellow-500/30 shadow-lg shadow-yellow-500/20">
    <svg viewBox="0 0 24 24" className="size-5 text-yellow-400">
      <path
        fill="currentColor"
        d="M12 1L3 5v6c0 5.55 3.84 10.74 9 12 5.16-1.26 9-6.45 9-12V5l-9-4zm0 4a3 3 0 110 6 3 3 0 010-6zm0 8c-2 0-6 1-6 3v1h12v-1c0-2-4-3-6-3z"
      />
    </svg>
  </div>
);

const LambdaIcon = () => (
  <div className="flex items-center justify-center size-10 rounded-xl bg-purple-500/20 border border-purple-500/30 shadow-lg shadow-purple-500/20">
    <svg viewBox="0 0 24 24" className="size-5 text-purple-400">
      <path
        fill="currentColor"
        d="M4 20h4l4-8-2-4-6 12zm8-16l2 4 6 12h-4l-4-8-4 8H4l8-16z"
      />
    </svg>
  </div>
);

const RDSIcon = () => (
  <div className="flex items-center justify-center size-10 rounded-xl bg-blue-500/20 border border-blue-500/30 shadow-lg shadow-blue-500/20">
    <svg viewBox="0 0 24 24" className="size-5 text-blue-400">
      <ellipse cx="12" cy="6" rx="8" ry="3" fill="currentColor" />
      <path
        fill="currentColor"
        d="M4 6v4c0 1.66 3.58 3 8 3s8-1.34 8-3V6c0 1.66-3.58 3-8 3S4 7.66 4 6z"
      />
      <path
        fill="currentColor"
        d="M4 10v4c0 1.66 3.58 3 8 3s8-1.34 8-3v-4c0 1.66-3.58 3-8 3s-8-1.34-8-3z"
      />
      <path
        fill="currentColor"
        d="M4 14v4c0 1.66 3.58 3 8 3s8-1.34 8-3v-4c0 1.66-3.58 3-8 3s-8-1.34-8-3z"
      />
    </svg>
  </div>
);

const DynamoDBIcon = () => (
  <div className="flex items-center justify-center size-10 rounded-xl bg-cyan-500/20 border border-cyan-500/30 shadow-lg shadow-cyan-500/20">
    <svg viewBox="0 0 24 24" className="size-5 text-cyan-400">
      <path
        fill="currentColor"
        d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm-1 17.93c-3.95-.49-7-3.85-7-7.93s3.05-7.44 7-7.93v15.86zm2-15.86c3.95.49 7 3.85 7 7.93s-3.05 7.44-7 7.93V4.07z"
      />
    </svg>
  </div>
);

// Central Shield Icon
const ShieldIcon = () => (
  <div className="flex items-center justify-center size-16 rounded-2xl bg-primary/20 border-2 border-primary/40 shadow-2xl shadow-primary/30">
    <svg viewBox="0 0 24 24" className="size-8 text-primary">
      <path
        fill="currentColor"
        d="M12 1L3 5v6c0 5.55 3.84 10.74 9 12 5.16-1.26 9-6.45 9-12V5l-9-4zm-1 6h2v2h-2V7zm0 4h2v6h-2v-6z"
      />
    </svg>
  </div>
);

export function Hero() {
  const { resolvedTheme } = useTheme();
  const [mounted, setMounted] = useState(false);

  // This is a standard pattern for handling hydration with next-themes
  useEffect(() => {
    setMounted(true); // eslint-disable-line react-hooks/set-state-in-effect
  }, []);

  // Determine particle color based on theme
  const particleColor = mounted
    ? resolvedTheme === "dark"
      ? "#ffffff"
      : "#000000"
    : "#ffffff";

  return (
    <section className="relative min-h-screen w-full flex flex-col items-center justify-center overflow-hidden bg-background">
      {/* Particles Background */}
      <Particles
        className="absolute inset-0 z-0"
        quantity={80}
        staticity={30}
        ease={80}
        color={particleColor}
        size={0.5}
      />

      {/* Retro Grid at bottom */}
      <RetroGrid
        className="z-0"
        angle={65}
        cellSize={60}
        opacity={0.3}
        lightLineColor="hsl(var(--primary) / 0.3)"
        darkLineColor="hsl(var(--primary) / 0.2)"
      />

      {/* Main Content */}
      <div className="relative z-10 flex flex-col items-center justify-center px-4 text-center">
        {/* Orbiting Circles with AWS Icons */}
        <div className="relative flex items-center justify-center mb-8">
          {/* Container with border beam */}
          <div className="relative size-[300px] flex items-center justify-center">
            {/* Border Beam Effect */}
            <div className="absolute inset-0 rounded-full">
              <BorderBeam
                size={80}
                duration={8}
                colorFrom="hsl(var(--primary))"
                colorTo="hsl(var(--chart-3))"
                borderWidth={2}
              />
            </div>

            {/* Inner orbiting circle */}
            <OrbitingCircles
              radius={80}
              duration={25}
              iconSize={40}
              path={true}
              speed={1}
            >
              <S3Icon />
              <EC2Icon />
              <IAMIcon />
            </OrbitingCircles>

            {/* Outer orbiting circle (reverse) */}
            <OrbitingCircles
              radius={130}
              duration={30}
              iconSize={40}
              path={true}
              reverse={true}
              speed={1}
            >
              <LambdaIcon />
              <RDSIcon />
              <DynamoDBIcon />
            </OrbitingCircles>

            {/* Center Shield */}
            <div className="absolute inset-0 flex items-center justify-center">
              <ShieldIcon />
            </div>
          </div>
        </div>

        {/* Logo/Brand */}
        <div className="mb-4">
          <h1 className="text-5xl md:text-7xl font-bold tracking-tight">
            <span className="bg-gradient-to-r from-primary via-chart-3 to-chart-5 bg-clip-text text-transparent">
              CloudCop
            </span>
          </h1>
        </div>

        {/* Animated Word Rotation */}
        <div className="flex items-center justify-center gap-3 text-2xl md:text-4xl font-semibold mb-4">
          <WordRotate
            words={["Scan", "Protect", "Secure", "Monitor"]}
            duration={2000}
            className="text-primary font-bold"
          />
          <span className="text-foreground">Your AWS Infrastructure</span>
        </div>

        {/* Subheadline */}
        <p className="mt-4 max-w-2xl text-lg md:text-xl text-muted-foreground">
          AI-powered cloud security platform that discovers attack paths,
          analyzes vulnerabilities, and generates executable fix commands.
        </p>

        {/* CTA Buttons */}
        <div className="mt-10 flex flex-col sm:flex-row items-center gap-4">
          <Link href="/dashboard">
            <ShimmerButton
              shimmerColor="hsl(var(--primary-foreground))"
              background="hsl(var(--primary))"
              borderRadius="12px"
              shimmerSize="0.1em"
              className="text-base font-semibold px-8 py-4"
            >
              Get Started
            </ShimmerButton>
          </Link>

          <Link
            href="#features"
            className="px-6 py-3 text-muted-foreground hover:text-foreground transition-colors font-medium border border-border rounded-xl hover:bg-muted"
          >
            Learn More
          </Link>
        </div>
      </div>

      {/* Scroll indicator */}
      <div className="absolute bottom-8 left-1/2 -translate-x-1/2 animate-bounce z-10">
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
