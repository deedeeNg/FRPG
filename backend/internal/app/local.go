package app

import (
	"context"
	"errors"

	"golang.org/x/crypto/bcrypt"

	"frpg-backend/internal/domain"
)

// LocalProvider authenticates email + password against the user repository.
type LocalProvider struct {
	Users domain.Repository
}

// NewLocalProvider wires a LocalProvider to a user repository.
func NewLocalProvider(repo domain.Repository) LocalProvider {
	return LocalProvider{Users: repo}
}

func (LocalProvider) Name() string { return "local" }

// Authenticate returns Authenticated=true only when the email exists, the
// account has a password, and the password matches. Every failure path returns a
// non-error AuthResult so the caller treats bad credentials as a 401, not a 500.
func (p LocalProvider) Authenticate(ctx context.Context, cred domain.Credential) (domain.AuthResult, error) {
	u, err := p.Users.GetByEmail(ctx, cred.Email)
	if errors.Is(err, domain.ErrNotFound) {
		return domain.Fail("unknown email"), nil
	}
	if err != nil {
		return domain.AuthResult{}, err
	}
	if u.PasswordHash == "" {
		return domain.Fail("account has no password; use social login"), nil
	}
	if bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(cred.Password)) != nil {
		return domain.Fail("wrong password"), nil
	}
	return domain.Success(domain.Identity{
		UserID:      u.UserID,
		Email:       u.Email,
		DisplayName: u.DisplayName,
		Provider:    "local",
	}), nil
}
