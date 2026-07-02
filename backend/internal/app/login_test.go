package app_test

import (
	"context"
	"errors"
	"testing"

	"frpg-backend/internal/adapters/inmem"
	"frpg-backend/internal/app"
	"frpg-backend/internal/domain"
)

func mintFake(id domain.Identity) (string, error) {
	return "session-for-" + id.Email, nil
}

// failingRepo makes GetByEmail return a non-ErrNotFound error, to drive a
// provider's transport-error path — which Login must propagate (→500), not turn
// into an *ErrUnauthenticated (→401).
type failingRepo struct{}

func (failingRepo) GetByEmail(context.Context, string) (domain.User, error) {
	return domain.User{}, errors.New("db is down")
}
func (failingRepo) Put(context.Context, domain.User) error { return nil }

// TestLogin_Routing checks the three outcomes Login maps, using a real
// OAuthProvider (fake verifier) rather than a stand-in provider.
func TestLogin_Routing(t *testing.T) {
	ctx := context.Background()

	t.Run("authenticated -> mints a session", func(t *testing.T) {
		p := app.NewOAuthProvider("google", fakeVerifier{
			profile: domain.ProviderProfile{ProviderUserID: "g|1", Email: "googler@frpg.dev", DisplayName: "Googler"},
		}, inmem.NewSeeded())
		session, err := app.Login(ctx, p, domain.Credential{Token: "ok"}, mintFake)
		if err != nil {
			t.Fatalf("expected success, got: %v", err)
		}
		if session != "session-for-googler@frpg.dev" {
			t.Fatalf("unexpected session: %q", session)
		}
	})

	t.Run("rejected -> *ErrUnauthenticated", func(t *testing.T) {
		p := app.NewOAuthProvider("google", fakeVerifier{err: errors.New("bad token")}, inmem.New())
		_, err := app.Login(ctx, p, domain.Credential{Token: "bad"}, mintFake)
		var unauth *domain.ErrUnauthenticated
		if !errors.As(err, &unauth) {
			t.Fatalf("expected *ErrUnauthenticated, got: %v", err)
		}
	})

	t.Run("transport error -> propagates (not a 401)", func(t *testing.T) {
		p := app.NewOAuthProvider("google", fakeVerifier{
			profile: domain.ProviderProfile{ProviderUserID: "g|2", Email: "x@frpg.dev"},
		}, failingRepo{})
		_, err := app.Login(ctx, p, domain.Credential{Token: "ok"}, mintFake)
		var unauth *domain.ErrUnauthenticated
		if err == nil || errors.As(err, &unauth) {
			t.Fatalf("expected a raw transport error, got: %v", err)
		}
	})
}
