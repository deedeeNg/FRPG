package auth_test

import (
	"context"
	"errors"
	"testing"

	"frpg-backend/internal/auth"
	"frpg-backend/internal/users"
)

// fakeVerifier stands in for the network call to Google/Facebook. It returns
// exactly what a real verifier returns (a ProviderProfile, or an error for a
// bad token), so these tests drive the REAL OAuthProvider — only the network
// boundary is faked. This is the "mock as close as the real thing" the design
// aims for: swap fakeVerifier for GoogleVerifier and the provider is unchanged.
type fakeVerifier struct {
	profile auth.ProviderProfile
	err     error
}

func (f fakeVerifier) Verify(context.Context, auth.Credential) (auth.ProviderProfile, error) {
	return f.profile, f.err
}

func TestOAuthProvider_ExistingSocialUser(t *testing.T) {
	repo := users.NewInMemorySeeded() // seed includes googler@frpg.dev (google)
	google := auth.NewOAuthProvider("google", fakeVerifier{
		profile: auth.ProviderProfile{
			ProviderUserID: "google-oauth2|1234567890",
			Email:          "googler@frpg.dev",
			DisplayName:    "Googler",
		},
	}, repo)

	res, err := google.Authenticate(context.Background(), auth.Credential{Token: "valid-google-id-token"})
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
	repo := users.NewInMemory() // empty: this Google user has never signed in
	google := auth.NewOAuthProvider("google", fakeVerifier{
		profile: auth.ProviderProfile{
			ProviderUserID: "google-oauth2|999",
			Email:          "newbie@frpg.dev",
			DisplayName:    "Newbie",
		},
	}, repo)

	res, err := google.Authenticate(context.Background(), auth.Credential{Token: "valid"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !res.Authenticated {
		t.Fatalf("expected new google user to authenticate: %+v", res)
	}

	// The account should now exist in the repo, mapped to the provider identity.
	u, err := repo.GetByEmail(context.Background(), "newbie@frpg.dev")
	if err != nil {
		t.Fatalf("expected auto-provisioned user, got: %v", err)
	}
	if u.Provider != "google" || u.ProviderUserID != "google-oauth2|999" || u.PasswordHash != "" {
		t.Fatalf("provisioned user is wrong: %+v", u)
	}
}

func TestOAuthProvider_InvalidToken(t *testing.T) {
	repo := users.NewInMemory()
	google := auth.NewOAuthProvider("google", fakeVerifier{
		err: errors.New("token rejected"),
	}, repo)

	res, err := google.Authenticate(context.Background(), auth.Credential{Token: "garbage"})
	if err != nil {
		t.Fatalf("a bad token is a normal failure, not an error: %v", err)
	}
	if res.Authenticated {
		t.Fatal("expected invalid token to fail authentication")
	}
}
