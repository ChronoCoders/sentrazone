package models

import "time"

type Organization struct {
	ID        string    `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type User struct {
	ID        string    `json:"id" db:"id"`
	OrgID     string    `json:"org_id" db:"org_id"`
	Email     string    `json:"email" db:"email"`
	Name      string    `json:"name" db:"name"`
	Password  string    `json:"-" db:"password"`
	Role      string    `json:"role" db:"role"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type Server struct {
	ID        string    `json:"id" db:"id"`
	OrgID     string    `json:"org_id" db:"org_id"`
	Hostname  string    `json:"hostname" db:"hostname"`
	PublicKey string    `json:"public_key" db:"public_key"`
	Endpoint  string    `json:"endpoint" db:"endpoint"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type Peer struct {
	PublicKey       string    `json:"public_key"`
	Endpoint        string    `json:"endpoint"`
	AllowedIPs      []string  `json:"allowed_ips"`
	LatestHandshake time.Time `json:"latest_handshake"`
	ReceiveBytes    int64     `json:"receive_bytes"`
	TransmitBytes   int64     `json:"transmit_bytes"`
	KeepAlive       int       `json:"persistent_keepalive"`
}

type SystemInfo struct {
	Hostname      string  `json:"hostname"`
	OS            string  `json:"os"`
	Arch          string  `json:"arch"`
	KernelVersion string  `json:"kernel_version"`
	Platform      string  `json:"platform"`
	CPUCount      int     `json:"cpu_count"`
	CPUPercent    float64 `json:"cpu_percent"`
	MemoryTotal   uint64  `json:"memory_total"`
	MemoryUsed    uint64  `json:"memory_used"`
	MemoryPercent float64 `json:"memory_percent"`
	DiskTotal     uint64  `json:"disk_total"`
	DiskUsed      uint64  `json:"disk_used"`
	DiskPercent   float64 `json:"disk_percent"`
	LoadAverage   float64 `json:"load_average"`
	Uptime        uint64  `json:"uptime"`
	NetBytesSent  uint64  `json:"net_bytes_sent"`
	NetBytesRecv  uint64  `json:"net_bytes_recv"`
	PingLatencyMs float64 `json:"ping_latency_ms"`
}

type Status struct {
	Interface  string     `json:"interface"`
	PublicKey  string     `json:"public_key"`
	ListenPort int        `json:"listen_port"`
	Peers      []Peer     `json:"peers"`
	System     SystemInfo `json:"system"`
}

type HistoryPoint struct {
	Time          time.Time `json:"time"`
	CPUPercent    float64   `json:"cpu_percent"`
	MemoryPercent float64   `json:"memory_percent"`
	DiskPercent   float64   `json:"disk_percent"`
	LoadAverage   float64   `json:"load_average"`
	NetBytesSent  uint64    `json:"net_bytes_sent"`
	NetBytesRecv  uint64    `json:"net_bytes_recv"`
	PeerCount     int       `json:"peer_count"`
	PingLatencyMs float64   `json:"ping_latency_ms"`
}

type PeerHistoryPoint struct {
	Time          time.Time `json:"time"`
	ReceiveBytes  int64     `json:"receive_bytes"`
	TransmitBytes int64     `json:"transmit_bytes"`
}

type AuditLog struct {
	ID        int64     `json:"id"`
	UserEmail string    `json:"user_email"`
	Action    string    `json:"action"`
	Details   string    `json:"details"`
	IP        string    `json:"ip"`
	CreatedAt time.Time `json:"created_at"`
}

type PeerMeta struct {
	Label      string     `json:"label"`
	Notes      string     `json:"notes"`
	ExpiresAt  *time.Time `json:"expires_at"`
	QuotaBytes int64      `json:"quota_bytes"`
}

type ServerSummary struct {
	ServerID    string  `json:"server_id"`
	UptimePC    float64 `json:"uptime_pc"`
	ActivePeers int     `json:"active_peers"`
	RxBytes     int64   `json:"rx_bytes"`
	TxBytes     int64   `json:"tx_bytes"`
}
