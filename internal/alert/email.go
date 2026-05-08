package alert

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"strings"
	"sync"
	"time"

	"github.com/ChronoCoders/sentra/internal/config"
	"github.com/ChronoCoders/sentra/internal/models"
	"github.com/rs/zerolog/log"
)

type Alerter struct {
	cfg       *config.Config
	mu        sync.Mutex
	offlineAt map[string]time.Time
	alerted   map[string]bool
}

func New(cfg *config.Config) *Alerter {
	return &Alerter{
		cfg:       cfg,
		offlineAt: make(map[string]time.Time),
		alerted:   make(map[string]bool),
	}
}

func (a *Alerter) Configured() bool {
	return a.cfg.SMTPHost != "" && a.cfg.AlertEmail != ""
}

func (a *Alerter) ServerWentOffline(serverID string) {
	a.mu.Lock()
	if _, already := a.offlineAt[serverID]; already {
		a.mu.Unlock()
		return
	}
	a.offlineAt[serverID] = time.Now()
	a.mu.Unlock()
	log.Warn().Str("server", serverID).Msg("server offline — sending alert")
	go a.sendEmail(
		fmt.Sprintf("Server %s offline — Sentrazone", serverID),
		fmt.Sprintf("Server %q went offline at %s UTC.\n\nCheck your Sentrazone dashboard.", serverID, time.Now().UTC().Format(time.RFC1123)),
	)
}

func (a *Alerter) ServerCameOnline(serverID string) {
	a.mu.Lock()
	t, was := a.offlineAt[serverID]
	if !was {
		a.mu.Unlock()
		return
	}
	delete(a.offlineAt, serverID)
	a.mu.Unlock()
	duration := time.Since(t).Round(time.Second)
	log.Info().Str("server", serverID).Dur("down_for", duration).Msg("server back online — sending alert")
	go a.sendEmail(
		fmt.Sprintf("Server %s back online — Sentrazone", serverID),
		fmt.Sprintf("Server %q is back online after being down for %s.", serverID, duration),
	)
}

func (a *Alerter) PeerQuotaExceeded(serverID, peerLabel string, usedBytes, quotaBytes int64) {
	key := serverID + "|" + peerLabel
	a.mu.Lock()
	if a.alerted[key] {
		a.mu.Unlock()
		return
	}
	a.alerted[key] = true
	a.mu.Unlock()
	go a.sendEmail(
		"Peer quota exceeded — Sentrazone",
		fmt.Sprintf("Peer %q on server %q has exceeded its bandwidth quota.\nUsed: %s / Quota: %s",
			peerLabel, serverID, fmtBytes(usedBytes), fmtBytes(quotaBytes)),
	)
}

func (a *Alerter) sendEmail(subject, body string) {
	if !a.Configured() {
		return
	}
	from := a.cfg.SMTPUser
	to := a.cfg.AlertEmail
	msg := fmt.Sprintf("From: Sentrazone <%s>\r\nTo: %s\r\nSubject: %s\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n%s\r\n\r\n-- Sentrazone Dashboard", from, to, subject, body)
	addr := fmt.Sprintf("%s:%d", a.cfg.SMTPHost, a.cfg.SMTPPort)
	conn, err := net.DialTimeout("tcp", addr, 10*time.Second)
	if err != nil {
		log.Error().Err(err).Msg("smtp dial failed")
		return
	}
	client, err := smtp.NewClient(conn, a.cfg.SMTPHost)
	if err != nil {
		conn.Close()
		log.Error().Err(err).Msg("smtp client failed")
		return
	}
	defer client.Close()
	tlsCfg := &tls.Config{ServerName: a.cfg.SMTPHost}
	if ok, _ := client.Extension("STARTTLS"); ok {
		if err := client.StartTLS(tlsCfg); err != nil {
			log.Error().Err(err).Msg("smtp starttls failed")
			return
		}
	}
	if a.cfg.SMTPUser != "" && a.cfg.SMTPPass != "" {
		if ok, _ := client.Extension("AUTH"); ok {
			auth := smtp.PlainAuth("", a.cfg.SMTPUser, a.cfg.SMTPPass, a.cfg.SMTPHost)
			if err := client.Auth(auth); err != nil {
				log.Error().Err(err).Msg("smtp auth failed")
				return
			}
		}
	}
	if err := client.Mail(from); err != nil {
		log.Error().Err(err).Msg("smtp mail from failed")
		return
	}
	if err := client.Rcpt(to); err != nil {
		log.Error().Err(err).Msg("smtp rcpt failed")
		return
	}
	wc, err := client.Data()
	if err != nil {
		log.Error().Err(err).Msg("smtp data failed")
		return
	}
	fmt.Fprint(wc, msg)
	if err := wc.Close(); err != nil {
		log.Error().Err(err).Msg("smtp close failed")
	}
}

func (a *Alerter) SendSummary(label string, hours int, servers []models.ServerSummary) {
	if !a.Configured() || len(servers) == 0 {
		return
	}
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Sentrazone %s Summary\n%s\n\n", label, time.Now().UTC().Format("January 2, 2006")))
	sb.WriteString(strings.Repeat("-", 42) + "\n\n")
	for _, s := range servers {
		sb.WriteString(fmt.Sprintf("%s\n", s.ServerID))
		sb.WriteString(fmt.Sprintf("  Uptime        %.1f%%\n", s.UptimePC))
		sb.WriteString(fmt.Sprintf("  Active peers  %d\n", s.ActivePeers))
		sb.WriteString(fmt.Sprintf("  Bandwidth     down %s  up %s\n\n", fmtBytes(s.RxBytes), fmtBytes(s.TxBytes)))
	}
	sb.WriteString(strings.Repeat("-", 42) + "\n")
	sb.WriteString("Sentrazone Dashboard\n")
	go a.sendEmail(
		fmt.Sprintf("Sentrazone %s Summary - %s", label, time.Now().UTC().Format("Jan 2, 2006")),
		sb.String(),
	)
}

func fmtBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}
