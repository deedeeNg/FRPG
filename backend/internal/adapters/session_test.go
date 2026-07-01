package adapters_test

import (
	"testing"
	"time"

	"frpg-backend/internal/adapters"
)

func TestSessionManager_MintAndParse(t *testing.T) {
	m := adapters.NewSessionManager("test-secret", time.Hour)

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

func TestSessionManager_RejectsWrongSecret(t *testing.T) {
	token, err := adapters.NewSessionManager("real-secret", time.Hour).Mint("u1", "a@b.c")
	if err != nil {
		t.Fatalf("mint: %v", err)
	}
	if _, err := adapters.NewSessionManager("attacker-secret", time.Hour).Parse(token); err == nil {
		t.Fatal("expected parse to fail for token signed with a different secret")
	}
}

func TestSessionManager_RejectsExpired(t *testing.T) {
	token, err := adapters.NewSessionManager("s", -time.Minute).Mint("u1", "a@b.c") // already expired
	if err != nil {
		t.Fatalf("mint: %v", err)
	}
	if _, err := adapters.NewSessionManager("s", time.Hour).Parse(token); err == nil {
		t.Fatal("expected parse to fail for expired token")
	}
}
