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
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"

	"frpg-backend/internal/adapters"
	"frpg-backend/internal/app"
	"frpg-backend/internal/domain"
	"frpg-backend/internal/ports"
)

// NewServer builds a fully wired HTTP server from the environment.
func NewServer(ctx context.Context) *ports.Server {
	repo := buildRepo(ctx)
	sessions := adapters.NewSessionManager(envOr("SESSION_SECRET", devSecret()), 24*time.Hour)

	return &ports.Server{
		Local:    app.NewLocalProvider(repo),
		Google:   app.NewOAuthProvider("google", adapters.GoogleVerifier{Audience: os.Getenv("GOOGLE_CLIENT_ID")}, repo),
		Facebook: app.NewOAuthProvider("facebook", adapters.FacebookVerifier{}, repo),
		Sessions: sessions,
	}
}

// buildRepo returns a DynamoDB-backed repository when AWS/DynamoDB is configured,
// and an in-memory seeded repository otherwise so the server runs offline in dev.
func buildRepo(ctx context.Context) domain.Repository {
	endpoint := os.Getenv("DYNAMODB_ENDPOINT")
	if endpoint == "" && os.Getenv("AWS_REGION") == "" {
		log.Println("no DynamoDB configured; using in-memory seeded repo (dev only)")
		return adapters.NewInMemorySeeded()
	}

	cfg, err := awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(envOr("AWS_REGION", "local")))
	if err != nil {
		log.Fatalf("load aws config: %v", err)
	}
	client := dynamodb.NewFromConfig(cfg, func(o *dynamodb.Options) {
		if endpoint != "" {
			o.BaseEndpoint = aws.String(endpoint)
		}
	})
	return adapters.NewDynamo(client, envOr("USERS_TABLE", "Users"))
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
