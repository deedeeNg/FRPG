package auth

import "context"

// ErrUnauthenticated wraps a provider's failure reason as an error, so the HTTP
// layer can map it to a 401 while keeping the human-readable reason.
type ErrUnauthenticated struct{ Reason string }

func (e *ErrUnauthenticated) Error() string { return "unauthenticated: " + e.Reason }

// MintSession turns a verified Identity into an opaque session token (JWT,
// cookie value, etc.). The app owns this, so sessions stay provider-agnostic.
type MintSession func(id Identity) (string, error)

// Login is the one integration seam. Give it any provider and a credential; it
// runs the provider and routes the yes/no:
//
//   - authenticated  -> mint and return a session
//   - not authenticated -> *ErrUnauthenticated (becomes a 401)
//   - transport error   -> the raw error (becomes a 500)
//
// At integration time each route just picks the right provider and calls this;
// nothing downstream changes when you add Google or Facebook.
func Login(ctx context.Context, p IdentityProvider, cred Credential, mint MintSession) (string, error) {
	res, err := p.Authenticate(ctx, cred)
	if err != nil {
		return "", err
	}
	if !res.Authenticated {
		return "", &ErrUnauthenticated{Reason: res.Reason}
	}
	return mint(*res.Identity)
}
