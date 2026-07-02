package ports_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"frpg-backend/internal/adapters/inmem"
	"frpg-backend/internal/adapters/jwt"
	"frpg-backend/internal/app"
	"frpg-backend/internal/domain"
	"frpg-backend/internal/ports"
)

// fakeVerifier stands in for the network call to Google/Facebook, so the routing
// tests run against real providers with only the network boundary faked.
type fakeVerifier struct {
	profile domain.ProviderProfile
	err     error
}

func (f fakeVerifier) Verify(context.Context, domain.Credential) (domain.ProviderProfile, error) {
	return f.profile, f.err
}

func newTestServer() *ports.Server {
	repo := inmem.NewSeeded() // has test@frpg.dev (local) and googler@frpg.dev (google)
	return &ports.Server{
		Identity: app.NewManager(
			app.NewLocalProvider(repo),
			app.NewOAuthProvider("google", fakeVerifier{
				profile: domain.ProviderProfile{ProviderUserID: "google-oauth2|1234567890", Email: "googler@frpg.dev", DisplayName: "Googler"},
			}, repo),
			app.NewOAuthProvider("facebook", fakeVerifier{err: errors.New("not linked")}, repo),
		),
		Sessions: jwt.NewManager("test-secret", time.Hour),
	}
}

func post(t *testing.T, h http.Handler, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(body))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestAuthRoute_Local(t *testing.T) {
	h := newTestServer().Routes()

	t.Run("valid credentials -> 200 + token", func(t *testing.T) {
		rec := post(t, h, "/auth/local", `{"email":"test@frpg.dev","password":"password123"}`)
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
		rec := post(t, h, "/auth/local", `{"email":"test@frpg.dev","password":"nope"}`)
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("status = %d, want 401", rec.Code)
		}
	})
}

func TestAuthRoute_Social(t *testing.T) {
	h := newTestServer().Routes()

	t.Run("google (verifier ok) -> 200 + token", func(t *testing.T) {
		rec := post(t, h, "/auth/google", `{"token":"any-id-token"}`)
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200 (body: %s)", rec.Code, rec.Body)
		}
	})

	t.Run("facebook (verifier rejects) -> 401", func(t *testing.T) {
		rec := post(t, h, "/auth/facebook", `{"token":"any"}`)
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("status = %d, want 401", rec.Code)
		}
	})
}

func TestAuthRoute_UnknownProvider(t *testing.T) {
	h := newTestServer().Routes()
	rec := post(t, h, "/auth/nope", `{"token":"x"}`)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rec.Code)
	}
}

// TestMeRoute exercises the full loop: log in, then use the token on a protected
// route, and confirm a missing token is rejected.
func TestMeRoute(t *testing.T) {
	h := newTestServer().Routes()

	login := post(t, h, "/auth/local", `{"email":"test@frpg.dev","password":"password123"}`)
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
