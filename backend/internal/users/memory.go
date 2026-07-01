package users

import (
	"context"
	"strings"
	"sync"
)

// InMemory is a Repository backed by a map. It is used by unit tests so auth
// logic can be exercised with no DynamoDB or network.
type InMemory struct {
	mu sync.RWMutex
	m  map[string]User
}

// NewInMemory returns an empty in-memory repository.
func NewInMemory() *InMemory {
	return &InMemory{m: make(map[string]User)}
}

// NewInMemorySeeded returns a repository preloaded with the canonical test users.
func NewInMemorySeeded() *InMemory {
	r := NewInMemory()
	for _, u := range SeedUsers() {
		_ = r.Put(context.Background(), u)
	}
	return r
}

func key(email string) string { return strings.ToLower(strings.TrimSpace(email)) }

func (r *InMemory) GetByEmail(_ context.Context, email string) (User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	u, ok := r.m[key(email)]
	if !ok {
		return User{}, ErrNotFound
	}
	return u, nil
}

func (r *InMemory) Put(_ context.Context, u User) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.m[key(u.Email)] = u
	return nil
}
