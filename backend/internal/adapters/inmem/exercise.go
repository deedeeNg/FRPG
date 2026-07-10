package inmem

import (
	"context"
	"sync"

	"frpg-backend/internal/domain"
)

// ExerciseStore is an in-memory domain.ExerciseStore for tests and offline dev.
// It mirrors the Dynamo adapter's behaviour without a database.
type ExerciseStore struct {
	mu    sync.RWMutex
	items map[string]domain.Exercise
}

// NewExerciseStore returns an empty in-memory store.
func NewExerciseStore() *ExerciseStore {
	return &ExerciseStore{items: map[string]domain.Exercise{}}
}

func (s *ExerciseStore) Get(_ context.Context, id string) (domain.Exercise, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	e, ok := s.items[id]
	if !ok {
		return domain.Exercise{}, domain.ErrExerciseNotFound
	}
	return e, nil
}

// Query returns up to limit items matching (level, skill). Order is unspecified
// (map iteration), which is fine for the "sample N items" access pattern.
func (s *ExerciseStore) Query(_ context.Context, level, skill string, limit int) ([]domain.Exercise, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]domain.Exercise, 0, limit)
	for _, e := range s.items {
		if e.Level == level && e.Skill == skill {
			out = append(out, e)
			if limit > 0 && len(out) >= limit {
				break
			}
		}
	}
	return out, nil
}

// Put inserts or overwrites by ID (idempotent upsert — same as Dynamo).
func (s *ExerciseStore) Put(_ context.Context, e domain.Exercise) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items[e.ID] = e
	return nil
}
