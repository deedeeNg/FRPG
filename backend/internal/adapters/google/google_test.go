package google_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"frpg-backend/internal/adapters/google"
	"frpg-backend/internal/domain"
)

// fakeGoogle stands in for Google: a tokeninfo endpoint that echoes an audience
// and a userinfo endpoint that returns a profile for a Bearer token.
func fakeGoogle(t *testing.T, aud string) (tokenInfoURL, userInfoURL string) {
	t.Helper()
	tokenInfo := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("access_token") == "" {
			http.Error(w, "no token", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"aud":"` + aud + `","azp":"` + aud + `"}`))
	}))
	userInfo := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.Header.Get("Authorization"), "Bearer ") {
			http.Error(w, "no token", http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"sub":"114","email":"g@frpg.dev","email_verified":true,"name":"Gee"}`))
	}))
	t.Cleanup(tokenInfo.Close)
	t.Cleanup(userInfo.Close)
	return tokenInfo.URL, userInfo.URL
}

func TestVerifier_Verify(t *testing.T) {
	t.Run("valid token with matching audience returns profile", func(t *testing.T) {
		ti, ui := fakeGoogle(t, "my-client-id")
		v := google.Verifier{Audience: "my-client-id", TokenInfoURL: ti, UserInfoURL: ui}
		got, err := v.Verify(context.Background(), domain.Credential{Token: "good"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.ProviderUserID != "114" || got.Email != "g@frpg.dev" || got.DisplayName != "Gee" {
			t.Fatalf("wrong profile: %+v", got)
		}
	})

	t.Run("audience mismatch is rejected", func(t *testing.T) {
		ti, ui := fakeGoogle(t, "someone-elses-client-id")
		v := google.Verifier{Audience: "my-client-id", TokenInfoURL: ti, UserInfoURL: ui}
		if _, err := v.Verify(context.Background(), domain.Credential{Token: "good"}); err == nil {
			t.Fatal("expected audience mismatch error")
		}
	})

	t.Run("empty Audience skips the check (dev)", func(t *testing.T) {
		_, ui := fakeGoogle(t, "ignored")
		v := google.Verifier{UserInfoURL: ui} // no Audience, no TokenInfoURL needed
		if _, err := v.Verify(context.Background(), domain.Credential{Token: "good"}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("missing token errors before any call", func(t *testing.T) {
		v := google.Verifier{Audience: "my-client-id"}
		if _, err := v.Verify(context.Background(), domain.Credential{}); err == nil {
			t.Fatal("expected error for empty token")
		}
	})

	t.Run("non-200 from userinfo is an error", func(t *testing.T) {
		down := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			http.Error(w, "nope", http.StatusUnauthorized)
		}))
		defer down.Close()
		v := google.Verifier{UserInfoURL: down.URL}
		if _, err := v.Verify(context.Background(), domain.Credential{Token: "x"}); err == nil {
			t.Fatal("expected error for non-200 response")
		}
	})
}
