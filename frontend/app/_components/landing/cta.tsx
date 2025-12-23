"use client";

import Link from "next/link";
import { Button } from "@/components/ui/moving-border";

export function CTA() {
  return (
    <section className="w-full py-20 md:py-32 px-4 bg-muted/30">
      <div className="max-w-4xl mx-auto text-center">
        <h2 className="text-3xl md:text-5xl font-bold mb-4">
          Secure Your Cloud Today
        </h2>
        <p className="text-lg md:text-xl text-muted-foreground mb-10">
          Start scanning in minutes. Discover vulnerabilities before attackers
          do.
        </p>

        <div className="flex justify-center">
          <Button
            as={Link}
            href="/dashboard"
            containerClassName="h-16 w-56"
            borderClassName="bg-[radial-gradient(var(--primary)_40%,transparent_60%)]"
            className="border-primary/20 bg-primary/10 text-foreground font-semibold text-lg hover:bg-primary/20"
          >
            Get Started
          </Button>
        </div>
      </div>
    </section>
  );
}
