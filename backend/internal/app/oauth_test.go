package app_test

import (
	"context"
	"errors"
	"testing"

	"frpg-backend/internal/adapters"
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
	repo := adapters.NewInMemorySeeded() // seed includes googler@frpg.dev (google)
	google := app.NewOAuthProvider("google", fakeVerifier{
		profile: domain.ProviderProfile{
			ProviderUserID: "google-oauth2|1234567890",
			Email:          "googler@frpg.dev",
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
	repo := adapters.NewInMemory() // empty: this Google user has never signed in
	google := app.NewOAuthProvider("google", fakeVerifier{
		profile: domain.ProviderProfile{
			ProviderUserID: "google-oauth2|999",
			Email:          "newbie@frpg.dev",
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

func TestOAuthProvider_InvalidToken(t *testing.T) {
	repo := adapters.NewInMemory()
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
