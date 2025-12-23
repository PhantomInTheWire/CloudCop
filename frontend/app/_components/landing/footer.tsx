export function Footer() {
  return (
    <footer className="w-full py-8 px-4 border-t border-border bg-background">
      <div className="max-w-7xl mx-auto flex flex-col md:flex-row items-center justify-between gap-4">
        <div className="flex items-center gap-2">
          <span className="text-xl font-bold bg-gradient-to-r from-primary via-chart-3 to-chart-5 bg-clip-text text-transparent">
            CloudCop
          </span>
        </div>

        <p className="text-sm text-muted-foreground">
          &copy; {new Date().getFullYear()} CloudCop. All rights reserved.
        </p>
      </div>
    </footer>
  );
}
