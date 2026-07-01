// Package adapters implements the domain's driven ports: storage (DynamoDB /
// in-memory), social profile verifiers (Google / Facebook), and session signing
// (JWT). It depends only on domain.
package adapters

import (
	"context"
	"strings"
	"sync"

	"frpg-backend/internal/domain"
)

// InMemory is a domain.Repository backed by a map, used by unit tests.
type InMemory struct {
	mu sync.RWMutex
	m  map[string]domain.User
}

// NewInMemory returns an empty in-memory repository.
func NewInMemory() *InMemory {
	return &InMemory{m: make(map[string]domain.User)}
}

// NewInMemorySeeded returns a repository preloaded with the canonical test users.
func NewInMemorySeeded() *InMemory {
	r := NewInMemory()
	for _, u := range domain.SeedUsers() {
		_ = r.Put(context.Background(), u)
	}
	return r
}

func key(email string) string { return strings.ToLower(strings.TrimSpace(email)) }

func (r *InMemory) GetByEmail(_ context.Context, email string) (domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	u, ok := r.m[key(email)]
	if !ok {
		return domain.User{}, domain.ErrNotFound
	}
	return u, nil
}

func (r *InMemory) Put(_ context.Context, u domain.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.m[key(u.Email)] = u
	return nil
}
