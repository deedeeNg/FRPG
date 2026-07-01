// Command seed creates the Users table in DynamoDB (local or AWS) and inserts
// the canonical test users. It shares domain.SeedUsers() with the unit tests, so
// "the seeded user" is identical in tests and in the database.
//
//	docker compose --profile seed up dynamo-seed
//	# or:  DYNAMODB_ENDPOINT=http://localhost:8000 go run ./cmd/seed
package main

import (
	"context"
	"errors"
	"log"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	"frpg-backend/internal/adapters"
	"frpg-backend/internal/domain"
)

func main() {
	ctx := context.Background()

	table := envOr("USERS_TABLE", "Users")
	endpoint := os.Getenv("DYNAMODB_ENDPOINT") // empty => real AWS

	cfg, err := awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(envOr("AWS_REGION", "local")))
	if err != nil {
		log.Fatalf("load aws config: %v", err)
	}

	client := dynamodb.NewFromConfig(cfg, func(o *dynamodb.Options) {
		if endpoint != "" {
			o.BaseEndpoint = aws.String(endpoint)
		}
	})

	if err := ensureTable(ctx, client, table); err != nil {
		log.Fatalf("ensure table: %v", err)
	}

	repo := adapters.NewDynamo(client, table)
	for _, u := range domain.SeedUsers() {
		if err := repo.Put(ctx, u); err != nil {
			log.Fatalf("put %s: %v", u.Email, err)
		}
		log.Printf("seeded %s (%s)", u.Email, u.Provider)
	}
	log.Printf("done: table %q ready", table)
}

// ensureTable creates the Users table if it does not already exist, then waits
// for it to become active.
func ensureTable(ctx context.Context, client *dynamodb.Client, table string) error {
	_, err := client.CreateTable(ctx, &dynamodb.CreateTableInput{
		TableName:   aws.String(table),
		BillingMode: types.BillingModePayPerRequest,
		AttributeDefinitions: []types.AttributeDefinition{
			{AttributeName: aws.String("email"), AttributeType: types.ScalarAttributeTypeS},
		},
		KeySchema: []types.KeySchemaElement{
			{AttributeName: aws.String("email"), KeyType: types.KeyTypeHash},
		},
	})

	var inUse *types.ResourceInUseException
	if err != nil && !errors.As(err, &inUse) {
		return err
	}
	if errors.As(err, &inUse) {
		log.Printf("table %q already exists", table)
		return nil
	}

	log.Printf("created table %q, waiting for it to become active...", table)
	waiter := dynamodb.NewTableExistsWaiter(client)
	return waiter.Wait(ctx, &dynamodb.DescribeTableInput{TableName: aws.String(table)}, 30*time.Second)
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
