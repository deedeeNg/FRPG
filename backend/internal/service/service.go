// Package service is the composition root: it wires adapters into use cases and
// builds the ports.Server. It is the only package that imports every layer.
package service

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	awsdynamodb "github.com/aws/aws-sdk-go-v2/service/dynamodb"

	"frpg-backend/internal/adapters/dynamo"
	"frpg-backend/internal/adapters/facebook"
	"frpg-backend/internal/adapters/google"
	"frpg-backend/internal/adapters/inmem"
	"frpg-backend/internal/adapters/jwt"
	"frpg-backend/internal/app"
	"frpg-backend/internal/domain"
	"frpg-backend/internal/ports"
)

// NewServer builds a fully wired HTTP server from the environment.
func NewServer(ctx context.Context) *ports.Server {
	repo := buildRepo(ctx)
	sessions := jwt.NewManager(envOr("SESSION_SECRET", devSecret()), 24*time.Hour)

	identity := app.NewManager(
		app.NewLocalProvider(repo),
		app.NewOAuthProvider("google", google.Verifier{Audience: os.Getenv("GOOGLE_CLIENT_ID")}, repo),
		app.NewOAuthProvider("facebook", facebook.Verifier{}, repo),
	)

	return &ports.Server{
		Identity: identity,
		Sessions: sessions,
	}
}

// buildRepo returns a DynamoDB-backed repository when AWS/DynamoDB is configured,
// and an in-memory seeded repository otherwise so the server runs offline in dev.
func buildRepo(ctx context.Context) domain.Repository {
	endpoint := os.Getenv("DYNAMODB_ENDPOINT")
	if endpoint == "" && os.Getenv("AWS_REGION") == "" {
		log.Println("no DynamoDB configured; using in-memory seeded repo (dev only)")
		return inmem.NewSeeded()
	}

	cfg, err := awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(envOr("AWS_REGION", "local")))
	if err != nil {
		log.Fatalf("load aws config: %v", err)
	}
	client := awsdynamodb.NewFromConfig(cfg, func(o *awsdynamodb.Options) {
		if endpoint != "" {
			o.BaseEndpoint = aws.String(endpoint)
		}
	})
	return dynamo.New(client, envOr("USERS_TABLE", "Users"))
}

func devSecret() string {
	log.Println("SESSION_SECRET not set; using an insecure dev secret")
	return "dev-insecure-secret-change-me"
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
