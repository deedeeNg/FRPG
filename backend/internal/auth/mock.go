package auth

import "context"

// MockProvider is a fully configurable IdentityProvider for tests. Its behavior
// is whatever function you hand it, so a test can make it return a success (with
// an Identity), a plain failure, or a transport error — the "yes/no" is flexible.
//
// This is the exact shape a real GoogleProvider/FacebookProvider will have at
// integration time: they too just return an AuthResult. So swapping the mock for
// the real one is a one-line change at the call site.
type MockProvider struct {
	NameValue string
	// AuthenticateFunc decides the outcome. If nil, the provider denies.
	AuthenticateFunc func(ctx context.Context, cred Credential) (AuthResult, error)
}

func (m MockProvider) Name() string {
	if m.NameValue == "" {
		return "mock"
	}
	return m.NameValue
}

func (m MockProvider) Authenticate(ctx context.Context, cred Credential) (AuthResult, error) {
	if m.AuthenticateFunc == nil {
		return fail("mock: no behavior configured"), nil
	}
	return m.AuthenticateFunc(ctx, cred)
}

// Allow returns a provider that always authenticates as the given identity.
func Allow(id Identity) MockProvider {
	return MockProvider{
		NameValue: "mock-allow",
		AuthenticateFunc: func(context.Context, Credential) (AuthResult, error) {
			return success(id), nil
		},
	}
}

// Deny returns a provider that always fails with the given reason.
func Deny(reason string) MockProvider {
	return MockProvider{
		NameValue: "mock-deny",
		AuthenticateFunc: func(context.Context, Credential) (AuthResult, error) {
			return fail(reason), nil
		},
	}
}
