package ports

import (
	"errors"
	"log"
	"net/http"

	"frpg-backend/internal/app"
	"frpg-backend/internal/domain"
)

// handleSignUp creates a local (email + password) account and logs the new user
// straight in (POST /signup). Social sign-up is the OAuth find-or-create path
// (POST /auth/{provider}), not this handler.
func (s *Server) handleSignUp(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if !decodeJSON(w, r, &body) {
		return
	}

	id, err := s.Signup.SignUp(r.Context(), body.Email, body.Password)
	if err != nil {
		var invalid *domain.ErrInvalidInput
		switch {
		case errors.As(err, &invalid):
			writeError(w, http.StatusBadRequest, invalid.Reason)
		case errors.Is(err, app.ErrEmailTaken):
			writeError(w, http.StatusConflict, "email already registered")
		default:
			log.Printf("signup error: %v", err)
			writeError(w, http.StatusInternalServerError, "internal error")
		}
		return
	}

	token, err := s.mint(id)
	if err != nil {
		log.Printf("signup mint error: %v", err)
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"token": token})
}
