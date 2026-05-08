function Endpoint({
  method,
  path,
  description,
  auth,
  body,
  response,
}: {
  method: string;
  path: string;
  description: string;
  auth: string;
  body?: string;
  response?: string;
}) {
  const methodColor: Record<string, string> = {
    GET: "text-emerald-400",
    POST: "text-blue-400",
    DELETE: "text-red-400",
  };
  return (
    <div className="not-prose border border-border rounded-lg overflow-hidden mb-6">
      <div className="flex items-center gap-3 px-5 py-3 bg-surface border-b border-border">
        <span className={`font-mono text-xs font-bold ${methodColor[method] ?? "text-gold"}`}>{method}</span>
        <code className="font-mono text-sm text-foreground">{path}</code>
        <span className="ml-auto text-xs text-very-muted">{auth}</span>
      </div>
      <div className="px-5 py-4">
        <p className="text-sm text-muted mb-3">{description}</p>
        {body && (
          <>
            <p className="text-xs text-very-muted uppercase tracking-wide mb-2">Request body</p>
            <pre className="bg-surface-2 rounded border border-border p-3 text-xs font-mono text-foreground overflow-x-auto mb-3">
              <code>{body}</code>
            </pre>
          </>
        )}
        {response && (
          <>
            <p className="text-xs text-very-muted uppercase tracking-wide mb-2">Response</p>
            <pre className="bg-surface-2 rounded border border-border p-3 text-xs font-mono text-foreground overflow-x-auto">
              <code>{response}</code>
            </pre>
          </>
        )}
      </div>
    </div>
  );
}

export default function ApiPage() {
  return (
    <>
      <h1>API Reference</h1>
      <p>
        All endpoints are served by the control plane at your base URL.
        Authenticated endpoints require a Bearer JWT token obtained from <code>/api/login</code>.
      </p>

      <hr />

      <h2>Authentication</h2>

      <Endpoint
        method="POST"
        path="/api/login"
        description="Exchange credentials for a JWT. The token expires after 24 hours."
        auth="Public"
        body={`{
  "email": "you@example.com",
  "password": "your-password"
}`}
        response={`{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}`}
      />

      <h2>Server Status</h2>

      <Endpoint
        method="GET"
        path="/api/statuses"
        description="Returns the latest status snapshot for all registered servers."
        auth="Bearer token"
        response={`[
  {
    "server_id": "los-angeles",
    "online": true,
    "ping_ms": 12.4,
    "peers": 5,
    "transfer_rx": 1048576,
    "transfer_tx": 524288,
    "reported_at": "2026-05-06T12:00:00Z"
  }
]`}
      />

      <Endpoint
        method="GET"
        path="/api/history"
        description="Returns latency and transfer history for all servers over the last 72 hours."
        auth="Bearer token"
        response={`[
  {
    "server_id": "los-angeles",
    "ping_ms": 12.4,
    "transfer_rx": 1048576,
    "transfer_tx": 524288,
    "recorded_at": "2026-05-06T12:00:00Z"
  }
]`}
      />

      <h2>Peer Management</h2>

      <Endpoint
        method="GET"
        path="/api/wg-clients/{serverID}"
        description="List all peers configured on a VPN server."
        auth="Bearer token (admin)"
        response={`[
  {
    "id": "abc123",
    "name": "alice-iphone",
    "enabled": true,
    "address": "10.8.0.2",
    "publicKey": "...",
    "latestHandshakeAt": "2026-05-06T11:55:00Z",
    "transferRx": 2097152,
    "transferTx": 1048576
  }
]`}
      />

      <Endpoint
        method="POST"
        path="/api/wg-clients/{serverID}"
        description="Create a new peer on a VPN server. Returns the new peer object."
        auth="Bearer token (admin)"
        body={`{
  "name": "alice-iphone"
}`}
        response={`{
  "id": "abc123",
  "name": "alice-iphone",
  "enabled": true,
  "address": "10.8.0.3",
  "publicKey": "..."
}`}
      />

      <Endpoint
        method="POST"
        path="/api/wg-clients/{serverID}/{clientID}/{action}"
        description={`Enable or disable a peer. action must be "enable" or "disable".`}
        auth="Bearer token (admin)"
      />

      <Endpoint
        method="GET"
        path="/api/wg-clients/{serverID}/{clientID}/qrcode"
        description="Returns the peer's WireGuard config as an SVG QR code for mobile scanning."
        auth="Bearer token (admin)"
        response="image/svg+xml"
      />

      <h2>Audit Log</h2>

      <Endpoint
        method="GET"
        path="/api/audit"
        description="Returns the audit log. Supports ?limit= and ?offset= query parameters."
        auth="Bearer token (admin)"
        response={`[
  {
    "id": 1,
    "actor": "you@example.com",
    "action": "wg_client_created",
    "target": "los-angeles/alice-iphone",
    "ip": "1.2.3.4",
    "created_at": "2026-05-06T12:01:00Z"
  }
]`}
      />

      <Endpoint
        method="GET"
        path="/api/audit/export"
        description="Export the full audit log as a CSV file."
        auth="Bearer token (admin)"
        response="text/csv"
      />

      <h2>Agent Reporting</h2>

      <Endpoint
        method="POST"
        path="/api/report"
        description="Used by agent binaries to push metrics to the control plane. Requires the agent auth token."
        auth="Bearer agent-token"
        body={`{
  "server_id": "los-angeles",
  "ping_ms": 12.4,
  "peers": 5,
  "transfer_rx": 1048576,
  "transfer_tx": 524288
}`}
      />
    </>
  );
}
