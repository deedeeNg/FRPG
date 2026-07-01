package auth_test

import (
	"context"
	"errors"
	"testing"

	"frpg-backend/internal/auth"
)

// mintFake stands in for real session minting (JWT/cookie). It just proves the
// Identity was routed through on success.
func mintFake(id auth.Identity) (string, error) {
	return "session-for-" + id.Email, nil
}

// TestMockProvider_YesNo shows the flexible yes/no: the mock is handed whatever
// behavior the test wants, and Login routes that result into a session or a 401.
func TestMockProvider_YesNo(t *testing.T) {
	ctx := context.Background()

	t.Run("yes -> session", func(t *testing.T) {
		provider := auth.Allow(auth.Identity{
			UserID:   "u_google_1",
			Email:    "googler@frpg.dev",
			Provider: "google",
		})

		session, err := auth.Login(ctx, provider, auth.Credential{Token: "any"}, mintFake)
		if err != nil {
			t.Fatalf("expected success, got error: %v", err)
		}
		if session != "session-for-googler@frpg.dev" {
			t.Fatalf("unexpected session: %q", session)
		}
	})

	t.Run("no -> unauthenticated", func(t *testing.T) {
		provider := auth.Deny("provider rejected the token")

		_, err := auth.Login(ctx, provider, auth.Credential{Token: "bad"}, mintFake)
		var unauth *auth.ErrUnauthenticated
		if !errors.As(err, &unauth) {
			t.Fatalf("expected *ErrUnauthenticated, got: %v", err)
		}
		if unauth.Reason != "provider rejected the token" {
			t.Fatalf("unexpected reason: %q", unauth.Reason)
		}
	})

	t.Run("custom behavior via injected function", func(t *testing.T) {
		// The mock's whole behavior is a function, so a test can decide yes/no
		// however it likes — here, based on the credential contents.
		provider := auth.MockProvider{
			NameValue: "mock-conditional",
			AuthenticateFunc: func(_ context.Context, cred auth.Credential) (auth.AuthResult, error) {
				if cred.Token == "valid" {
					return auth.AuthResult{
						Authenticated: true,
						Identity:      &auth.Identity{Email: "someone@frpg.dev", Provider: "facebook"},
					}, nil
				}
				return auth.AuthResult{Authenticated: false, Reason: "token not recognized"}, nil
			},
		}

		if _, err := auth.Login(ctx, provider, auth.Credential{Token: "valid"}, mintFake); err != nil {
			t.Fatalf("expected valid token to succeed, got: %v", err)
		}
		if _, err := auth.Login(ctx, provider, auth.Credential{Token: "nope"}, mintFake); err == nil {
			t.Fatal("expected invalid token to fail")
		}
	})

	t.Run("transport error propagates (not a 401)", func(t *testing.T) {
		boom := errors.New("provider unreachable")
		provider := auth.MockProvider{
			AuthenticateFunc: func(context.Context, auth.Credential) (auth.AuthResult, error) {
				return auth.AuthResult{}, boom
			},
		}

		_, err := auth.Login(ctx, provider, auth.Credential{}, mintFake)
		if !errors.Is(err, boom) {
			t.Fatalf("expected transport error to propagate, got: %v", err)
		}
	})
}
