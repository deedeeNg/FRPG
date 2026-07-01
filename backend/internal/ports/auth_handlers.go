package ports

import (
	"errors"
	"log"
	"net/http"

	"frpg-backend/internal/app"
	"frpg-backend/internal/domain"
)

// handleLogin authenticates an email/password against the local provider.
func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if !decodeJSON(w, r, &body) {
		return
	}
	token, err := app.Login(r.Context(), s.Local, domain.Credential{
		Email:    body.Email,
		Password: body.Password,
	}, s.mint)
	s.writeLogin(w, token, err)
}

// handleOAuth returns a handler that authenticates a social token against the
// given provider.
func (s *Server) handleOAuth(provider domain.IdentityProvider) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if provider == nil {
			writeError(w, http.StatusNotImplemented, "provider not configured")
			return
		}
		var body struct {
			Token string `json:"token"`
		}
		if !decodeJSON(w, r, &body) {
			return
		}
		token, err := app.Login(r.Context(), provider, domain.Credential{Token: body.Token}, s.mint)
		s.writeLogin(w, token, err)
	}
}

// writeLogin maps a Login outcome to an HTTP response: 200 with a session token,
// 401 for bad credentials, 500 for anything else.
func (s *Server) writeLogin(w http.ResponseWriter, token string, err error) {
	if err != nil {
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
