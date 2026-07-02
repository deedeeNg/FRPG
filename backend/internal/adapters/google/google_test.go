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

// Points the real verifier at an httptest server mimicking Google's userinfo
// endpoint — exercising real request building, the Bearer header, and decoding.
func TestVerifier_Verify(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.Header.Get("Authorization"), "Bearer ") {
			http.Error(w, "no token", http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"sub":"114","email":"g@frpg.dev","name":"Gee"}`))
	}))
	defer srv.Close()

	t.Run("valid access token returns profile", func(t *testing.T) {
		v := google.Verifier{UserInfoURL: srv.URL}
		got, err := v.Verify(context.Background(), domain.Credential{Token: "good"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.ProviderUserID != "114" || got.Email != "g@frpg.dev" || got.DisplayName != "Gee" {
			t.Fatalf("wrong profile: %+v", got)
		}
	})

	t.Run("missing token errors before any call", func(t *testing.T) {
		v := google.Verifier{UserInfoURL: srv.URL}
		if _, err := v.Verify(context.Background(), domain.Credential{}); err == nil {
			t.Fatal("expected error for empty token")
		}
	})

	t.Run("non-200 from provider is an error", func(t *testing.T) {
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
