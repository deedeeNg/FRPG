package ports

import (
	"errors"
	"log"
	"net/http"

	"frpg-backend/internal/app"
	"frpg-backend/internal/domain"
)

// handleAuth authenticates against the provider named in the path
// (POST /auth/{provider}). It decodes a superset body and lets each provider
// read only the fields it needs (local: email/password; social: token), so no
// per-provider branching is required.
func (s *Server) handleAuth(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("provider")

	var body struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		Token    string `json:"token"`
	}
	if !decodeJSON(w, r, &body) {
		return
	}

	token, err := s.Identity.Login(r.Context(), name, domain.Credential{
		Email:    body.Email,
		Password: body.Password,
		Token:    body.Token,
	}, s.mint)
	s.writeLogin(w, name, token, err)
}

// writeLogin maps a Login outcome to an HTTP response: 200 with a session token,
// 404 for an unknown provider, 401 for bad credentials, 500 for anything else.
func (s *Server) writeLogin(w http.ResponseWriter, provider, token string, err error) {
	if err != nil {
		if errors.Is(err, app.ErrUnknownProvider) {
			writeError(w, http.StatusNotFound, "unknown provider: "+provider)
			return
		}
		var unauth *domain.ErrUnauthenticated
		if errors.As(err, &unauth) {
			writeError(w, http.StatusUnauthorized, unauth.Reason)
			return
		}
		log.Printf("login error: %v", err)
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"token": token})
}
