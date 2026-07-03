package app_test

import (
	"context"
	"errors"
	"testing"

	"frpg-backend/internal/adapters/inmem"
	"frpg-backend/internal/app"
	"frpg-backend/internal/domain"
)

func TestLocalSignUp(t *testing.T) {
	ctx := context.Background()

	t.Run("creates a hashed, password-only local user", func(t *testing.T) {
		repo := inmem.New()
		id, err := app.NewLocalSignUp(repo).SignUp(ctx, "Alice@FRPG.dev", "hunter2!!")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if id.Provider != "local" || id.Email != "alice@frpg.dev" {
			t.Fatalf("wrong identity: %+v", id)
		}
		u, err := repo.GetByEmail(ctx, "alice@frpg.dev")
		if err != nil {
			t.Fatalf("user not stored: %v", err)
		}
		if u.Provider != "local" || u.ProviderUserID != "" {
			t.Fatalf("expected a local (no provider id) user: %+v", u)
		}
		if u.PasswordHash == "" || u.PasswordHash == "hunter2!!" {
			t.Fatalf("password was not hashed: %q", u.PasswordHash)
		}
	})

	t.Run("duplicate email -> ErrEmailTaken", func(t *testing.T) {
		repo := inmem.NewSeeded() // test@frpg.dev already exists
		_, err := app.NewLocalSignUp(repo).SignUp(ctx, "test@frpg.dev", "hunter2!!")
		if !errors.Is(err, app.ErrEmailTaken) {
			t.Fatalf("expected ErrEmailTaken, got: %v", err)
		}
	})

	t.Run("short password -> ErrInvalidInput", func(t *testing.T) {
		repo := inmem.New()
		_, err := app.NewLocalSignUp(repo).SignUp(ctx, "bob@frpg.dev", "short")
		var invalid *domain.ErrInvalidInput
		if !errors.As(err, &invalid) {
			t.Fatalf("expected *ErrInvalidInput, got: %v", err)
		}
	})

	t.Run("bad email -> ErrInvalidInput", func(t *testing.T) {
		repo := inmem.New()
		_, err := app.NewLocalSignUp(repo).SignUp(ctx, "not-an-email", "hunter2!!")
		var invalid *domain.ErrInvalidInput
		if !errors.As(err, &invalid) {
			t.Fatalf("expected *ErrInvalidInput, got: %v", err)
		}
	})
}
