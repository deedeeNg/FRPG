// Package ports is the inbound HTTP delivery layer (the "driving adapter"). It
// depends on app (use cases) and domain (types + ports), and is wired by service.
package ports

import (
	"net/http"

	"frpg-backend/internal/domain"
)

// Server holds the HTTP handlers' dependencies as domain ports, injected by the
// service layer in production and by fakes/mocks in tests.
type Server struct {
	Local    domain.IdentityProvider
	Google   domain.IdentityProvider
	Facebook domain.IdentityProvider
	Sessions domain.SessionManager
}

// Routes builds the request router (Go 1.22 method+pattern routing).
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

// mint adapts the session manager to domain.MintSession.
func (s *Server) mint(id domain.Identity) (string, error) {
	return s.Sessions.Mint(id.UserID, id.Email)
}
