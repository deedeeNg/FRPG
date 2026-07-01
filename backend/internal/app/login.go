// Package app holds the use cases: it orchestrates domain entities through the
// domain ports. It depends only on domain.
package app

import (
	"context"

	"frpg-backend/internal/domain"
)

// Login is the single authentication use case. Give it any provider and a
// credential; it runs the provider and routes the yes/no:
//
//   - authenticated     -> mint and return a session
//   - not authenticated -> *domain.ErrUnauthenticated (becomes a 401)
//   - transport error   -> the raw error (becomes a 500)
func Login(ctx context.Context, p domain.IdentityProvider, cred domain.Credential, mint domain.MintSession) (string, error) {
	res, err := p.Authenticate(ctx, cred)
	if err != nil {
		return "", err
	}
	if !res.Authenticated {
		return "", &domain.ErrUnauthenticated{Reason: res.Reason}
	}
	return mint(*res.Identity)
}
