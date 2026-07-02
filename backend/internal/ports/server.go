// Package ports is the inbound HTTP delivery layer (the "driving adapter"). It
// depends on app (use cases) and domain (types + ports), and is wired by service.
package ports

import (
	"net/http"

	"frpg-backend/internal/app"
	"frpg-backend/internal/domain"
)

// Server holds the HTTP handlers' dependencies, injected by the service layer in
// production and by fakes in tests. Identity is the provider registry, so a
// single route handles local/google/facebook.
type Server struct {
	Identity *app.Manager
	Sessions domain.SessionManager
}

// Routes builds the request router (Go 1.22 method+pattern routing). One
// endpoint fronts every provider: POST /auth/{provider} (local|google|facebook).
func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/health", s.handleHealth)
	mux.HandleFunc("POST /auth/{provider}", s.handleAuth)
	mux.Handle("GET /api/me", s.RequireAuth(http.HandlerFunc(s.handleMe)))

	return mux
}

func (s *Server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// mint adapts the session manager to domain.MintSession.
func (s *Server) mint(id domain.Identity) (string, error) {
	return s.Sessions.Mint(id.UserID, id.Email)
}
