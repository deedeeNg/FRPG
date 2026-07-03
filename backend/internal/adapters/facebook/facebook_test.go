package facebook_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"frpg-backend/internal/adapters/facebook"
	"frpg-backend/internal/domain"
)

// fakeGraph stands in for Facebook: a debug_token endpoint reporting the token's
// app, and a /me endpoint returning a profile.
func fakeGraph(t *testing.T, appID string) (debugURL, meURL string) {
	t.Helper()
	debug := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("input_token") == "" || r.URL.Query().Get("access_token") == "" {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"app_id":"` + appID + `","is_valid":true}}`))
	}))
	me := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("access_token") == "" {
			http.Error(w, "no token", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"fb-42","name":"Zuck","email":"z@frpg.dev"}`))
	}))
	t.Cleanup(debug.Close)
	t.Cleanup(me.Close)
	return debug.URL, me.URL
}

func TestVerifier_Verify(t *testing.T) {
	t.Run("valid token for our app returns profile", func(t *testing.T) {
		dbg, me := fakeGraph(t, "our-app-id")
		v := facebook.Verifier{AppID: "our-app-id", AppToken: "our-app-id|secret", DebugTokenURL: dbg, GraphURL: me}
		got, err := v.Verify(context.Background(), domain.Credential{Token: "good"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.ProviderUserID != "fb-42" || got.Email != "z@frpg.dev" || got.DisplayName != "Zuck" {
			t.Fatalf("wrong profile: %+v", got)
		}
	})

	t.Run("token from another app is rejected", func(t *testing.T) {
		dbg, me := fakeGraph(t, "someone-elses-app")
		v := facebook.Verifier{AppID: "our-app-id", AppToken: "our-app-id|secret", DebugTokenURL: dbg, GraphURL: me}
		if _, err := v.Verify(context.Background(), domain.Credential{Token: "good"}); err == nil {
			t.Fatal("expected app mismatch error")
		}
	})

	t.Run("no app credentials skips the check (dev)", func(t *testing.T) {
		_, me := fakeGraph(t, "ignored")
		v := facebook.Verifier{GraphURL: me}
		if _, err := v.Verify(context.Background(), domain.Credential{Token: "good"}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("missing token errors", func(t *testing.T) {
		v := facebook.Verifier{}
		if _, err := v.Verify(context.Background(), domain.Credential{}); err == nil {
			t.Fatal("expected error for empty token")
		}
	})
}
