// Package api exposes the HTTP layer. A Server holds its dependencies as fields
// (the identity providers and the session manager), which are injected by main
// in production and by fakes/mocks in tests.
package api

import (
	"encoding/json"
	"log"
	"net/http"

	"frpg-backend/internal/auth"
	"frpg-backend/internal/session"
)

// Server wires the HTTP handlers to their dependencies.
type Server struct {
	Local    auth.IdentityProvider
	Google   auth.IdentityProvider
	Facebook auth.IdentityProvider
	Sessions *session.Manager
}

// Routes builds the request router. Go 1.22 method+pattern routing keeps this
// close to an Express-style route table.
func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/health", s.handleHealth)
	mux.HandleFunc("POST /auth/login", s.handleLogin)
	mux.HandleFunc("POST /auth/oauth/google", s.handleOAuth(s.Google))
	mux.HandleFunc("POST /auth/oauth/facebook", s.handleOAuth(s.Facebook))
	mux.Handle("GET /api/me", s.RequireAuth(http.HandlerFunc(s.handleMe)))

	return mux
}

func (s *Server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// mint adapts the session manager to auth.MintSession.
func (s *Server) mint(id auth.Identity) (string, error) {
	return s.Sessions.Mint(id.UserID, id.Email)
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Printf("write json: %v", err)
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
