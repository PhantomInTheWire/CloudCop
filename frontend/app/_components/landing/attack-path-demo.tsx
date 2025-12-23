"use client";

import { useRef } from "react";
import { AnimatedBeam } from "@/components/ui/animated-beam";
import { cn } from "@/lib/utils";

const Circle = ({
  className,
  children,
  ref,
}: {
  className?: string;
  children?: React.ReactNode;
  ref?: React.RefObject<HTMLDivElement | null>;
}) => {
  return (
    <div
      ref={ref}
      className={cn(
        "z-10 flex size-12 items-center justify-center rounded-full border-2 bg-background p-3 shadow-[0_0_20px_-12px_rgba(0,0,0,0.8)]",
        className,
      )}
    >
      {children}
    </div>
  );
};

// Icons for the attack path demo
const InternetIcon = () => (
  <svg viewBox="0 0 24 24" className="h-6 w-6 text-red-500">
    <circle cx="12" cy="12" r="10" fill="currentColor" opacity="0.2" />
    <path
      d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm-1 17.93c-3.95-.49-7-3.85-7-7.93 0-.62.08-1.21.21-1.79L9 15v1c0 1.1.9 2 2 2v1.93zm6.9-2.54c-.26-.81-1-1.39-1.9-1.39h-1v-3c0-.55-.45-1-1-1H8v-2h2c.55 0 1-.45 1-1V7h2c1.1 0 2-.9 2-2v-.41c2.93 1.19 5 4.06 5 7.41 0 2.08-.8 3.97-2.1 5.39z"
      fill="currentColor"
    />
  </svg>
);

const EC2Icon = () => (
  <svg viewBox="0 0 24 24" className="h-6 w-6 text-orange-500">
    <rect
      x="3"
      y="3"
      width="18"
      height="18"
      rx="2"
      fill="currentColor"
      opacity="0.2"
    />
    <path
      d="M20 3H4c-1.1 0-2 .9-2 2v14c0 1.1.9 2 2 2h16c1.1 0 2-.9 2-2V5c0-1.1-.9-2-2-2zm-8 12H8v-2h4v2zm0-4H8V9h4v2zm6 4h-4v-2h4v2zm0-4h-4V9h4v2z"
      fill="currentColor"
    />
  </svg>
);

const IAMIcon = () => (
  <svg viewBox="0 0 24 24" className="h-6 w-6 text-yellow-500">
    <path
      d="M12 2C9.24 2 7 4.24 7 7v1H6c-1.1 0-2 .9-2 2v10c0 1.1.9 2 2 2h12c1.1 0 2-.9 2-2V10c0-1.1-.9-2-2-2h-1V7c0-2.76-2.24-5-5-5zm0 2c1.66 0 3 1.34 3 3v1H9V7c0-1.66 1.34-3 3-3z"
      fill="currentColor"
    />
  </svg>
);

const S3Icon = () => (
  <svg viewBox="0 0 24 24" className="h-6 w-6 text-green-500">
    <path
      d="M12 2L2 7l10 5 10-5-10-5zm0 6.5L4.5 6 12 3.5 19.5 6 12 8.5zM2 17l10 5 10-5M2 12l10 5 10-5"
      stroke="currentColor"
      fill="none"
      strokeWidth="2"
    />
  </svg>
);

const DatabaseIcon = () => (
  <svg viewBox="0 0 24 24" className="h-6 w-6 text-blue-500">
    <ellipse cx="12" cy="5" rx="9" ry="3" fill="currentColor" opacity="0.2" />
    <path
      d="M12 2C6.48 2 3 3.79 3 5v14c0 1.21 3.48 3 9 3s9-1.79 9-3V5c0-1.21-3.48-3-9-3zm0 4c4.42 0 7 1.34 7 2s-2.58 2-7 2-7-1.34-7-2 2.58-2 7-2zm7 12c0 .66-2.58 2-7 2s-7-1.34-7-2v-3.5c1.52.9 4.14 1.5 7 1.5s5.48-.6 7-1.5V18zm0-6c0 .66-2.58 2-7 2s-7-1.34-7-2V8.5c1.52.9 4.14 1.5 7 1.5s5.48-.6 7-1.5V12z"
      fill="currentColor"
    />
  </svg>
);

export function AttackPathDemo({ className }: { className?: string }) {
  const containerRef = useRef<HTMLDivElement>(null);
  const internetRef = useRef<HTMLDivElement>(null);
  const ec2Ref = useRef<HTMLDivElement>(null);
  const iamRef = useRef<HTMLDivElement>(null);
  const s3Ref = useRef<HTMLDivElement>(null);
  const dbRef = useRef<HTMLDivElement>(null);

  return (
    <div
      className={cn(
        "relative flex w-full items-center justify-center overflow-hidden p-10",
        className,
      )}
      ref={containerRef}
    >
      <div className="flex size-full max-w-lg flex-row items-stretch justify-between gap-10">
        <div className="flex flex-col justify-center">
          <Circle ref={internetRef} className="border-red-500/50">
            <InternetIcon />
          </Circle>
        </div>
        <div className="flex flex-col justify-center gap-4">
          <Circle ref={ec2Ref} className="border-orange-500/50">
            <EC2Icon />
          </Circle>
          <Circle ref={iamRef} className="border-yellow-500/50">
            <IAMIcon />
          </Circle>
        </div>
        <div className="flex flex-col justify-center gap-4">
          <Circle ref={s3Ref} className="border-green-500/50">
            <S3Icon />
          </Circle>
          <Circle ref={dbRef} className="border-blue-500/50">
            <DatabaseIcon />
          </Circle>
        </div>
      </div>

      {/* Beams showing attack paths */}
      <AnimatedBeam
        containerRef={containerRef}
        fromRef={internetRef}
        toRef={ec2Ref}
        gradientStartColor="#ef4444"
        gradientStopColor="#f97316"
        curvature={-20}
      />
      <AnimatedBeam
        containerRef={containerRef}
        fromRef={internetRef}
        toRef={iamRef}
        gradientStartColor="#ef4444"
        gradientStopColor="#eab308"
        curvature={20}
      />
      <AnimatedBeam
        containerRef={containerRef}
        fromRef={ec2Ref}
        toRef={s3Ref}
        gradientStartColor="#f97316"
        gradientStopColor="#22c55e"
        curvature={-20}
      />
      <AnimatedBeam
        containerRef={containerRef}
        fromRef={iamRef}
        toRef={s3Ref}
        gradientStartColor="#eab308"
        gradientStopColor="#22c55e"
      />
      <AnimatedBeam
        containerRef={containerRef}
        fromRef={iamRef}
        toRef={dbRef}
        gradientStartColor="#eab308"
        gradientStopColor="#3b82f6"
        curvature={20}
      />
    </div>
  );
}
