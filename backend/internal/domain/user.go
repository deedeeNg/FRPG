package domain

import (
	"context"
	"errors"
)

// ErrNotFound is returned by a Repository when no user matches.
var ErrNotFound = errors.New("user not found")

// User is the canonical account entity. It supports both local (password) and
// social (provider) accounts; a social user has an empty PasswordHash.
//
// The dynamodbav tags let adapters marshal this directly; kept here for
// pragmatism rather than a separate persistence model.
type User struct {
	Email          string `dynamodbav:"email"`
	UserID         string `dynamodbav:"userId"`
	Provider       string `dynamodbav:"provider"`
	ProviderUserID string `dynamodbav:"providerUserId"`
	DisplayName    string `dynamodbav:"displayName"`
	PasswordHash   string `dynamodbav:"passwordHash"`
	CreatedAt      string `dynamodbav:"createdAt"`
}

// Repository is the driven port for user storage. Adapters implement it
// (InMemory for tests, Dynamo for production).
type Repository interface {
	GetByEmail(ctx context.Context, email string) (User, error)
	Put(ctx context.Context, u User) error
}
