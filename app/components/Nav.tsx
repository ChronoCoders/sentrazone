"use client";

import { useState } from "react";
import Link from "next/link";

export default function Nav() {
  const [open, setOpen] = useState(false);

  return (
    <header className="sticky top-0 z-50 border-b border-border-subtle bg-background/80 backdrop-blur-sm">
      <div className="mx-auto flex h-16 max-w-6xl items-center justify-between px-6">
        <Link href="/" className="flex items-center gap-2.5">
          <span className="text-gold font-semibold tracking-tight text-lg">Sentrazone</span>
        </Link>

        <nav className="hidden md:flex items-center gap-8">
          <Link href="/#features" className="text-sm text-muted hover:text-foreground transition-colors">Features</Link>
          <Link href="/#how-it-works" className="text-sm text-muted hover:text-foreground transition-colors">How it works</Link>
          <Link href="/docs" className="text-sm text-muted hover:text-foreground transition-colors">Docs</Link>
          <Link href="/#pricing" className="text-sm text-muted hover:text-foreground transition-colors">Pricing</Link>
        </nav>

        <div className="hidden md:flex items-center gap-3">
          <a
            href="https://app.sentrazone.com"
            className="text-sm text-muted hover:text-foreground transition-colors"
          >
            Sign in
          </a>
          <a
            href="https://app.sentrazone.com"
            className="rounded-md bg-gold px-4 py-2 text-sm font-medium text-black hover:bg-gold-light transition-colors"
          >
            Get Access
          </a>
        </div>

        <button
          className="md:hidden p-2 text-muted"
          onClick={() => setOpen(!open)}
          aria-label="Toggle menu"
        >
          {open ? (
            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
              <path d="M18 6L6 18M6 6l12 12" />
            </svg>
          ) : (
            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
              <path d="M3 12h18M3 6h18M3 18h18" />
            </svg>
          )}
        </button>
      </div>

      {open && (
        <div className="md:hidden border-t border-border-subtle bg-surface px-6 py-4 flex flex-col gap-4">
          <Link href="/#features" className="text-sm text-muted" onClick={() => setOpen(false)}>Features</Link>
          <Link href="/#how-it-works" className="text-sm text-muted" onClick={() => setOpen(false)}>How it works</Link>
          <Link href="/docs" className="text-sm text-muted" onClick={() => setOpen(false)}>Docs</Link>
          <Link href="/#pricing" className="text-sm text-muted" onClick={() => setOpen(false)}>Pricing</Link>
          <hr className="border-border" />
          <a href="https://app.sentrazone.com" className="text-sm text-muted">Sign in</a>
          <a
            href="https://app.sentrazone.com"
            className="rounded-md bg-gold px-4 py-2 text-sm font-medium text-black text-center"
          >
            Get Access
          </a>
        </div>
      )}
    </header>
  );
}
