package app_test

import (
	"context"
	"errors"
	"testing"

	"frpg-backend/internal/app"
	"frpg-backend/internal/domain"
)

// mintFake stands in for real session minting; it just proves the Identity was
// routed through on success.
func mintFake(id domain.Identity) (string, error) {
	return "session-for-" + id.Email, nil
}

// TestMockProvider_YesNo shows the flexible yes/no: the mock is handed whatever
// behavior the test wants, and Login routes that result into a session or a 401.
func TestMockProvider_YesNo(t *testing.T) {
	ctx := context.Background()

	t.Run("yes -> session", func(t *testing.T) {
		provider := app.Allow(domain.Identity{
			UserID:   "u_google_1",
			Email:    "googler@frpg.dev",
			Provider: "google",
		})
		session, err := app.Login(ctx, provider, domain.Credential{Token: "any"}, mintFake)
		if err != nil {
			t.Fatalf("expected success, got error: %v", err)
		}
		if session != "session-for-googler@frpg.dev" {
			t.Fatalf("unexpected session: %q", session)
		}
	})

	t.Run("no -> unauthenticated", func(t *testing.T) {
		provider := app.Deny("provider rejected the token")
		_, err := app.Login(ctx, provider, domain.Credential{Token: "bad"}, mintFake)
		var unauth *domain.ErrUnauthenticated
		if !errors.As(err, &unauth) {
			t.Fatalf("expected *ErrUnauthenticated, got: %v", err)
		}
		if unauth.Reason != "provider rejected the token" {
			t.Fatalf("unexpected reason: %q", unauth.Reason)
		}
	})

	t.Run("custom behavior via injected function", func(t *testing.T) {
		provider := app.MockProvider{
			NameValue: "mock-conditional",
			AuthenticateFunc: func(_ context.Context, cred domain.Credential) (domain.AuthResult, error) {
				if cred.Token == "valid" {
					return domain.AuthResult{
						Authenticated: true,
						Identity:      &domain.Identity{Email: "someone@frpg.dev", Provider: "facebook"},
					}, nil
				}
				return domain.AuthResult{Authenticated: false, Reason: "token not recognized"}, nil
			},
		}
		if _, err := app.Login(ctx, provider, domain.Credential{Token: "valid"}, mintFake); err != nil {
			t.Fatalf("expected valid token to succeed, got: %v", err)
		}
		if _, err := app.Login(ctx, provider, domain.Credential{Token: "nope"}, mintFake); err == nil {
			t.Fatal("expected invalid token to fail")
		}
	})

	t.Run("transport error propagates (not a 401)", func(t *testing.T) {
		boom := errors.New("provider unreachable")
		provider := app.MockProvider{
			AuthenticateFunc: func(context.Context, domain.Credential) (domain.AuthResult, error) {
				return domain.AuthResult{}, boom
			},
		}
		_, err := app.Login(ctx, provider, domain.Credential{}, mintFake)
		if !errors.Is(err, boom) {
			t.Fatalf("expected transport error to propagate, got: %v", err)
		}
	})
}
