package api

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"frpg-backend/internal/auth"
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
	token, err := auth.Login(r.Context(), s.Local, auth.Credential{
		Email:    body.Email,
		Password: body.Password,
	}, s.mint)
	s.writeLogin(w, token, err)
}

// handleOAuth returns a handler that authenticates a social token against the
// given provider. The frontend posts the id/access token it received from
// Google/Facebook; the backend verifies it and issues our own session.
func (s *Server) handleOAuth(provider auth.IdentityProvider) http.HandlerFunc {
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
		token, err := auth.Login(r.Context(), provider, auth.Credential{Token: body.Token}, s.mint)
		s.writeLogin(w, token, err)
	}
}

// writeLogin maps a Login outcome to an HTTP response: 200 with a session token,
// 401 for bad credentials, 500 for anything else.
func (s *Server) writeLogin(w http.ResponseWriter, token string, err error) {
	if err != nil {
		var unauth *auth.ErrUnauthenticated
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

func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) bool {
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return false
	}
	return true
}
