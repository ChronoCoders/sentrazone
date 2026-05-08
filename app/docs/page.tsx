import Link from "next/link";

export default function DocsPage() {
  return (
    <>
      <h1>Documentation</h1>
      <p>
        Sentrazone is a self-hosted control plane for managing private VPN infrastructure.
        This documentation covers deployment, configuration, and API usage.
      </p>

      <hr />

      <h2>Quick start</h2>
      <p>
        The fastest path to a running Sentrazone instance is Docker Compose. The stack runs
        on a single server and includes the control plane, reverse proxy, and VPN server management.
      </p>
      <ol>
        <li>Clone the repository and copy the environment file.</li>
        <li>Configure your secrets in <code>.env</code>.</li>
        <li>Run <code>docker compose up -d</code>.</li>
        <li>Access the dashboard at your domain.</li>
      </ol>
      <p>
        For a detailed walkthrough including DNS configuration and firewall rules,
        see the <Link href="/docs/setup">Setup Guide</Link>.
      </p>

      <h2>Architecture</h2>
      <p>
        Sentrazone consists of two binaries: the <strong>control plane</strong> and the <strong>agent</strong>.
      </p>
      <ul>
        <li>
          <strong>Control plane</strong> — Serves the dashboard, stores history in SQLite, and manages
          authentication. Exposes a REST API and a WebSocket endpoint for live metrics.
        </li>
        <li>
          <strong>Agent</strong> — Runs on each VPN node. Collects interface metrics every 10 seconds
          and reports them to the control plane over an authenticated HTTP channel.
        </li>
      </ul>
      <p>
        In the default Docker Compose setup, all components run on a single host. Agents can also
        run on separate machines and report back over the public internet using the agent token.
      </p>

      <h2>Guides</h2>
      <div className="grid sm:grid-cols-2 gap-4 not-prose mt-6">
        {[
          { href: "/docs/setup", title: "Setup Guide", desc: "Full deployment walkthrough from scratch." },
          { href: "/docs/api", title: "API Reference", desc: "REST endpoints for the control plane." },
        ].map((card) => (
          <Link
            key={card.href}
            href={card.href}
            className="block rounded-lg border border-border bg-surface p-5 hover:border-very-muted transition-colors group"
          >
            <h3 className="font-medium text-foreground mb-1 text-sm group-hover:text-gold transition-colors">
              {card.title}
            </h3>
            <p className="text-xs text-muted leading-relaxed">{card.desc}</p>
          </Link>
        ))}
      </div>
    </>
  );
}
