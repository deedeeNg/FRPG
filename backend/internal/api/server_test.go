package api_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"frpg-backend/internal/api"
	"frpg-backend/internal/auth"
	"frpg-backend/internal/session"
	"frpg-backend/internal/users"
)

func newTestServer() *api.Server {
	repo := users.NewInMemorySeeded()
	return &api.Server{
		Local:    auth.NewLocalProvider(repo),
		Google:   auth.Allow(auth.Identity{UserID: "u_google_1", Email: "googler@frpg.dev", Provider: "google"}),
		Facebook: auth.Deny("facebook not linked"),
		Sessions: session.NewManager("test-secret", time.Hour),
	}
}

func post(t *testing.T, h http.Handler, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(body))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestLoginRoute(t *testing.T) {
	h := newTestServer().Routes()

	t.Run("valid credentials -> 200 + token", func(t *testing.T) {
		rec := post(t, h, "/auth/login", `{"email":"test@frpg.dev","password":"password123"}`)
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200 (body: %s)", rec.Code, rec.Body)
		}
		var resp map[string]string
		_ = json.Unmarshal(rec.Body.Bytes(), &resp)
		if resp["token"] == "" {
			t.Fatalf("expected a token, got: %v", resp)
		}
	})

	t.Run("wrong password -> 401", func(t *testing.T) {
		rec := post(t, h, "/auth/login", `{"email":"test@frpg.dev","password":"nope"}`)
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("status = %d, want 401", rec.Code)
		}
	})
}

func TestOAuthRoute(t *testing.T) {
	h := newTestServer().Routes()

	t.Run("google (allowed) -> 200 + token", func(t *testing.T) {
		rec := post(t, h, "/auth/oauth/google", `{"token":"any-id-token"}`)
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200 (body: %s)", rec.Code, rec.Body)
		}
	})

	t.Run("facebook (denied) -> 401", func(t *testing.T) {
		rec := post(t, h, "/auth/oauth/facebook", `{"token":"any"}`)
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("status = %d, want 401", rec.Code)
		}
	})
}

// TestMeRoute exercises the full loop: log in, then use the token on a protected
// route, and confirm a missing token is rejected.
func TestMeRoute(t *testing.T) {
	h := newTestServer().Routes()

	login := post(t, h, "/auth/login", `{"email":"test@frpg.dev","password":"password123"}`)
	var resp map[string]string
	_ = json.Unmarshal(login.Body.Bytes(), &resp)
	token := resp["token"]

	req := httptest.NewRequest(http.MethodGet, "/api/me", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 (body: %s)", rec.Code, rec.Body)
	}
	var me map[string]string
	_ = json.Unmarshal(rec.Body.Bytes(), &me)
	if me["email"] != "test@frpg.dev" {
		t.Fatalf("unexpected /me payload: %v", me)
	}

	noAuth := httptest.NewRequest(http.MethodGet, "/api/me", nil)
	rec2 := httptest.NewRecorder()
	h.ServeHTTP(rec2, noAuth)
	if rec2.Code != http.StatusUnauthorized {
		t.Fatalf("missing token status = %d, want 401", rec2.Code)
	}
}
