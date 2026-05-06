package api

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/ChronoCoders/sentra/internal/auth"
	"github.com/ChronoCoders/sentra/internal/config"
	"github.com/ChronoCoders/sentra/internal/control"
	"github.com/ChronoCoders/sentra/internal/models"
	"github.com/ChronoCoders/sentra/internal/store"
	"github.com/ChronoCoders/sentra/internal/wgeasy"
	"github.com/ChronoCoders/sentra/internal/ws"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/bcrypt"
)

type Server struct {
	cfg          *config.Config
	store        *store.Store
	client       control.AgentClient
	hub          *ws.Hub
	bus          *control.EventBus
	auth         *auth.JWTManager
	router       *chi.Mux
	loginLimiter *loginRateLimiter
	wgEasy       map[string]*wgeasy.Client
}

func NewServer(cfg *config.Config, store *store.Store, client control.AgentClient, hub *ws.Hub, bus *control.EventBus) *Server {
	r := chi.NewRouter()
	s := &Server{
		cfg:          cfg,
		store:        store,
		client:       client,
		hub:          hub,
		bus:          bus,
		auth:         auth.NewJWTManager(cfg.JWTSecret),
		router:       r,
		loginLimiter: newLoginRateLimiter(),
		wgEasy:       make(map[string]*wgeasy.Client),
	}
	for serverID, host := range wgeasy.ServerHosts {
		if c, err := wgeasy.New(host, cfg.WGEasyPassword); err == nil {
			s.wgEasy[serverID] = c
		}
	}
	s.setupRoutes()
	return s
}

func (s *Server) setupRoutes() {
	s.router.Use(middleware.Logger)
	s.router.Use(middleware.Recoverer)

	s.router.Post("/api/login", s.handleLogin)
	s.router.Post("/api/report", s.handleReport)

	s.router.Group(func(r chi.Router) {
		r.Use(s.jwtMiddleware)

		r.Get("/api/health", s.handleHealth)
		r.Get("/api/status", s.handleStatus)
		r.Get("/api/statuses", s.handleStatuses)
		r.Get("/api/history", s.handleHistory)
		r.Get("/api/history/export", s.handleHistoryExport)
		r.Get("/api/peer-labels", s.handleGetPeerLabels)
		r.Get("/api/peer-meta", s.handleGetPeerMetas)
		r.Get("/api/peer-history", s.handleGetPeerHistory)
		r.Get("/api/uptime", s.handleUptime)
		r.Get("/api/alert-config", s.handleAlertConfig)
		r.Get("/api/tunnel/status", s.handleTunnelStatus)
		r.Get("/api/tunnel/config", s.handleTunnelConfig)
		r.Get("/ws", s.handleWs)

		r.Group(func(r chi.Router) {
			r.Use(s.RequireRole("admin"))

			r.Get("/api/users", s.handleListUsers)
			r.Post("/api/users", s.handleCreateUser)
			r.Delete("/api/users/{id}", s.handleDeleteUser)
			r.Post("/api/peer-labels", s.handleUpsertPeerLabel)
			r.Delete("/api/peer-labels", s.handleDeletePeerLabel)
			r.Put("/api/peer-meta", s.handleUpsertPeerMeta)
			r.Get("/api/audit", s.handleListAudit)

			r.Get("/api/wg-clients/{serverID}", s.handleWGClients)
			r.Post("/api/wg-clients/{serverID}", s.handleWGCreateClient)
			r.Post("/api/wg-clients/{serverID}/{clientID}/{action}", s.handleWGSetEnabled)
			r.Get("/api/wg-clients/{serverID}/{clientID}/qrcode", s.handleWGQRCode)
		})

		r.Put("/api/users/{id}/password", s.handleChangePassword)
	})

	workDir, _ := os.Getwd()
	filesDir := http.Dir(filepath.Join(workDir, "web"))
	FileServer(s.router, "/", filesDir)
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

func claimsFromCtx(r *http.Request) *auth.UserClaims {
	claims, _ := r.Context().Value(contextKey("user")).(*auth.UserClaims)
	return claims
}

func jsonOK(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}

func extractIP(r *http.Request) string {
	ip := r.Header.Get("X-Forwarded-For")
	if ip != "" {
		return strings.TrimSpace(strings.Split(ip, ",")[0])
	}
	ip = r.RemoteAddr
	if idx := strings.LastIndex(ip, ":"); idx != -1 {
		return ip[:idx]
	}
	return ip
}

func (s *Server) audit(r *http.Request, userEmail, action, details string) {
	go s.store.InsertAuditLog(context.Background(), userEmail, action, details, extractIP(r))
}

func FileServer(r chi.Router, path string, root http.FileSystem) {
	if strings.ContainsAny(path, "{}*") {
		panic("FileServer does not permit any URL parameters.")
	}
	if path != "/" && path[len(path)-1] != '/' {
		r.Get(path, http.RedirectHandler(path+"/", 301).ServeHTTP)
		path += "/"
	}
	path += "*"
	r.Get(path, func(w http.ResponseWriter, r *http.Request) {
		rctx := chi.RouteContext(r.Context())
		pathPrefix := strings.TrimSuffix(rctx.RoutePattern(), "/*")
		fs := http.StripPrefix(pathPrefix, http.FileServer(root))
		fs.ServeHTTP(w, r)
	})
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	serverID := r.URL.Query().Get("server_id")
	if serverID == "" {
		http.Error(w, "missing server_id", http.StatusBadRequest)
		return
	}
	status, err := s.client.GetStatus(r.Context(), serverID)
	if err != nil {
		log.Error().Err(err).Msg("failed to get status")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if status == nil {
		http.Error(w, "server not found", http.StatusNotFound)
		return
	}
	jsonOK(w, status)
}

func (s *Server) handleStatuses(w http.ResponseWriter, r *http.Request) {
	jsonOK(w, s.client.GetAllStatuses())
}

func (s *Server) handleWs(w http.ResponseWriter, r *http.Request) {
	client := ws.ServeWs(s.hub, w, r)
	if client != nil {
		for _, event := range s.client.GetAllStatuses() {
			client.Send(event)
		}
	}
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	ip := extractIP(r)
	if !s.loginLimiter.Allow(ip) {
		http.Error(w, "too many requests", http.StatusTooManyRequests)
		return
	}

	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	user, err := s.store.GetUserByEmail(r.Context(), req.Email)
	if err != nil {
		log.Error().Err(err).Msg("failed to get user")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if user == nil || bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)) != nil {
		s.audit(r, req.Email, "login_failed", "invalid credentials")
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	token, err := s.auth.GenerateAccessToken(user.ID, user.Role)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	s.audit(r, user.Email, "login", "")
	jsonOK(w, map[string]string{"token": token, "role": user.Role})
}

func (s *Server) handleReport(w http.ResponseWriter, r *http.Request) {
	if s.cfg.AuthToken == "" || r.Header.Get("Authorization") != "Bearer "+s.cfg.AuthToken {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	var event models.StatusEvent
	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	event.Time = time.Now()
	s.bus.Publish(event)
	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleHistory(w http.ResponseWriter, r *http.Request) {
	serverID := r.URL.Query().Get("server_id")
	if serverID == "" {
		http.Error(w, "missing server_id", http.StatusBadRequest)
		return
	}
	hours := 24
	if h := r.URL.Query().Get("hours"); h != "" {
		if n, err := strconv.Atoi(h); err == nil && n > 0 && n <= 72 {
			hours = n
		}
	}
	points, err := s.store.GetHistory(r.Context(), serverID, hours)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if points == nil {
		points = []models.HistoryPoint{}
	}
	jsonOK(w, points)
}

func (s *Server) handleHistoryExport(w http.ResponseWriter, r *http.Request) {
	serverID := r.URL.Query().Get("server_id")
	if serverID == "" {
		http.Error(w, "missing server_id", http.StatusBadRequest)
		return
	}
	hours := 24
	if h := r.URL.Query().Get("hours"); h != "" {
		if n, err := strconv.Atoi(h); err == nil && n > 0 && n <= 72 {
			hours = n
		}
	}
	points, err := s.store.GetHistory(r.Context(), serverID, hours)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	filename := fmt.Sprintf("sentra-%s-%dh.csv", serverID, hours)
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment; filename="+filename)

	cw := csv.NewWriter(w)
	cw.Write([]string{"time", "cpu_percent", "memory_percent", "disk_percent", "load_average", "net_bytes_sent", "net_bytes_recv", "peer_count", "ping_latency_ms"})
	for _, p := range points {
		cw.Write([]string{
			p.Time.UTC().Format(time.RFC3339),
			fmt.Sprintf("%.2f", p.CPUPercent),
			fmt.Sprintf("%.2f", p.MemoryPercent),
			fmt.Sprintf("%.2f", p.DiskPercent),
			fmt.Sprintf("%.2f", p.LoadAverage),
			strconv.FormatUint(p.NetBytesSent, 10),
			strconv.FormatUint(p.NetBytesRecv, 10),
			strconv.Itoa(p.PeerCount),
			fmt.Sprintf("%.1f", p.PingLatencyMs),
		})
	}
	cw.Flush()
}

func (s *Server) handleUptime(w http.ResponseWriter, r *http.Request) {
	hours := 24
	if h := r.URL.Query().Get("hours"); h != "" {
		if n, err := strconv.Atoi(h); err == nil && n > 0 && n <= 72 {
			hours = n
		}
	}
	events := s.client.GetAllStatuses()
	result := make(map[string]float64)
	for _, e := range events {
		pct, err := s.store.GetUptimePercent(r.Context(), e.ServerID, hours)
		if err == nil {
			result[e.ServerID] = pct
		}
	}
	jsonOK(w, result)
}

func (s *Server) handleGetPeerHistory(w http.ResponseWriter, r *http.Request) {
	serverID := r.URL.Query().Get("server_id")
	publicKey := r.URL.Query().Get("public_key")
	if serverID == "" || publicKey == "" {
		http.Error(w, "missing server_id or public_key", http.StatusBadRequest)
		return
	}
	hours := 6
	if h := r.URL.Query().Get("hours"); h != "" {
		if n, err := strconv.Atoi(h); err == nil && n > 0 && n <= 24 {
			hours = n
		}
	}
	points, err := s.store.GetPeerHistory(r.Context(), serverID, publicKey, hours)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if points == nil {
		points = []models.PeerHistoryPoint{}
	}
	jsonOK(w, points)
}

func (s *Server) handleAlertConfig(w http.ResponseWriter, r *http.Request) {
	jsonOK(w, map[string]interface{}{
		"configured":   s.cfg.SMTPHost != "" && s.cfg.AlertEmail != "",
		"alert_email":  s.cfg.AlertEmail,
		"smtp_host":    s.cfg.SMTPHost,
	})
}

func (s *Server) handleGetPeerLabels(w http.ResponseWriter, r *http.Request) {
	labels, err := s.store.GetPeerLabels(r.Context())
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	jsonOK(w, labels)
}

func (s *Server) handleGetPeerMetas(w http.ResponseWriter, r *http.Request) {
	metas, err := s.store.GetPeerMetas(r.Context())
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if metas == nil {
		metas = make(map[string]models.PeerMeta)
	}
	jsonOK(w, metas)
}

func (s *Server) handleUpsertPeerLabel(w http.ResponseWriter, r *http.Request) {
	var req struct {
		PublicKey string `json:"public_key"`
		Label     string `json:"label"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.PublicKey == "" {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	claims := claimsFromCtx(r)
	if err := s.store.UpsertPeerLabel(r.Context(), req.PublicKey, req.Label); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	s.audit(r, claims.Subject, "peer_label_set", req.PublicKey[:8]+"... = "+req.Label)
	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleDeletePeerLabel(w http.ResponseWriter, r *http.Request) {
	var req struct {
		PublicKey string `json:"public_key"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.PublicKey == "" {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	if err := s.store.DeletePeerLabel(r.Context(), req.PublicKey); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleUpsertPeerMeta(w http.ResponseWriter, r *http.Request) {
	var req struct {
		PublicKey  string  `json:"public_key"`
		Label      string  `json:"label"`
		Notes      string  `json:"notes"`
		ExpiresAt  *string `json:"expires_at"`
		QuotaBytes int64   `json:"quota_bytes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.PublicKey == "" {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	var expiresAt *time.Time
	if req.ExpiresAt != nil && *req.ExpiresAt != "" {
		t, err := time.Parse(time.RFC3339, *req.ExpiresAt)
		if err != nil {
			http.Error(w, "invalid expires_at format (RFC3339 required)", http.StatusBadRequest)
			return
		}
		expiresAt = &t
	}
	claims := claimsFromCtx(r)
	if err := s.store.UpsertPeerMeta(r.Context(), req.PublicKey, req.Label, req.Notes, expiresAt, req.QuotaBytes); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	s.audit(r, claims.Subject, "peer_meta_updated", req.PublicKey[:8]+"...")
	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleListAudit(w http.ResponseWriter, r *http.Request) {
	limit := 100
	if l := r.URL.Query().Get("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 {
			limit = n
		}
	}
	logs, err := s.store.ListAuditLogs(r.Context(), limit)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if logs == nil {
		logs = []models.AuditLog{}
	}
	jsonOK(w, logs)
}

func (s *Server) handleListUsers(w http.ResponseWriter, r *http.Request) {
	users, err := s.store.ListUsers(r.Context())
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if users == nil {
		users = []models.User{}
	}
	jsonOK(w, users)
}

func (s *Server) handleCreateUser(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Name     string `json:"name"`
		Password string `json:"password"`
		Role     string `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	if req.Email == "" || req.Password == "" {
		http.Error(w, "email and password required", http.StatusBadRequest)
		return
	}
	if req.Role != "admin" && req.Role != "viewer" {
		req.Role = "viewer"
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	u := &models.User{
		ID:        uuid.New().String(),
		OrgID:     "org1",
		Email:     req.Email,
		Name:      req.Name,
		Role:      req.Role,
		Password:  string(hash),
		CreatedAt: time.Now(),
	}
	if err := s.store.CreateUser(r.Context(), u); err != nil {
		if strings.Contains(err.Error(), "UNIQUE") {
			http.Error(w, "email already exists", http.StatusConflict)
			return
		}
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	claims := claimsFromCtx(r)
	s.audit(r, claims.Subject, "user_created", req.Email+" role="+req.Role)
	u.Password = ""
	w.WriteHeader(http.StatusCreated)
	jsonOK(w, u)
}

func (s *Server) handleDeleteUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	claims := claimsFromCtx(r)

	if claims != nil && claims.Subject == id {
		http.Error(w, "cannot delete yourself", http.StatusBadRequest)
		return
	}

	target, err := s.store.GetUserByID(r.Context(), id)
	if err != nil || target == nil {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}

	if target.Role == "admin" {
		count, err := s.store.CountAdmins(r.Context())
		if err != nil || count <= 1 {
			http.Error(w, "cannot delete the last admin", http.StatusBadRequest)
			return
		}
	}

	if err := s.store.DeleteUser(r.Context(), id); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	s.audit(r, claims.Subject, "user_deleted", target.Email)
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleChangePassword(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	claims := claimsFromCtx(r)

	if claims == nil || (claims.Role != "admin" && claims.Subject != id) {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	var req struct {
		Password    string `json:"password"`
		OldPassword string `json:"old_password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Password == "" {
		http.Error(w, "password required", http.StatusBadRequest)
		return
	}

	if claims.Role != "admin" {
		user, err := s.store.GetUserByID(r.Context(), id)
		if err != nil || user == nil {
			http.Error(w, "user not found", http.StatusNotFound)
			return
		}
		if bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.OldPassword)) != nil {
			http.Error(w, "current password incorrect", http.StatusUnauthorized)
			return
		}
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if err := s.store.UpdateUserPassword(r.Context(), id, string(hash)); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	s.audit(r, claims.Subject, "password_changed", "user_id="+id)
	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleTunnelStatus(w http.ResponseWriter, r *http.Request) {
	jsonOK(w, map[string]string{
		"wstunnel": tcpReachable("host.docker.internal:443"),
	})
}

func (s *Server) handleTunnelConfig(w http.ResponseWriter, r *http.Request) {
	host := s.cfg.HostIP
	jsonOK(w, map[string]interface{}{
		"host_ip":      host,
		"wstunnel_url": "wss://" + host + ":443",
		"wg_port":      51820,
		"wg_interface": s.cfg.WGInterface,
	})
}

func tcpReachable(addr string) string {
	conn, err := net.DialTimeout("tcp", addr, 2*time.Second)
	if err != nil {
		return "stopped"
	}
	conn.Close()
	return "running"
}

var _ = context.Background
