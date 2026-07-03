package app

import (
	"context"
	"errors"
	"time"

	"frpg-backend/internal/domain"
)

// OAuthProvider is the use case for any social provider. It delegates token
// verification to a domain.ProfileVerifier (implemented by an adapter), then
// finds-or-creates the local user. Google and Facebook are the same code with a
// different verifier.
type OAuthProvider struct {
	ProviderName string
	Verifier     domain.ProfileVerifier
	Users        domain.Repository
	Now          func() time.Time
}

// NewOAuthProvider wires a social provider around its verifier and the user repo.
func NewOAuthProvider(name string, verifier domain.ProfileVerifier, repo domain.Repository) OAuthProvider {
	return OAuthProvider{ProviderName: name, Verifier: verifier, Users: repo, Now: time.Now}
}

func (p OAuthProvider) Name() string { return p.ProviderName }

// Authenticate verifies the token, then resolves it to a local account by the
// provider identity (provider + providerUserID), creating one on first sign-in.
// A bad token is a normal failure, not an error.
func (p OAuthProvider) Authenticate(ctx context.Context, cred domain.Credential) (domain.AuthResult, error) {
	profile, err := p.Verifier.Verify(ctx, cred)
	if err != nil {
		return domain.Fail(p.ProviderName + ": " + err.Error()), nil
	}
	if profile.Email == "" {
		return domain.Fail(p.ProviderName + ": provider returned no email"), nil
	}
	if !profile.EmailVerified {
		return domain.Fail(p.ProviderName + ": email is not verified"), nil
	}

	u, err := p.Users.GetByEmail(ctx, profile.Email)
	switch {
	case errors.Is(err, domain.ErrNotFound):
		// First time we've seen this identity — create the account.
		u = domain.User{
			Email:          profile.Email,
			UserID:         p.ProviderName + ":" + profile.ProviderUserID,
			Provider:       p.ProviderName,
			ProviderUserID: profile.ProviderUserID,
			DisplayName:    profile.DisplayName,
			CreatedAt:      p.now().UTC().Format(time.RFC3339),
		}
		if err := p.Users.Put(ctx, u); err != nil {
			return domain.AuthResult{}, err
		}
	case err != nil:
		return domain.AuthResult{}, err
	default:
		// An account already owns this email. Accept it only if it is the SAME
		// identity (same provider + providerUserID); never silently link a
		// different sign-in method to it. This blocks account pre-hijacking (e.g.
		// a Google login claiming a pre-existing local/other-provider account).
		if u.Provider != p.ProviderName || u.ProviderUserID != profile.ProviderUserID {
			return domain.Fail(p.ProviderName + ": email already registered with a different sign-in method"), nil
		}
	}

	return domain.Success(domain.Identity{
		UserID:         u.UserID,
		Email:          u.Email,
		DisplayName:    u.DisplayName,
		Provider:       p.ProviderName,
		ProviderUserID: u.ProviderUserID,
	}), nil
}

func (p OAuthProvider) now() time.Time {
	if p.Now != nil {
		return p.Now()
	}
	return time.Now()
}
