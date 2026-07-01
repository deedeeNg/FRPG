package auth

import (
	"context"
	"errors"
	"time"

	"frpg-backend/internal/users"
)

// ProviderProfile is the identity a social provider (Google/Facebook) returns
// about a user once their token has been verified. It is the only thing that
// crosses the network boundary; everything past it is our own logic.
type ProviderProfile struct {
	ProviderUserID string
	Email          string
	DisplayName    string
}

// ProfileVerifier turns an OAuth credential into a verified ProviderProfile.
// This is the single external boundary: the real implementation calls
// Google/Facebook, tests inject a fake. Because only this is mocked, tests
// exercise the real OAuthProvider logic (find-or-create, identity mapping)
// instead of a stand-in for the whole provider.
type ProfileVerifier interface {
	Verify(ctx context.Context, cred Credential) (ProviderProfile, error)
}

// OAuthProvider is the IdentityProvider for any social provider. It delegates
// token verification to a ProfileVerifier, then finds-or-creates the local user.
// Google and Facebook are the same code with a different Verifier.
type OAuthProvider struct {
	ProviderName string
	Verifier     ProfileVerifier
	Users        users.Repository
	Now          func() time.Time
}

// NewOAuthProvider wires a social provider around its verifier and the user repo.
func NewOAuthProvider(name string, verifier ProfileVerifier, repo users.Repository) OAuthProvider {
	return OAuthProvider{ProviderName: name, Verifier: verifier, Users: repo, Now: time.Now}
}

func (p OAuthProvider) Name() string { return p.ProviderName }

// Authenticate verifies the token via the provider, then maps the profile to a
// local account (creating one on first sign-in). A bad/expired token is a normal
// authentication failure, not a server error.
func (p OAuthProvider) Authenticate(ctx context.Context, cred Credential) (AuthResult, error) {
	profile, err := p.Verifier.Verify(ctx, cred)
	if err != nil {
		return fail(p.ProviderName + ": " + err.Error()), nil
	}
	if profile.Email == "" {
		return fail(p.ProviderName + ": provider returned no email"), nil
	}

	u, err := p.Users.GetByEmail(ctx, profile.Email)
	switch {
	case errors.Is(err, users.ErrNotFound):
		// First sign-in for this social user: auto-provision an account.
		u = users.User{
			Email:          profile.Email,
			UserID:         p.ProviderName + ":" + profile.ProviderUserID,
			Provider:       p.ProviderName,
			ProviderUserID: profile.ProviderUserID,
			DisplayName:    profile.DisplayName,
			CreatedAt:      p.now().UTC().Format(time.RFC3339),
		}
		if err := p.Users.Put(ctx, u); err != nil {
			return AuthResult{}, err
		}
	case err != nil:
		return AuthResult{}, err
	}

	return success(Identity{
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
