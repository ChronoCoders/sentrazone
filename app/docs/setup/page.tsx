export default function SetupPage() {
  return (
    <>
      <h1>Setup Guide</h1>
      <p>
        This guide walks through deploying a full Sentrazone stack on a single Linux server
        using Docker Compose with Caddy for TLS termination.
      </p>

      <hr />

      <h2>Prerequisites</h2>
      <ul>
        <li>A Linux server (Ubuntu 22.04+ recommended) with a public IP</li>
        <li>A domain name with DNS managed by Cloudflare or similar</li>
        <li>Docker and Docker Compose installed</li>
        <li>UFW or iptables for firewall management</li>
      </ul>

      <h2>1. Clone the repository</h2>
      <pre><code>{`git clone https://github.com/ChronoCoders/sentrazone.git /root/sentra
cd /root/sentra`}</code></pre>

      <h2>2. Configure environment variables</h2>
      <p>Copy the example file and fill in your values:</p>
      <pre><code>{`cp .env.example .env
chmod 600 .env`}</code></pre>
      <p>Required variables:</p>
      <pre><code>{`SENTRA_JWT_SECRET=<random 64-char string>
SENTRA_AUTH_TOKEN=<random token for agent authentication>
SENTRA_ADMIN_EMAIL=you@example.com
SENTRA_ADMIN_PASSWORD=<strong password>
SENTRA_WG_EASY_PASSWORD=<wg-easy admin password>`}</code></pre>

      <h2>3. Configure your domain</h2>
      <p>
        Create DNS A records pointing your domain and subdomains to your server IP.
        The Caddyfile expects:
      </p>
      <ul>
        <li><code>yourdomain.com</code> → control plane dashboard</li>
        <li><code>wg-la.yourdomain.com</code> → Los Angeles VPN admin panel</li>
        <li><code>wg-va.yourdomain.com</code> → Virginia VPN admin panel</li>
        <li><code>wg-tx.yourdomain.com</code> → Dallas VPN admin panel</li>
      </ul>

      <h2>4. Open firewall ports</h2>
      <pre><code>{`ufw allow 80/tcp
ufw allow 443/tcp
ufw allow 8443/tcp
ufw allow 51820/udp
ufw allow 51830/udp
ufw allow 51832/udp`}</code></pre>

      <h2>5. Start the stack</h2>
      <pre><code>{`docker compose up -d`}</code></pre>
      <p>
        Caddy will automatically obtain TLS certificates. The dashboard will be available
        at your domain within 30–60 seconds.
      </p>

      <h2>6. Log in</h2>
      <p>
        Navigate to your domain and log in with the admin email and password you configured
        in <code>.env</code>. The default session lifetime is 24 hours.
      </p>

      <h2>Updating</h2>
      <p>To deploy a new version:</p>
      <pre><code>{`git pull
docker compose build control
docker compose up -d control`}</code></pre>

      <h2>Agent deployment (remote servers)</h2>
      <p>
        To run the agent on a separate machine rather than sharing a host with the control plane:
      </p>
      <pre><code>{`# On the remote server
SENTRA_CONTROL_URL=https://yourdomain.com \\
SENTRA_AUTH_TOKEN=<your token> \\
SENTRA_SERVER_ID=my-server \\
./sentra-agent`}</code></pre>
      <p>
        The agent binary is built during <code>docker compose build</code> and can be extracted
        from the image or compiled separately with <code>go build ./cmd/agent</code>.
      </p>
    </>
  );
}
