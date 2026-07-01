package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"

	"frpg-backend/internal/api"
	"frpg-backend/internal/auth"
	"frpg-backend/internal/session"
	"frpg-backend/internal/users"
)

func main() {
	ctx := context.Background()

	repo := buildRepo(ctx)
	sessions := session.NewManager(envOr("SESSION_SECRET", devSecret()), 24*time.Hour)

	server := &api.Server{
		Local:    auth.NewLocalProvider(repo),
		Google:   auth.NewGoogleProvider(os.Getenv("GOOGLE_CLIENT_ID"), repo),
		Facebook: auth.NewFacebookProvider(repo),
		Sessions: sessions,
	}

	port := envOr("PORT", "8080")
	log.Printf("backend listening on :%s", port)
	if err := http.ListenAndServe(":"+port, server.Routes()); err != nil {
		log.Fatal(err)
	}
}

// buildRepo returns a DynamoDB-backed repository when AWS/DynamoDB is configured,
// and an in-memory seeded repository otherwise so the server runs offline in dev.
func buildRepo(ctx context.Context) users.Repository {
	endpoint := os.Getenv("DYNAMODB_ENDPOINT")
	if endpoint == "" && os.Getenv("AWS_REGION") == "" {
		log.Println("no DynamoDB configured; using in-memory seeded repo (dev only)")
		return users.NewInMemorySeeded()
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
	return users.NewDynamo(client, envOr("USERS_TABLE", "Users"))
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
