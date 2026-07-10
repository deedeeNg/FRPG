package inmem

import (
	"context"
	"errors"
	"testing"

	"frpg-backend/internal/domain"
)

func TestExerciseStorePutGetQuery(t *testing.T) {
	ctx := context.Background()
	s := NewExerciseStore()

	items := []domain.Exercise{
		{ID: "ex1", Level: "A1", Skill: "reading"},
		{ID: "ex2", Level: "A1", Skill: "reading"},
		{ID: "ex3", Level: "A2", Skill: "reading"},
	}
	for _, e := range items {
		if err := s.Put(ctx, e); err != nil {
			t.Fatalf("put %s: %v", e.ID, err)
		}
	}

	got, err := s.Get(ctx, "ex2")
	if err != nil || got.ID != "ex2" {
		t.Fatalf("get ex2: %v, %+v", err, got)
	}

	if _, err := s.Get(ctx, "missing"); !errors.Is(err, domain.ErrExerciseNotFound) {
		t.Fatalf("want ErrExerciseNotFound, got %v", err)
	}

	a1, err := s.Query(ctx, "A1", "reading", 10)
	if err != nil || len(a1) != 2 {
		t.Fatalf("query A1: %v, got %d items", err, len(a1))
	}

	limited, err := s.Query(ctx, "A1", "reading", 1)
	if err != nil || len(limited) != 1 {
		t.Fatalf("query A1 limit 1: %v, got %d items", err, len(limited))
	}
}

// Put is an idempotent upsert: re-putting the same id overwrites, never duplicates.
func TestExerciseStorePutIsUpsert(t *testing.T) {
	ctx := context.Background()
	s := NewExerciseStore()
	_ = s.Put(ctx, domain.Exercise{ID: "ex1", Level: "A1", Skill: "reading", Format: "multiple_choice"})
	_ = s.Put(ctx, domain.Exercise{ID: "ex1", Level: "A1", Skill: "reading", Format: "fill_blank"})

	got, _ := s.Get(ctx, "ex1")
	if got.Format != "fill_blank" {
		t.Fatalf("want overwrite to fill_blank, got %q", got.Format)
	}
	all, _ := s.Query(ctx, "A1", "reading", 10)
	if len(all) != 1 {
		t.Fatalf("want 1 item after upsert, got %d", len(all))
	}
}
