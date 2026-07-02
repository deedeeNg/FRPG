package facebook_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"frpg-backend/internal/adapters/facebook"
	"frpg-backend/internal/domain"
)

func TestVerifier_Verify(t *testing.T) {
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
		v := facebook.Verifier{GraphURL: srv.URL}
		got, err := v.Verify(context.Background(), domain.Credential{Token: "good"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.ProviderUserID != "fb-42" || got.Email != "z@frpg.dev" || got.DisplayName != "Zuck" {
			t.Fatalf("wrong profile: %+v", got)
		}
	})

	t.Run("missing token errors", func(t *testing.T) {
		v := facebook.Verifier{GraphURL: srv.URL}
		if _, err := v.Verify(context.Background(), domain.Credential{}); err == nil {
			t.Fatal("expected error for empty token")
		}
	})
}
