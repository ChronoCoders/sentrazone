import Link from "next/link";
import TrafficFlow from "./components/TrafficFlow";

const features = [
  {
    title: "Multi-Server Management",
    description: "Monitor every VPN node from a single dashboard. Status, bandwidth, and peer counts at a glance.",
    icon: "◈",
  },
  {
    title: "Real-Time Metrics",
    description: "Live latency, transfer rates, and connection status pushed over WebSocket — no polling.",
    icon: "◎",
  },
  {
    title: "Peer Lifecycle",
    description: "Provision, enable, and disable peers in one click. QR codes for instant mobile onboarding.",
    icon: "◇",
  },
  {
    title: "Residential Exit Routing",
    description: "Route traffic through city-specific residential exit nodes for authentic, location-accurate connections.",
    icon: "⊕",
  },
  {
    title: "Immutable Audit Trail",
    description: "Every administrative action is logged with timestamp, actor, and target. Export to CSV.",
    icon: "≡",
  },
  {
    title: "Encrypted Tunneling",
    description: "All traffic encrypted end-to-end. No plaintext, no exceptions. Obfuscated transport available.",
    icon: "⬡",
  },
];

const steps = [
  {
    number: "01",
    title: "Connect your servers",
    description: "Deploy the agent binary on each VPN node. It reports metrics back to your control plane over an authenticated channel.",
  },
  {
    number: "02",
    title: "Distribute access",
    description: "Create peer configurations directly from the dashboard. Generate QR codes for instant client setup on any device.",
  },
  {
    number: "03",
    title: "Monitor in real time",
    description: "Watch bandwidth, latency, and handshake status across all nodes from a single live view. Get alerted when anything goes offline.",
  },
];

const plans = [
  {
    name: "Solo",
    price: "$29",
    period: "/month",
    description: "For individuals managing their own infrastructure.",
    features: [
      "1 admin user",
      "Up to 3 VPN servers",
      "25 peers per server",
      "7-day audit history",
      "QR code provisioning",
    ],
    cta: "Get started",
    href: "https://app.sentrazone.com",
    highlight: false,
  },
  {
    name: "Team",
    price: "$79",
    period: "/month",
    description: "For teams with shared infrastructure and multiple admins.",
    features: [
      "5 admin users",
      "Unlimited VPN servers",
      "Unlimited peers",
      "90-day audit history",
      "CSV export",
      "Priority support",
    ],
    cta: "Get started",
    href: "https://app.sentrazone.com",
    highlight: true,
  },
  {
    name: "Enterprise",
    price: "Custom",
    period: "",
    description: "For organizations with advanced compliance and scale requirements.",
    features: [
      "Unlimited users",
      "Unlimited servers",
      "Unlimited peers",
      "1-year audit retention",
      "SSO / SAML",
      "Dedicated support",
    ],
    cta: "Contact us",
    href: "mailto:contact@chronocoder.dev",
    highlight: false,
  },
];

export default function Home() {
  return (
    <>
      {/* Hero */}
      <section className="relative overflow-hidden">
        {/* Dot grid background */}
        <div
          className="absolute inset-0 pointer-events-none"
          style={{
            backgroundImage: "radial-gradient(circle, #1f1f1f 1px, transparent 1px)",
            backgroundSize: "28px 28px",
          }}
        />
        {/* Radial fade */}
        <div
          className="absolute inset-0 pointer-events-none"
          style={{
            background: "radial-gradient(ellipse 80% 60% at 50% 0%, transparent 40%, #080808 100%)",
          }}
        />

        <div className="relative mx-auto max-w-6xl px-6 pt-24 pb-20">
          {/* Badge */}
          <div className="flex justify-center mb-8">
            <span className="inline-flex items-center gap-2 rounded-full border border-gold-dim/40 bg-gold/5 px-4 py-1.5 text-xs text-gold tracking-wide font-medium">
              <span className="inline-block h-1.5 w-1.5 rounded-full bg-gold" />
              Private infrastructure, managed
            </span>
          </div>

          {/* Headline */}
          <h1 className="text-center text-5xl sm:text-6xl lg:text-7xl font-semibold tracking-tight leading-[1.05] mb-6 animate-fade-up">
            Control your network.
            <br />
            <span className="text-muted">Own your access.</span>
          </h1>

          <p className="mx-auto max-w-xl text-center text-base sm:text-lg text-muted leading-relaxed mb-10 animate-fade-up-delay-1">
            Sentrazone is a unified control plane for your private VPN infrastructure.
            Monitor nodes, manage peers, and route through residential exit points — all from one dashboard.
          </p>

          {/* CTAs */}
          <div className="flex flex-col sm:flex-row items-center justify-center gap-3 mb-20 animate-fade-up-delay-2">
            <Link
              href="https://app.sentrazone.com"
              className="rounded-md bg-gold px-6 py-3 text-sm font-medium text-black hover:bg-gold-light transition-colors"
            >
              Get Access
            </Link>
            <Link
              href="/docs"
              className="rounded-md border border-border px-6 py-3 text-sm text-muted hover:text-foreground hover:border-very-muted transition-colors"
            >
              View Documentation
            </Link>
          </div>

          {/* Traffic Flow Diagram */}
          <div className="rounded-xl border border-border bg-surface/60 p-6 sm:p-8 animate-fade-up-delay-3">
            <p className="text-xs text-very-muted uppercase tracking-widest mb-6 text-center font-medium">
              How your traffic moves
            </p>
            <TrafficFlow />
          </div>
        </div>
      </section>

      {/* Features */}
      <section id="features" className="py-24 border-t border-border-subtle">
        <div className="mx-auto max-w-6xl px-6">
          <div className="mb-14">
            <p className="text-xs text-gold uppercase tracking-widest font-medium mb-3">Features</p>
            <h2 className="text-3xl sm:text-4xl font-semibold tracking-tight mb-4">
              Everything you need to manage
              <br />
              private access at scale.
            </h2>
            <p className="text-muted max-w-lg leading-relaxed">
              Built for teams that take network access seriously. No dashboards built on top of dashboards — just the primitives you need.
            </p>
          </div>

          <div className="grid sm:grid-cols-2 lg:grid-cols-3 gap-px bg-border-subtle">
            {features.map((f) => (
              <div key={f.title} className="bg-background p-8">
                <span className="text-gold text-2xl mb-5 block">{f.icon}</span>
                <h3 className="font-medium text-foreground mb-2">{f.title}</h3>
                <p className="text-sm text-muted leading-relaxed">{f.description}</p>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* How it works */}
      <section id="how-it-works" className="py-24 border-t border-border-subtle">
        <div className="mx-auto max-w-6xl px-6">
          <div className="mb-14">
            <p className="text-xs text-gold uppercase tracking-widest font-medium mb-3">How it works</p>
            <h2 className="text-3xl sm:text-4xl font-semibold tracking-tight">
              Up in minutes, not weeks.
            </h2>
          </div>

          <div className="grid sm:grid-cols-3 gap-8">
            {steps.map((step) => (
              <div key={step.number} className="flex flex-col gap-4">
                <span className="font-mono text-4xl font-bold text-gold-dim/50">{step.number}</span>
                <h3 className="font-semibold text-foreground">{step.title}</h3>
                <p className="text-sm text-muted leading-relaxed">{step.description}</p>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* Pricing */}
      <section id="pricing" className="py-24 border-t border-border-subtle">
        <div className="mx-auto max-w-6xl px-6">
          <div className="mb-14">
            <p className="text-xs text-gold uppercase tracking-widest font-medium mb-3">Pricing</p>
            <h2 className="text-3xl sm:text-4xl font-semibold tracking-tight mb-4">Simple, predictable pricing.</h2>
            <p className="text-muted">No usage fees. No seat surprises. Pay once, manage everything.</p>
          </div>

          <div className="grid sm:grid-cols-3 gap-6">
            {plans.map((plan) => (
              <div
                key={plan.name}
                className={`rounded-xl border p-8 flex flex-col gap-6 ${
                  plan.highlight
                    ? "border-gold/30 bg-gold/5"
                    : "border-border bg-surface"
                }`}
              >
                {plan.highlight && (
                  <span className="self-start rounded-full bg-gold/10 border border-gold/20 px-3 py-1 text-xs text-gold font-medium">
                    Most popular
                  </span>
                )}
                <div>
                  <h3 className="font-semibold text-foreground mb-1">{plan.name}</h3>
                  <p className="text-sm text-muted leading-relaxed">{plan.description}</p>
                </div>
                <div className="flex items-baseline gap-1">
                  <span className="text-4xl font-bold tracking-tight">{plan.price}</span>
                  <span className="text-muted text-sm">{plan.period}</span>
                </div>
                <ul className="flex flex-col gap-2.5">
                  {plan.features.map((f) => (
                    <li key={f} className="flex items-center gap-2.5 text-sm text-muted">
                      <svg width="14" height="14" viewBox="0 0 14 14" fill="none">
                        <circle cx="7" cy="7" r="6" stroke="#C9A84C" strokeWidth="1" opacity="0.5" />
                        <path d="M4.5 7L6.5 9L9.5 5" stroke="#C9A84C" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round" />
                      </svg>
                      {f}
                    </li>
                  ))}
                </ul>
                <Link
                  href={plan.href}
                  className={`mt-auto rounded-md px-5 py-2.5 text-sm font-medium text-center transition-colors ${
                    plan.highlight
                      ? "bg-gold text-black hover:bg-gold-light"
                      : "border border-border text-muted hover:text-foreground hover:border-very-muted"
                  }`}
                >
                  {plan.cta}
                </Link>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* CTA strip */}
      <section className="py-24 border-t border-border-subtle">
        <div className="mx-auto max-w-6xl px-6 text-center">
          <h2 className="text-3xl sm:text-4xl font-semibold tracking-tight mb-4">
            Ready to take control?
          </h2>
          <p className="text-muted mb-10 max-w-md mx-auto leading-relaxed">
            Deploy your own private infrastructure and manage it from a single, secure dashboard.
          </p>
          <div className="flex flex-col sm:flex-row gap-3 justify-center">
            <Link
              href="https://app.sentrazone.com"
              className="rounded-md bg-gold px-6 py-3 text-sm font-medium text-black hover:bg-gold-light transition-colors"
            >
              Get Access
            </Link>
            <Link
              href="/docs"
              className="rounded-md border border-border px-6 py-3 text-sm text-muted hover:text-foreground transition-colors"
            >
              Read the docs
            </Link>
          </div>
        </div>
      </section>
    </>
  );
}
