// Package inmem implements domain.Repository with an in-memory map, used by tests
// and for offline dev. It imports nothing external.
package inmem

import (
	"context"
	"strings"
	"sync"

	"frpg-backend/internal/domain"
)

// Repository is a domain.Repository backed by a map.
type Repository struct {
	mu sync.RWMutex
	m  map[string]domain.User
}

// New returns an empty in-memory repository.
func New() *Repository {
	return &Repository{m: make(map[string]domain.User)}
}

// NewSeeded returns a repository preloaded with the canonical test users.
func NewSeeded() *Repository {
	r := New()
	for _, u := range domain.SeedUsers() {
		_ = r.Put(context.Background(), u)
	}
	return r
}

func key(email string) string { return strings.ToLower(strings.TrimSpace(email)) }

func (r *Repository) GetByEmail(_ context.Context, email string) (domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	u, ok := r.m[key(email)]
	if !ok {
		return domain.User{}, domain.ErrNotFound
	}
	return u, nil
}

func (r *Repository) Put(_ context.Context, u domain.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.m[key(u.Email)] = u
	return nil
}
