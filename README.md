# Sentrazone

Sentrazone is a lightweight WireGuard VPN control plane with a real-time dashboard. It monitors multiple WireGuard servers, tracks system metrics, manages VPN peers, and routes traffic through residential proxies per server.

## Architecture

- **Control Plane** (`cmd/control`) — central API server, SQLite storage, WebSocket hub, JWT auth
- **Agent** (`cmd/agent`) — runs inside each WireGuard container, reports metrics every 10s
- **Event Bus** — in-memory pub/sub decoupling agents from the control plane
- **Status Cache** — serves live status to the dashboard without blocking on WireGuard calls

## Features

- Real-time dashboard with WebSocket live updates
- System metrics — CPU, memory, disk, load, ping latency, bandwidth
- WireGuard peer management via wg-easy API (create peers, enable/disable, QR codes)
- History recording with charts and CSV export
- Email alerting (offline detection, daily/weekly summaries via SMTP)
- Audit log with per-user action tracking
- Multi-user with role-based access (admin / viewer)
- Caddy reverse proxy with automatic Let's Encrypt TLS
- Subdomain routing for wg-easy management panels

## Getting Started

### Prerequisites

- Go 1.22+
- Docker & Docker Compose
- WireGuard kernel module on the host

### Installation

```bash
git clone https://github.com/ChronoCoders/sentrazone.git
cd sentrazone
cp .env.example .env  # fill in your values
docker compose up -d --build
```

### Configuration

Key environment variables (set in `.env`):

| Variable | Description |
|---|---|
| `SENTRA_JWT_SECRET` | Secret for JWT signing |
| `SENTRA_AUTH_TOKEN` | Shared token for agent → control auth |
| `SENTRA_ADMIN_EMAIL` | Initial admin account email |
| `SENTRA_ADMIN_PASSWORD` | Initial admin account password |
| `SENTRA_DB` | Path to SQLite database file |
| `SENTRA_SMTP_HOST` | SMTP host for email alerts |
| `SENTRA_ALERT_EMAIL` | Recipient address for alerts |
| `SENTRA_WG_EASY_PASSWORD` | Password for wg-easy API access |

## License

MIT
