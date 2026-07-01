package auth

import (
	"context"
	"errors"

	"golang.org/x/crypto/bcrypt"

	"frpg-backend/internal/users"
)

// LocalProvider authenticates email + password against the user repository.
type LocalProvider struct {
	Users users.Repository
}

// NewLocalProvider wires a LocalProvider to a user repository.
func NewLocalProvider(repo users.Repository) LocalProvider {
	return LocalProvider{Users: repo}
}

func (LocalProvider) Name() string { return "local" }

// Authenticate returns Authenticated=true only when the email exists, the
// account has a password (i.e. is not a social-only account), and the password
// matches. Every failure path returns a non-error AuthResult so the caller can
// treat "wrong credentials" as a normal 401 rather than a server error.
func (p LocalProvider) Authenticate(ctx context.Context, cred Credential) (AuthResult, error) {
	u, err := p.Users.GetByEmail(ctx, cred.Email)
	if errors.Is(err, users.ErrNotFound) {
		return fail("unknown email"), nil
	}
	if err != nil {
		return AuthResult{}, err
	}
	if u.PasswordHash == "" {
		return fail("account has no password; use social login"), nil
	}
	if bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(cred.Password)) != nil {
		return fail("wrong password"), nil
	}
	return success(Identity{
		UserID:      u.UserID,
		Email:       u.Email,
		DisplayName: u.DisplayName,
		Provider:    "local",
	}), nil
}
