package app

import (
	"context"
	"fmt"

	"frpg-backend/internal/domain"
)

// ErrUnknownProvider is returned when a login is attempted against a provider
// name that isn't registered. The ports layer maps it to a 404.
var ErrUnknownProvider = fmt.Errorf("unknown identity provider")

// Manager is a registry of identity providers keyed by their Name(). It lets the
// delivery layer authenticate against any provider by name, so adding a provider
// is a registration change with no new route or handler.
type Manager struct {
	providers map[string]domain.IdentityProvider
}

// NewManager builds a registry from the given providers, keyed by Name().
func NewManager(providers ...domain.IdentityProvider) *Manager {
	m := &Manager{providers: make(map[string]domain.IdentityProvider, len(providers))}
	for _, p := range providers {
		m.providers[p.Name()] = p
	}
	return m
}

// Login looks up the named provider and runs the authentication use case. An
// unknown name is ErrUnknownProvider; everything else follows Login's routing
// (session on success, *domain.ErrUnauthenticated on rejection, error otherwise).
func (m *Manager) Login(ctx context.Context, name string, cred domain.Credential, mint domain.MintSession) (string, error) {
	p, ok := m.providers[name]
	if !ok {
		return "", fmt.Errorf("%w: %s", ErrUnknownProvider, name)
	}
	return Login(ctx, p, cred, mint)
}
