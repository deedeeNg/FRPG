package api

import (
	"context"
	"net/http"
	"strings"

	"frpg-backend/internal/session"
)

type ctxKey int

const sessionKey ctxKey = 0

// RequireAuth rejects requests without a valid Bearer session token, and stores
// the verified session in the request context for downstream handlers.
func (s *Server) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, ok := bearerToken(r)
		if !ok {
			writeError(w, http.StatusUnauthorized, "missing bearer token")
			return
		}
		sess, err := s.Sessions.Parse(token)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "invalid session")
			return
		}
		ctx := context.WithValue(r.Context(), sessionKey, sess)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// sessionFrom returns the session stored by RequireAuth.
func sessionFrom(ctx context.Context) (session.Session, bool) {
	sess, ok := ctx.Value(sessionKey).(session.Session)
	return sess, ok
}

func bearerToken(r *http.Request) (string, bool) {
	h := r.Header.Get("Authorization")
	if h == "" {
		return "", false
	}
	const prefix = "Bearer "
	if !strings.HasPrefix(h, prefix) {
		return "", false
	}
	return strings.TrimPrefix(h, prefix), true
}

// handleMe returns the authenticated user's session (a simple protected route).
func (s *Server) handleMe(w http.ResponseWriter, r *http.Request) {
	sess, ok := sessionFrom(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "no session")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{
		"userId": sess.Subject,
		"email":  sess.Email,
	})
}
