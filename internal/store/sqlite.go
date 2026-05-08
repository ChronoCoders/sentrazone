package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/ChronoCoders/sentra/internal/models"
	"golang.org/x/crypto/bcrypt"
	_ "modernc.org/sqlite"
)

type Store struct {
	db *sql.DB
}

func New(path, adminEmail, adminPassword string) (*Store, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("failed to open db: %w", err)
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping db: %w", err)
	}
	if err := initSchema(db); err != nil {
		return nil, fmt.Errorf("failed to init schema: %w", err)
	}
	runMigrations(db)

	if adminEmail != "" && adminPassword != "" {
		var count int
		if db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count) == nil && count == 0 {
			hash, err := bcrypt.GenerateFromPassword([]byte(adminPassword), bcrypt.DefaultCost)
			if err == nil {
				_, _ = db.Exec(`INSERT OR IGNORE INTO organizations (id, name) VALUES ('org1', 'Default Org')`)
				_, _ = db.Exec(`INSERT OR IGNORE INTO users (id, org_id, email, name, role, password) VALUES ('admin', 'org1', ?, 'Admin', 'admin', ?)`, adminEmail, string(hash))
			}
		}
	}

	return &Store{db: db}, nil
}

func initSchema(db *sql.DB) error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS organizations (id TEXT PRIMARY KEY, name TEXT NOT NULL, created_at DATETIME DEFAULT CURRENT_TIMESTAMP)`,
		`CREATE TABLE IF NOT EXISTS users (id TEXT PRIMARY KEY, org_id TEXT NOT NULL, email TEXT NOT NULL UNIQUE, name TEXT, role TEXT DEFAULT 'viewer', password TEXT, created_at DATETIME DEFAULT CURRENT_TIMESTAMP, FOREIGN KEY(org_id) REFERENCES organizations(id))`,
		`CREATE TABLE IF NOT EXISTS servers (id TEXT PRIMARY KEY, org_id TEXT NOT NULL, hostname TEXT, public_key TEXT, endpoint TEXT, created_at DATETIME DEFAULT CURRENT_TIMESTAMP, FOREIGN KEY(org_id) REFERENCES organizations(id))`,
		`CREATE TABLE IF NOT EXISTS peers (public_key TEXT PRIMARY KEY, endpoint TEXT, allowed_ips TEXT, latest_handshake DATETIME, receive_bytes INTEGER, transmit_bytes INTEGER)`,
		`CREATE TABLE IF NOT EXISTS status_history (id INTEGER PRIMARY KEY AUTOINCREMENT, server_id TEXT NOT NULL, cpu_percent REAL, memory_percent REAL, disk_percent REAL, load_average REAL, net_bytes_sent INTEGER, net_bytes_recv INTEGER, peer_count INTEGER, ping_latency_ms REAL DEFAULT 0, recorded_at DATETIME DEFAULT CURRENT_TIMESTAMP)`,
		`CREATE INDEX IF NOT EXISTS idx_history_server_time ON status_history(server_id, recorded_at)`,
		`CREATE TABLE IF NOT EXISTS peer_labels (public_key TEXT PRIMARY KEY, label TEXT NOT NULL DEFAULT '', notes TEXT NOT NULL DEFAULT '', expires_at DATETIME DEFAULT NULL, quota_bytes INTEGER NOT NULL DEFAULT 0, updated_at DATETIME DEFAULT CURRENT_TIMESTAMP)`,
		`CREATE TABLE IF NOT EXISTS audit_log (id INTEGER PRIMARY KEY AUTOINCREMENT, user_email TEXT NOT NULL DEFAULT '', action TEXT NOT NULL, details TEXT NOT NULL DEFAULT '', ip TEXT NOT NULL DEFAULT '', created_at DATETIME DEFAULT CURRENT_TIMESTAMP)`,
		`CREATE INDEX IF NOT EXISTS idx_audit_time ON audit_log(created_at)`,
		`CREATE TABLE IF NOT EXISTS peer_history (id INTEGER PRIMARY KEY AUTOINCREMENT, server_id TEXT NOT NULL, public_key TEXT NOT NULL, receive_bytes INTEGER NOT NULL DEFAULT 0, transmit_bytes INTEGER NOT NULL DEFAULT 0, recorded_at DATETIME DEFAULT CURRENT_TIMESTAMP)`,
		`CREATE INDEX IF NOT EXISTS idx_peer_hist ON peer_history(server_id, public_key, recorded_at)`,
	}
	for _, q := range queries {
		if _, err := db.Exec(q); err != nil {
			return err
		}
	}
	return nil
}

func runMigrations(db *sql.DB) {
	_, _ = db.Exec("ALTER TABLE users ADD COLUMN role TEXT DEFAULT 'viewer'")
	_, _ = db.Exec("ALTER TABLE users ADD COLUMN password TEXT DEFAULT ''")
	_, _ = db.Exec("ALTER TABLE status_history ADD COLUMN ping_latency_ms REAL DEFAULT 0")
	_, _ = db.Exec("ALTER TABLE peer_labels ADD COLUMN notes TEXT NOT NULL DEFAULT ''")
	_, _ = db.Exec("ALTER TABLE peer_labels ADD COLUMN expires_at DATETIME DEFAULT NULL")
	_, _ = db.Exec("ALTER TABLE peer_labels ADD COLUMN quota_bytes INTEGER NOT NULL DEFAULT 0")
}

func (s *Store) CreateUser(ctx context.Context, u *models.User) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO users (id, org_id, email, name, role, password, created_at) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		u.ID, u.OrgID, u.Email, u.Name, u.Role, u.Password, u.CreatedAt)
	return err
}

func (s *Store) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	return s.scanUser(s.db.QueryRowContext(ctx,
		`SELECT id, org_id, email, name, role, password, created_at FROM users WHERE email = ?`, email))
}

func (s *Store) GetUserByID(ctx context.Context, id string) (*models.User, error) {
	return s.scanUser(s.db.QueryRowContext(ctx,
		`SELECT id, org_id, email, name, role, password, created_at FROM users WHERE id = ?`, id))
}

func (s *Store) scanUser(row *sql.Row) (*models.User, error) {
	u := &models.User{}
	if err := row.Scan(&u.ID, &u.OrgID, &u.Email, &u.Name, &u.Role, &u.Password, &u.CreatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return u, nil
}

func (s *Store) ListUsers(ctx context.Context) ([]models.User, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, org_id, email, name, role, created_at FROM users ORDER BY created_at`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var users []models.User
	for rows.Next() {
		var u models.User
		if err := rows.Scan(&u.ID, &u.OrgID, &u.Email, &u.Name, &u.Role, &u.CreatedAt); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, rows.Err()
}

func (s *Store) UpdateUserPassword(ctx context.Context, id, hashedPassword string) error {
	_, err := s.db.ExecContext(ctx, `UPDATE users SET password = ? WHERE id = ?`, hashedPassword, id)
	return err
}

func (s *Store) DeleteUser(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM users WHERE id = ?`, id)
	return err
}

func (s *Store) CountAdmins(ctx context.Context) (int, error) {
	var n int
	err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM users WHERE role = 'admin'`).Scan(&n)
	return n, err
}

func (s *Store) InsertHistory(ctx context.Context, serverID string, status *models.Status) error {
	sys := status.System
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO status_history (server_id, cpu_percent, memory_percent, disk_percent, load_average, net_bytes_sent, net_bytes_recv, peer_count, ping_latency_ms, recorded_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		serverID, sys.CPUPercent, sys.MemoryPercent, sys.DiskPercent, sys.LoadAverage,
		sys.NetBytesSent, sys.NetBytesRecv, len(status.Peers), sys.PingLatencyMs, time.Now(),
	)
	if err != nil {
		return err
	}
	_, _ = s.db.ExecContext(ctx,
		`DELETE FROM status_history WHERE server_id = ? AND recorded_at < datetime('now', '-72 hours')`, serverID)
	return nil
}

func (s *Store) GetHistory(ctx context.Context, serverID string, hours int) ([]models.HistoryPoint, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT recorded_at, cpu_percent, memory_percent, disk_percent, load_average, net_bytes_sent, net_bytes_recv, peer_count, ping_latency_ms FROM status_history WHERE server_id = ? AND recorded_at > datetime('now', ? || ' hours') ORDER BY recorded_at ASC LIMIT 1000`,
		serverID, fmt.Sprintf("-%d", hours),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var points []models.HistoryPoint
	for rows.Next() {
		var p models.HistoryPoint
		if err := rows.Scan(&p.Time, &p.CPUPercent, &p.MemoryPercent, &p.DiskPercent, &p.LoadAverage, &p.NetBytesSent, &p.NetBytesRecv, &p.PeerCount, &p.PingLatencyMs); err != nil {
			return nil, err
		}
		points = append(points, p)
	}
	return points, rows.Err()
}

func (s *Store) GetUptimePercent(ctx context.Context, serverID string, hours int) (float64, error) {
	var count int
	err := s.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM status_history WHERE server_id = ? AND recorded_at > datetime('now', ? || ' hours')`,
		serverID, fmt.Sprintf("-%d", hours),
	).Scan(&count)
	if err != nil {
		return 0, err
	}
	expected := hours * 360
	if expected == 0 {
		return 100, nil
	}
	pct := float64(count) / float64(expected) * 100
	if pct > 100 {
		pct = 100
	}
	return pct, nil
}

func (s *Store) InsertPeerHistory(ctx context.Context, serverID, publicKey string, rxBytes, txBytes int64) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO peer_history (server_id, public_key, receive_bytes, transmit_bytes, recorded_at) VALUES (?, ?, ?, ?, ?)`,
		serverID, publicKey, rxBytes, txBytes, time.Now(),
	)
	if err != nil {
		return err
	}
	_, _ = s.db.ExecContext(ctx,
		`DELETE FROM peer_history WHERE server_id = ? AND public_key = ? AND recorded_at < datetime('now', '-24 hours')`,
		serverID, publicKey,
	)
	return nil
}

func (s *Store) GetPeerHistory(ctx context.Context, serverID, publicKey string, hours int) ([]models.PeerHistoryPoint, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT recorded_at, receive_bytes, transmit_bytes FROM peer_history WHERE server_id = ? AND public_key = ? AND recorded_at > datetime('now', ? || ' hours') ORDER BY recorded_at ASC LIMIT 500`,
		serverID, publicKey, fmt.Sprintf("-%d", hours),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var points []models.PeerHistoryPoint
	for rows.Next() {
		var p models.PeerHistoryPoint
		if err := rows.Scan(&p.Time, &p.ReceiveBytes, &p.TransmitBytes); err != nil {
			return nil, err
		}
		points = append(points, p)
	}
	return points, rows.Err()
}

func (s *Store) UpsertPeerLabel(ctx context.Context, publicKey, label string) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO peer_labels (public_key, label, updated_at) VALUES (?, ?, ?) ON CONFLICT(public_key) DO UPDATE SET label = excluded.label, updated_at = excluded.updated_at`,
		publicKey, label, time.Now(),
	)
	return err
}

func (s *Store) UpsertPeerMeta(ctx context.Context, publicKey, label, notes string, expiresAt *time.Time, quotaBytes int64) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO peer_labels (public_key, label, notes, expires_at, quota_bytes, updated_at) VALUES (?, ?, ?, ?, ?, ?) ON CONFLICT(public_key) DO UPDATE SET label = excluded.label, notes = excluded.notes, expires_at = excluded.expires_at, quota_bytes = excluded.quota_bytes, updated_at = excluded.updated_at`,
		publicKey, label, notes, expiresAt, quotaBytes, time.Now(),
	)
	return err
}

func (s *Store) DeletePeerLabel(ctx context.Context, publicKey string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM peer_labels WHERE public_key = ?`, publicKey)
	return err
}

func (s *Store) GetPeerLabels(ctx context.Context) (map[string]string, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT public_key, label FROM peer_labels`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	labels := make(map[string]string)
	for rows.Next() {
		var k, v string
		if err := rows.Scan(&k, &v); err != nil {
			return nil, err
		}
		labels[k] = v
	}
	return labels, rows.Err()
}

func (s *Store) GetPeerMetas(ctx context.Context) (map[string]models.PeerMeta, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT public_key, label, notes, expires_at, quota_bytes FROM peer_labels`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	metas := make(map[string]models.PeerMeta)
	for rows.Next() {
		var k string
		var m models.PeerMeta
		var expiresAt sql.NullTime
		if err := rows.Scan(&k, &m.Label, &m.Notes, &expiresAt, &m.QuotaBytes); err != nil {
			return nil, err
		}
		if expiresAt.Valid {
			m.ExpiresAt = &expiresAt.Time
		}
		metas[k] = m
	}
	return metas, rows.Err()
}

func (s *Store) InsertAuditLog(ctx context.Context, userEmail, action, details, ip string) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO audit_log (user_email, action, details, ip, created_at) VALUES (?, ?, ?, ?, ?)`,
		userEmail, action, details, ip, time.Now(),
	)
	if err != nil {
		return err
	}
	_, _ = s.db.ExecContext(ctx,
		`DELETE FROM audit_log WHERE id NOT IN (SELECT id FROM audit_log ORDER BY id DESC LIMIT 10000)`)
	return nil
}

func (s *Store) ListAuditLogs(ctx context.Context, limit int) ([]models.AuditLog, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, user_email, action, details, ip, created_at FROM audit_log ORDER BY id DESC LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var logs []models.AuditLog
	for rows.Next() {
		var l models.AuditLog
		if err := rows.Scan(&l.ID, &l.UserEmail, &l.Action, &l.Details, &l.IP, &l.CreatedAt); err != nil {
			return nil, err
		}
		logs = append(logs, l)
	}
	return logs, rows.Err()
}






func (s *Store) GetBandwidthSummary(ctx context.Context, serverID string, hours int) (rx, tx int64, err error) {
	since := fmt.Sprintf("-%d", hours)
	q := "SELECT COALESCE(SUM(CASE WHEN rx_delta < 0 THEN 0 ELSE rx_delta END), 0)," +
		" COALESCE(SUM(CASE WHEN tx_delta < 0 THEN 0 ELSE tx_delta END), 0)" +
		" FROM (SELECT MAX(rx_bytes) - MIN(rx_bytes) AS rx_delta," +
		" MAX(tx_bytes) - MIN(tx_bytes) AS tx_delta" +
		" FROM peer_history" +
		" WHERE server_id = ? AND recorded_at > datetime('now', ? || ' hours')" +
		" GROUP BY public_key)"
	err = s.db.QueryRowContext(ctx, q, serverID, since).Scan(&rx, &tx)
	return
}

func (s *Store) Close() error {
	return s.db.Close()
}

