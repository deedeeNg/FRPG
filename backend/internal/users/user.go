package users

import (
	"context"
	"errors"
)

// ErrNotFound is returned by a Repository when no user matches.
var ErrNotFound = errors.New("user not found")

// User is the canonical account record. It supports both local (password) and
// social (provider) accounts; a social user simply has an empty PasswordHash.
type User struct {
	Email          string `dynamodbav:"email"`
	UserID         string `dynamodbav:"userId"`
	Provider       string `dynamodbav:"provider"`       // "local" | "google" | "facebook"
	ProviderUserID string `dynamodbav:"providerUserId"` // subject id from the social provider
	DisplayName    string `dynamodbav:"displayName"`
	PasswordHash   string `dynamodbav:"passwordHash"` // bcrypt; empty for social accounts
	CreatedAt      string `dynamodbav:"createdAt"`
}

// Repository is the storage seam. Tests use InMemory; production uses Dynamo.
type Repository interface {
	GetByEmail(ctx context.Context, email string) (User, error)
	Put(ctx context.Context, u User) error
}
