package app_test

import (
	"context"
	"strings"
	"testing"

	"frpg-backend/internal/adapters"
	"frpg-backend/internal/app"
	"frpg-backend/internal/domain"
)

// TestSeededUserExists is the "we seeded a user" test: the canonical seed data
// lands in the repository and is retrievable by email.
func TestSeededUserExists(t *testing.T) {
	repo := adapters.NewInMemorySeeded()

	u, err := repo.GetByEmail(context.Background(), "test@frpg.dev")
	if err != nil {
		t.Fatalf("expected seeded user, got error: %v", err)
	}
	if u.UserID != "u_local_1" || u.PasswordHash == "" {
		t.Fatalf("seeded user looks wrong: %+v", u)
	}

	if _, err := repo.GetByEmail(context.Background(), "nobody@frpg.dev"); err != domain.ErrNotFound {
		t.Fatalf("expected ErrNotFound for unknown user, got: %v", err)
	}
}

func TestLocalProvider_Authenticate(t *testing.T) {
	provider := app.NewLocalProvider(adapters.NewInMemorySeeded())

	cases := []struct {
		name          string
		email         string
		password      string
		wantAuthed    bool
		wantReasonSub string
	}{
		{"correct email + correct password", "test@frpg.dev", "password123", true, ""},
		{"correct email + wrong password", "test@frpg.dev", "nope", false, "wrong password"},
		{"wrong email", "ghost@frpg.dev", "password123", false, "unknown email"},
		{"social account has no password", "googler@frpg.dev", "anything", false, "no password"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			res, err := provider.Authenticate(context.Background(), domain.Credential{
				Email:    tc.email,
				Password: tc.password,
			})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if res.Authenticated != tc.wantAuthed {
				t.Fatalf("Authenticated = %v, want %v (reason: %q)", res.Authenticated, tc.wantAuthed, res.Reason)
			}
			if tc.wantAuthed {
				if res.Identity == nil || res.Identity.Email != tc.email {
					t.Fatalf("expected identity for %s, got %+v", tc.email, res.Identity)
				}
			} else if tc.wantReasonSub != "" && !strings.Contains(res.Reason, tc.wantReasonSub) {
				t.Fatalf("reason %q does not contain %q", res.Reason, tc.wantReasonSub)
			}
		})
	}
}
