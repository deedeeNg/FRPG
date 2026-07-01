package app

import (
	"context"

	"frpg-backend/internal/domain"
)

// MockProvider is a fully configurable IdentityProvider for tests. Its behavior
// is whatever function you hand it, so a test can make it return a success, a
// plain failure, or a transport error — the yes/no is flexible.
type MockProvider struct {
	NameValue        string
	AuthenticateFunc func(ctx context.Context, cred domain.Credential) (domain.AuthResult, error)
}

func (m MockProvider) Name() string {
	if m.NameValue == "" {
		return "mock"
	}
	return m.NameValue
}

func (m MockProvider) Authenticate(ctx context.Context, cred domain.Credential) (domain.AuthResult, error) {
	if m.AuthenticateFunc == nil {
		return domain.Fail("mock: no behavior configured"), nil
	}
	return m.AuthenticateFunc(ctx, cred)
}

// Allow returns a provider that always authenticates as the given identity.
func Allow(id domain.Identity) MockProvider {
	return MockProvider{
		NameValue: "mock-allow",
		AuthenticateFunc: func(context.Context, domain.Credential) (domain.AuthResult, error) {
			return domain.Success(id), nil
		},
	}
}

// Deny returns a provider that always fails with the given reason.
func Deny(reason string) MockProvider {
	return MockProvider{
		NameValue: "mock-deny",
		AuthenticateFunc: func(context.Context, domain.Credential) (domain.AuthResult, error) {
			return domain.Fail(reason), nil
		},
	}
}
