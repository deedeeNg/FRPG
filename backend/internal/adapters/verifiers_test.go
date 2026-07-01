package adapters_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"frpg-backend/internal/adapters"
	"frpg-backend/internal/domain"
)

// These tests point the real verifiers at an httptest server that mimics
// Google's tokeninfo / Facebook's Graph API — exercising the real HTTP request
// building, JSON decoding, and audience check without reaching the network.

func TestGoogleVerifier_Verify(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("id_token") == "" {
			http.Error(w, "no token", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"sub":"114","email":"g@frpg.dev","name":"Gee","aud":"my-client-id"}`))
	}))
	defer srv.Close()

	t.Run("valid token with matching audience", func(t *testing.T) {
		v := adapters.GoogleVerifier{TokenInfoURL: srv.URL, Audience: "my-client-id"}
		got, err := v.Verify(context.Background(), domain.Credential{Token: "good"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.ProviderUserID != "114" || got.Email != "g@frpg.dev" || got.DisplayName != "Gee" {
			t.Fatalf("wrong profile: %+v", got)
		}
	})

	t.Run("audience mismatch is rejected", func(t *testing.T) {
		v := adapters.GoogleVerifier{TokenInfoURL: srv.URL, Audience: "someone-elses-client-id"}
		if _, err := v.Verify(context.Background(), domain.Credential{Token: "good"}); err == nil {
			t.Fatal("expected audience mismatch error")
		}
	})

	t.Run("missing token errors before any call", func(t *testing.T) {
		v := adapters.GoogleVerifier{TokenInfoURL: srv.URL}
		if _, err := v.Verify(context.Background(), domain.Credential{}); err == nil {
			t.Fatal("expected error for empty token")
		}
	})

	t.Run("non-200 from provider is an error", func(t *testing.T) {
		down := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			http.Error(w, "nope", http.StatusUnauthorized)
		}))
		defer down.Close()
		v := adapters.GoogleVerifier{TokenInfoURL: down.URL}
		if _, err := v.Verify(context.Background(), domain.Credential{Token: "x"}); err == nil {
			t.Fatal("expected error for non-200 response")
		}
	})
}

func TestFacebookVerifier_Verify(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("access_token") == "" {
			http.Error(w, "no token", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"fb-42","name":"Zuck","email":"z@frpg.dev"}`))
	}))
	defer srv.Close()

	t.Run("valid access token returns profile", func(t *testing.T) {
		v := adapters.FacebookVerifier{GraphURL: srv.URL}
		got, err := v.Verify(context.Background(), domain.Credential{Token: "good"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.ProviderUserID != "fb-42" || got.Email != "z@frpg.dev" || got.DisplayName != "Zuck" {
			t.Fatalf("wrong profile: %+v", got)
		}
	})

	t.Run("missing token errors", func(t *testing.T) {
		v := adapters.FacebookVerifier{GraphURL: srv.URL}
		if _, err := v.Verify(context.Background(), domain.Credential{}); err == nil {
			t.Fatal("expected error for empty token")
		}
	})
}
