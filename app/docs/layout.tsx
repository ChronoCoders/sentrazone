import Link from "next/link";

const nav = [
  {
    group: "Getting Started",
    links: [
      { href: "/docs", label: "Overview" },
      { href: "/docs/setup", label: "Setup Guide" },
    ],
  },
  {
    group: "Reference",
    links: [
      { href: "/docs/api", label: "API Reference" },
    ],
  },
];

export default function DocsLayout({ children }: { children: React.ReactNode }) {
  return (
    <div className="mx-auto max-w-6xl px-6 py-12">
      <div className="flex gap-12">
        {/* Sidebar */}
        <aside className="hidden md:flex flex-col gap-8 w-52 shrink-0 sticky top-28 self-start">
          {nav.map((section) => (
            <div key={section.group}>
              <p className="text-xs font-medium text-very-muted uppercase tracking-widest mb-3">
                {section.group}
              </p>
              <ul className="flex flex-col gap-1">
                {section.links.map((link) => (
                  <li key={link.href}>
                    <Link
                      href={link.href}
                      className="block text-sm text-muted hover:text-foreground transition-colors py-1"
                    >
                      {link.label}
                    </Link>
                  </li>
                ))}
              </ul>
            </div>
          ))}
        </aside>

        {/* Content */}
        <article className="flex-1 min-w-0 prose prose-invert prose-sm max-w-none
          prose-headings:font-semibold prose-headings:tracking-tight
          prose-h1:text-3xl prose-h1:mb-8
          prose-h2:text-xl prose-h2:mt-12 prose-h2:mb-4
          prose-h3:text-base prose-h3:mt-8 prose-h3:mb-3
          prose-p:text-muted prose-p:leading-relaxed
          prose-code:text-gold prose-code:bg-surface prose-code:px-1.5 prose-code:py-0.5 prose-code:rounded prose-code:text-xs prose-code:font-mono
          prose-pre:bg-surface prose-pre:border prose-pre:border-border prose-pre:rounded-lg
          prose-li:text-muted
          prose-a:text-gold prose-a:no-underline hover:prose-a:text-gold-light
          prose-hr:border-border-subtle">
          {children}
        </article>
      </div>
    </div>
  );
}
