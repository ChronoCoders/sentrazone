package api

import (
	"encoding/json"
	"net/http"

	"github.com/ChronoCoders/sentra/internal/wgeasy"
	"github.com/go-chi/chi/v5"
)

func (s *Server) wgClient(serverID string) *wgeasy.Client {
	if s.wgEasy == nil {
		return nil
	}
	return s.wgEasy[serverID]
}

func (s *Server) handleWGCreateClient(w http.ResponseWriter, r *http.Request) {
	serverID := chi.URLParam(r, "serverID")
	c := s.wgClient(serverID)
	if c == nil {
		http.Error(w, "unknown server", http.StatusNotFound)
		return
	}
	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Name == "" {
		http.Error(w, "name required", http.StatusBadRequest)
		return
	}
	client, err := c.CreateClient(r.Context(), req.Name)
	if err != nil {
		http.Error(w, "failed to create client", http.StatusBadGateway)
		return
	}
	claims := claimsFromCtx(r)
	s.audit(r, claims.Subject, "wg_client_created", serverID+"/"+req.Name)
	w.WriteHeader(http.StatusCreated)
	jsonOK(w, client)
}

func (s *Server) handleWGClients(w http.ResponseWriter, r *http.Request) {
	serverID := chi.URLParam(r, "serverID")
	c := s.wgClient(serverID)
	if c == nil {
		http.Error(w, "unknown server", http.StatusNotFound)
		return
	}
	clients, err := c.GetClients(r.Context())
	if err != nil {
		http.Error(w, "failed to fetch clients", http.StatusBadGateway)
		return
	}
	jsonOK(w, clients)
}

func (s *Server) handleWGSetEnabled(w http.ResponseWriter, r *http.Request) {
	serverID := chi.URLParam(r, "serverID")
	clientID := chi.URLParam(r, "clientID")
	action := chi.URLParam(r, "action")

	c := s.wgClient(serverID)
	if c == nil {
		http.Error(w, "unknown server", http.StatusNotFound)
		return
	}
	enabled := action == "enable"
	if err := c.SetEnabled(r.Context(), clientID, enabled); err != nil {
		http.Error(w, "failed to update client", http.StatusBadGateway)
		return
	}
	claims := claimsFromCtx(r)
	s.audit(r, claims.Subject, "wg_client_"+action+"d", serverID+"/"+clientID[:8])
	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleWGQRCode(w http.ResponseWriter, r *http.Request) {
	serverID := chi.URLParam(r, "serverID")
	clientID := chi.URLParam(r, "clientID")

	c := s.wgClient(serverID)
	if c == nil {
		http.Error(w, "unknown server", http.StatusNotFound)
		return
	}
	svg, err := c.GetQRCode(r.Context(), clientID)
	if err != nil {
		http.Error(w, "failed to fetch qr code", http.StatusBadGateway)
		return
	}
	w.Header().Set("Content-Type", "image/svg+xml")
	w.Write(svg)
}
