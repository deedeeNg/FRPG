package auth

import "context"

// Identity is the normalized user info every provider produces on success.
// It is deliberately provider-agnostic: local, Google and Facebook all reduce
// to this shape, so the rest of the app never cares which provider was used.
type Identity struct {
	UserID         string
	Email          string
	DisplayName    string
	Provider       string // "local" | "google" | "facebook"
	ProviderUserID string // subject id from the social provider (empty for local)
}

// AuthResult is the single yes/no outcome every provider returns.
//
//   - Authenticated == true  -> Identity is set; the caller mints a session.
//   - Authenticated == false -> Reason explains why (for a 401 / logging).
//
// This is the seam the whole system routes on: providers differ wildly in how
// they verify a credential, but they all collapse to this one result, so the
// login flow only ever has to check Authenticated.
type AuthResult struct {
	Authenticated bool
	Identity      *Identity
	Reason        string
}

// Credential is provider-agnostic auth input. A local login uses Email+Password;
// an OAuth/OIDC login uses Token or Code. A provider reads only what it needs.
type Credential struct {
	Email    string
	Password string
	Token    string // OAuth/OIDC id_token or access token
	Code     string // OAuth authorization code
}

// IdentityProvider verifies a credential and returns a normalized AuthResult.
// Implementations: LocalProvider (password), MockProvider (tests), and later
// GoogleProvider / FacebookProvider (OAuth) — all interchangeable.
type IdentityProvider interface {
	Name() string
	Authenticate(ctx context.Context, cred Credential) (AuthResult, error)
}

// success / fail are small constructors so providers read cleanly.
func success(id Identity) AuthResult { return AuthResult{Authenticated: true, Identity: &id} }
func fail(reason string) AuthResult  { return AuthResult{Authenticated: false, Reason: reason} }
