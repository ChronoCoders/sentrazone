import Link from "next/link";

export default function Footer() {
  return (
    <footer className="border-t border-border-subtle">
      <div className="mx-auto max-w-6xl px-6 py-12">
        <div className="flex flex-col md:flex-row items-start justify-between gap-10">
          <div className="flex flex-col gap-3">
            <span className="text-gold font-semibold tracking-tight">Sentrazone</span>
            <p className="text-sm text-muted max-w-xs leading-relaxed">
              Private infrastructure, managed from one place.
            </p>
          </div>

          <div className="grid grid-cols-2 sm:grid-cols-3 gap-8 text-sm">
            <div className="flex flex-col gap-3">
              <span className="text-foreground font-medium">Product</span>
              <Link href="/#features" className="text-muted hover:text-foreground transition-colors">Features</Link>
              <Link href="/#pricing" className="text-muted hover:text-foreground transition-colors">Pricing</Link>
              <Link href="/#how-it-works" className="text-muted hover:text-foreground transition-colors">How it works</Link>
            </div>
            <div className="flex flex-col gap-3">
              <span className="text-foreground font-medium">Developers</span>
              <Link href="/docs" className="text-muted hover:text-foreground transition-colors">Documentation</Link>
              <Link href="/docs/api" className="text-muted hover:text-foreground transition-colors">API Reference</Link>
              <Link href="/docs/setup" className="text-muted hover:text-foreground transition-colors">Setup Guide</Link>
            </div>
            <div className="flex flex-col gap-3">
              <span className="text-foreground font-medium">Company</span>
              <a href="https://github.com/ChronoCoders/sentrazone" target="_blank" rel="noopener noreferrer" className="text-muted hover:text-foreground transition-colors">GitHub</a>
              <a href="https://app.sentrazone.com" className="text-muted hover:text-foreground transition-colors">Dashboard</a>
            </div>
          </div>
        </div>

        <div className="mt-12 pt-6 border-t border-border-subtle flex flex-col sm:flex-row items-center justify-between gap-4">
          <p className="text-xs text-very-muted">© {new Date().getFullYear()} Sentrazone. All rights reserved.</p>
          <p className="text-xs text-very-muted">Built for teams that take access seriously.</p>
        </div>
      </div>
    </footer>
  );
}
