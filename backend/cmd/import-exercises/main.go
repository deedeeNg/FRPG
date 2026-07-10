// Command import-exercises reads exercises as JSON Lines (from a file arg or
// stdin), ensures the Exercises table exists, and Puts each item into DynamoDB
// (local or AWS). It is the Go side of the exercises.jsonl handoff and the single
// schema authority: a line that does not fit domain.Exercise is rejected here.
//
//	go run ./cmd/gen-exercises | go run ./cmd/import-exercises
//	DYNAMODB_ENDPOINT=http://localhost:8000 go run ./cmd/import-exercises seed.jsonl
package main

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"os"
	"sort"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	"frpg-backend/internal/adapters/dynamo"
	"frpg-backend/internal/domain"
)

func main() {
	ctx := context.Background()

	table := envOr("EXERCISES_TABLE", "Exercises")
	endpoint := os.Getenv("DYNAMODB_ENDPOINT") // empty => real AWS

	// Input: a file path arg, or stdin (so `gen | import` works).
	var in io.Reader = os.Stdin
	if args := os.Args[1:]; len(args) > 0 && args[0] != "" {
		f, err := os.Open(args[0])
		if err != nil {
			log.Fatalf("open %s: %v", args[0], err)
		}
		defer f.Close()
		in = f
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

	if err := ensureTable(ctx, client, table); err != nil {
		log.Fatalf("ensure table: %v", err)
	}

	store := dynamo.NewExerciseStore(client, table)
	dist := map[string]int{}
	count := 0

	sc := bufio.NewScanner(in)
	sc.Buffer(make([]byte, 0, 1024*1024), 1024*1024) // allow long lines
	for line := 1; sc.Scan(); line++ {
		raw := sc.Bytes()
		if len(raw) == 0 {
			continue
		}
		var e domain.Exercise
		if err := json.Unmarshal(raw, &e); err != nil {
			log.Fatalf("line %d: bad json: %v", line, err)
		}
		if e.ID == "" || e.Skill == "" || e.Format == "" || e.Level == "" {
			log.Fatalf("line %d: missing required field (exerciseId/skill/format/level)", line)
		}
		if err := store.Put(ctx, e); err != nil {
			log.Fatalf("put %s: %v", e.ID, err)
		}
		dist[e.Contrast.SkillPoint]++
		count++
	}
	if err := sc.Err(); err != nil {
		log.Fatalf("read input: %v", err)
	}

	log.Printf("imported %d exercises into %q", count, table)
	for _, sp := range sortedKeys(dist) {
		log.Printf("  %-16s %d", sp, dist[sp])
	}
}

// ensureTable creates the Exercises table (PK exerciseId) if absent, then waits
// for it to become active. GSIs (byLevelSkill) are added in the serve phase.
func ensureTable(ctx context.Context, client *dynamodb.Client, table string) error {
	_, err := client.CreateTable(ctx, &dynamodb.CreateTableInput{
		TableName:   aws.String(table),
		BillingMode: types.BillingModePayPerRequest,
		AttributeDefinitions: []types.AttributeDefinition{
			{AttributeName: aws.String("exerciseId"), AttributeType: types.ScalarAttributeTypeS},
		},
		KeySchema: []types.KeySchemaElement{
			{AttributeName: aws.String("exerciseId"), KeyType: types.KeyTypeHash},
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

func sortedKeys(m map[string]int) []string {
	ks := make([]string, 0, len(m))
	for k := range m {
		ks = append(ks, k)
	}
	sort.Strings(ks)
	return ks
}
