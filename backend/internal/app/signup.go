package app

import (
	"context"
	"errors"
	"regexp"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"frpg-backend/internal/domain"
)

// ErrEmailTaken is returned when signing up an email that already has an account
// (of any sign-in method). The ports layer maps it to a 409.
var ErrEmailTaken = errors.New("email already registered")

// emailShape is a deliberately loose check: it only confirms the value LOOKS like
// an email (local@domain.tld). We intentionally do NOT enforce full RFC 5322 rules
// and do NOT verify the mailbox exists / is deliverable — any well-formed-looking
// address is accepted as-is.
var emailShape = regexp.MustCompile(`^[^@\s]+@[^@\s]+\.[^@\s]+$`)

// LocalSignUp is the use case for creating a local (email + password) account.
// Social sign-up is handled by the OAuth find-or-create path, not here.
type LocalSignUp struct {
	Users domain.Repository
	Now   func() time.Time
}

// NewLocalSignUp wires a LocalSignUp to a user repository.
func NewLocalSignUp(repo domain.Repository) LocalSignUp {
	return LocalSignUp{Users: repo, Now: time.Now}
}

// SignUp validates the input, ensures the email is free, bcrypt-hashes the
// password, and stores the new user. It returns the created Identity; the caller
// mints a session (so sign-up logs the user straight in).
func (s LocalSignUp) SignUp(ctx context.Context, email, password string) (domain.Identity, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	// Loose shape check only (see emailShape) — we don't validate or verify the email.
	if !emailShape.MatchString(email) {
		return domain.Identity{}, &domain.ErrInvalidInput{Reason: "enter an email like name@example.com"}
	}
	if len(password) < 8 {
		return domain.Identity{}, &domain.ErrInvalidInput{Reason: "password must be at least 8 characters"}
	}

	switch _, err := s.Users.GetByEmail(ctx, email); {
	case err == nil:
		return domain.Identity{}, ErrEmailTaken
	case !errors.Is(err, domain.ErrNotFound):
		return domain.Identity{}, err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return domain.Identity{}, err
	}

	u := domain.User{
		Email:        email,
		UserID:       "local:" + email,
		Provider:     "local",
		PasswordHash: string(hash),
		CreatedAt:    s.now().UTC().Format(time.RFC3339),
	}
	if err := s.Users.Put(ctx, u); err != nil {
		return domain.Identity{}, err
	}
	return domain.Identity{UserID: u.UserID, Email: u.Email, Provider: "local"}, nil
}

func (s LocalSignUp) now() time.Time {
	if s.Now != nil {
		return s.Now()
	}
	return time.Now()
}
