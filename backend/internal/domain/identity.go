// Package domain holds the pure business entities, value objects, and the port
// interfaces the outer layers implement. It depends on nothing else in the
// module — dependencies point inward, toward domain.
package domain

import "context"

// Identity is the normalized user info a provider produces on success. It is
// provider-agnostic: local, Google and Facebook all reduce to this shape.
type Identity struct {
	UserID         string
	Email          string
	DisplayName    string
	Provider       string // "local" | "google" | "facebook"
	ProviderUserID string
}

// AuthResult is the single yes/no outcome every provider returns.
//   - Authenticated == true  -> Identity is set; the caller mints a session.
//   - Authenticated == false -> Reason explains why (for a 401 / logging).
type AuthResult struct {
	Authenticated bool
	Identity      *Identity
	Reason        string
}

// Credential is provider-agnostic auth input. Local uses Email+Password; an
// OAuth login uses Token or Code. A provider reads only what it needs.
type Credential struct {
	Email    string
	Password string
	Token    string
	Code     string
}

// IdentityProvider verifies a credential and returns a normalized AuthResult.
// It is a driving port: the app layer implements it, the ports layer calls it.
type IdentityProvider interface {
	Name() string
	Authenticate(ctx context.Context, cred Credential) (AuthResult, error)
}

// ProviderProfile is the identity a social provider returns once a token is
// verified. It is the only thing that crosses the network boundary.
type ProviderProfile struct {
	ProviderUserID string
	Email          string
	EmailVerified  bool
	DisplayName    string
}

// ProfileVerifier turns an OAuth credential into a verified ProviderProfile.
// It is a driven port: adapters (Google/Facebook) implement it.
type ProfileVerifier interface {
	Verify(ctx context.Context, cred Credential) (ProviderProfile, error)
}

// MintSession turns a verified Identity into a session token.
type MintSession func(id Identity) (string, error)

// Success / Fail construct AuthResults so providers read cleanly.
func Success(id Identity) AuthResult { return AuthResult{Authenticated: true, Identity: &id} }
func Fail(reason string) AuthResult  { return AuthResult{Authenticated: false, Reason: reason} }

// ErrUnauthenticated wraps a rejection reason as an error so the ports layer can
// map it to a 401 while keeping the human-readable reason.
type ErrUnauthenticated struct{ Reason string }

func (e *ErrUnauthenticated) Error() string { return "unauthenticated: " + e.Reason }

// ErrInvalidInput marks a bad request (e.g. a malformed email or too-short
// password); the ports layer maps it to a 400 and surfaces Reason to the client.
type ErrInvalidInput struct{ Reason string }

func (e *ErrInvalidInput) Error() string { return "invalid input: " + e.Reason }
