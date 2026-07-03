package app_test

import (
	"context"
	"errors"
	"testing"

	"frpg-backend/internal/adapters/inmem"
	"frpg-backend/internal/app"
	"frpg-backend/internal/domain"
)

// fakeVerifier stands in for the network call to Google/Facebook, so these tests
// drive the real OAuthProvider with only the network boundary faked.
type fakeVerifier struct {
	profile domain.ProviderProfile
	err     error
}

func (f fakeVerifier) Verify(context.Context, domain.Credential) (domain.ProviderProfile, error) {
	return f.profile, f.err
}

func TestOAuthProvider_ExistingSocialUser(t *testing.T) {
	repo := inmem.NewSeeded() // seed includes googler@frpg.dev (google)
	google := app.NewOAuthProvider("google", fakeVerifier{
		profile: domain.ProviderProfile{
			ProviderUserID: "google-oauth2|1234567890",
			Email:          "googler@frpg.dev",
			EmailVerified:  true,
			DisplayName:    "Googler",
		},
	}, repo)

	res, err := google.Authenticate(context.Background(), domain.Credential{Token: "valid-google-id-token"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !res.Authenticated {
		t.Fatalf("expected existing google user to authenticate: %+v", res)
	}
	if res.Identity.Email != "googler@frpg.dev" || res.Identity.Provider != "google" {
		t.Fatalf("wrong identity: %+v", res.Identity)
	}
}

func TestOAuthProvider_AutoProvisionsNewUser(t *testing.T) {
	repo := inmem.New() // empty: this Google user has never signed in
	google := app.NewOAuthProvider("google", fakeVerifier{
		profile: domain.ProviderProfile{
			ProviderUserID: "google-oauth2|999",
			Email:          "newbie@frpg.dev",
			EmailVerified:  true,
			DisplayName:    "Newbie",
		},
	}, repo)

	res, err := google.Authenticate(context.Background(), domain.Credential{Token: "valid"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !res.Authenticated {
		t.Fatalf("expected new google user to authenticate: %+v", res)
	}

	u, err := repo.GetByEmail(context.Background(), "newbie@frpg.dev")
	if err != nil {
		t.Fatalf("expected auto-provisioned user, got: %v", err)
	}
	if u.Provider != "google" || u.ProviderUserID != "google-oauth2|999" || u.PasswordHash != "" {
		t.Fatalf("provisioned user is wrong: %+v", u)
	}
}

func TestOAuthProvider_RejectsEmailOwnedByAnotherMethod(t *testing.T) {
	repo := inmem.NewSeeded() // test@frpg.dev is a LOCAL (password) account
	// A Google identity whose email collides with the existing local account.
	google := app.NewOAuthProvider("google", fakeVerifier{
		profile: domain.ProviderProfile{
			ProviderUserID: "google-oauth2|attacker",
			Email:          "test@frpg.dev",
			EmailVerified:  true,
			DisplayName:    "Not Test",
		},
	}, repo)

	res, err := google.Authenticate(context.Background(), domain.Credential{Token: "valid"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Authenticated {
		t.Fatal("expected rejection: email belongs to a different sign-in method")
	}
	// The local account must be untouched (not clobbered / relinked).
	u, _ := repo.GetByEmail(context.Background(), "test@frpg.dev")
	if u.Provider != "local" || u.PasswordHash == "" {
		t.Fatalf("local account was altered: %+v", u)
	}
}

func TestOAuthProvider_RejectsUnverifiedEmail(t *testing.T) {
	repo := inmem.New()
	google := app.NewOAuthProvider("google", fakeVerifier{
		profile: domain.ProviderProfile{
			ProviderUserID: "google-oauth2|555",
			Email:          "unverified@frpg.dev",
			EmailVerified:  false,
			DisplayName:    "Unverified",
		},
	}, repo)

	res, err := google.Authenticate(context.Background(), domain.Credential{Token: "valid"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Authenticated {
		t.Fatal("expected rejection for an unverified provider email")
	}
	if _, err := repo.GetByEmail(context.Background(), "unverified@frpg.dev"); !errors.Is(err, domain.ErrNotFound) {
		t.Fatal("no account should have been created for an unverified email")
	}
}

func TestOAuthProvider_InvalidToken(t *testing.T) {
	repo := inmem.New()
	google := app.NewOAuthProvider("google", fakeVerifier{
		err: errors.New("token rejected"),
	}, repo)

	res, err := google.Authenticate(context.Background(), domain.Credential{Token: "garbage"})
	if err != nil {
		t.Fatalf("a bad token is a normal failure, not an error: %v", err)
	}
	if res.Authenticated {
		t.Fatal("expected invalid token to fail authentication")
	}
}
