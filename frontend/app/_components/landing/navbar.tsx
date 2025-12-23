"use client";

import Link from "next/link";
import { useState, useEffect } from "react";
import { motion, AnimatePresence } from "motion/react";
import { cn } from "@/lib/utils";
import { ThemeToggle } from "@/components/theme-toggle";

const navItems = [
  { name: "Features", link: "#features" },
  { name: "Dashboard", link: "/dashboard" },
];

export function Navbar() {
  const [visible, setVisible] = useState(true);
  const [lastScrollY, setLastScrollY] = useState(0);
  const [atTop, setAtTop] = useState(true);

  useEffect(() => {
    const handleScroll = () => {
      const currentScrollY = window.scrollY;

      // Check if at top of page
      if (currentScrollY < 50) {
        setAtTop(true);
        setVisible(true);
      } else {
        setAtTop(false);
        // Show navbar when scrolling up, hide when scrolling down
        if (currentScrollY < lastScrollY) {
          setVisible(true);
        } else {
          setVisible(false);
        }
      }

      setLastScrollY(currentScrollY);
    };

    window.addEventListener("scroll", handleScroll, { passive: true });
    return () => window.removeEventListener("scroll", handleScroll);
  }, [lastScrollY]);

  return (
    <AnimatePresence mode="wait">
      <motion.nav
        initial={{ opacity: 1, y: 0 }}
        animate={{
          y: visible ? 0 : -100,
          opacity: visible ? 1 : 0,
        }}
        transition={{ duration: 0.2 }}
        className={cn(
          "fixed top-4 inset-x-0 mx-auto z-[5000] flex max-w-fit items-center justify-center gap-2 rounded-full px-4 py-2",
          atTop
            ? "bg-transparent"
            : "border border-border bg-background/80 backdrop-blur-lg shadow-lg",
        )}
      >
        {/* Logo */}
        <Link
          href="/"
          className="px-3 py-1.5 text-lg font-bold bg-gradient-to-r from-primary via-chart-3 to-chart-5 bg-clip-text text-transparent"
        >
          CloudCop
        </Link>

        {/* Nav Items */}
        <div className="flex items-center gap-1">
          {navItems.map((item, idx) => (
            <Link
              key={idx}
              href={item.link}
              className="px-3 py-1.5 text-sm text-muted-foreground hover:text-foreground transition-colors rounded-full hover:bg-muted"
            >
              {item.name}
            </Link>
          ))}
        </div>

        {/* Theme Toggle */}
        <div className="ml-2">
          <ThemeToggle />
        </div>

        {/* Dashboard Button */}
        <Link
          href="/dashboard"
          className="ml-2 px-4 py-1.5 text-sm font-medium rounded-full bg-primary text-primary-foreground hover:bg-primary/90 transition-colors"
        >
          Get Started
        </Link>
      </motion.nav>
    </AnimatePresence>
  );
}
