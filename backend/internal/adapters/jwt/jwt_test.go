package jwt_test

import (
	"testing"
	"time"

	"frpg-backend/internal/adapters/jwt"
)

func TestManager_MintAndParse(t *testing.T) {
	m := jwt.NewManager("test-secret", time.Hour)

	token, err := m.Mint("u_local_1", "test@frpg.dev")
	if err != nil {
		t.Fatalf("mint: %v", err)
	}

	got, err := m.Parse(token)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if got.Subject != "u_local_1" || got.Email != "test@frpg.dev" {
		t.Fatalf("round-trip mismatch: %+v", got)
	}
}

func TestManager_RejectsWrongSecret(t *testing.T) {
	token, err := jwt.NewManager("real-secret", time.Hour).Mint("u1", "a@b.c")
	if err != nil {
		t.Fatalf("mint: %v", err)
	}
	if _, err := jwt.NewManager("attacker-secret", time.Hour).Parse(token); err == nil {
		t.Fatal("expected parse to fail for token signed with a different secret")
	}
}

func TestManager_RejectsExpired(t *testing.T) {
	token, err := jwt.NewManager("s", -time.Minute).Mint("u1", "a@b.c") // already expired
	if err != nil {
		t.Fatalf("mint: %v", err)
	}
	if _, err := jwt.NewManager("s", time.Hour).Parse(token); err == nil {
		t.Fatal("expected parse to fail for expired token")
	}
}
