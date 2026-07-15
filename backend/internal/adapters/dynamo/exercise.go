package dynamo

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	"frpg-backend/internal/domain"
)

// exerciseAPI is the slice of the DynamoDB client the exercise store needs.
type exerciseAPI interface {
	GetItem(ctx context.Context, in *dynamodb.GetItemInput, opts ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error)
	PutItem(ctx context.Context, in *dynamodb.PutItemInput, opts ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error)
	Scan(ctx context.Context, in *dynamodb.ScanInput, opts ...func(*dynamodb.Options)) (*dynamodb.ScanOutput, error)
}

// ExerciseStore is a domain.ExerciseStore backed by DynamoDB. PK = exerciseId.
type ExerciseStore struct {
	client exerciseAPI
	table  string
}

// NewExerciseStore builds an exercise store around a DynamoDB client.
func NewExerciseStore(client *dynamodb.Client, table string) *ExerciseStore {
	return &ExerciseStore{client: client, table: table}
}

func (s *ExerciseStore) Get(ctx context.Context, id string) (domain.Exercise, error) {
	out, err := s.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(s.table),
		Key: map[string]types.AttributeValue{
			"exerciseId": &types.AttributeValueMemberS{Value: id},
		},
	})
	if err != nil {
		return domain.Exercise{}, err
	}
	if out.Item == nil {
		return domain.Exercise{}, domain.ErrExerciseNotFound
	}
	var e domain.Exercise
	if err := attributevalue.UnmarshalMap(out.Item, &e); err != nil {
		return domain.Exercise{}, err
	}
	return e, nil
}

// Query returns up to limit items for (level, skill). v0 uses a filtered Scan;
// once traffic grows this moves to the byLevelSkill GSI (PK level#skill) — the
// method signature stays the same, so callers don't change.
func (s *ExerciseStore) Query(ctx context.Context, level, skill string, limit int) ([]domain.Exercise, error) {
	out, err := s.client.Scan(ctx, &dynamodb.ScanInput{
		TableName:        aws.String(s.table),
		FilterExpression: aws.String("#lvl = :lvl AND #skl = :skl"),
		ExpressionAttributeNames: map[string]string{
			"#lvl": "level",
			"#skl": "skill",
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":lvl": &types.AttributeValueMemberS{Value: level},
			":skl": &types.AttributeValueMemberS{Value: skill},
		},
	})
	if err != nil {
		return nil, err
	}
	items := out.Items
	if limit > 0 && len(items) > limit {
		items = items[:limit]
	}
	exs := make([]domain.Exercise, 0, len(items))
	for _, item := range items {
		var e domain.Exercise
		if err := attributevalue.UnmarshalMap(item, &e); err != nil {
			return nil, err
		}
		exs = append(exs, e)
	}
	return exs, nil
}

// Put inserts or overwrites one item by exerciseId (idempotent upsert).
func (s *ExerciseStore) Put(ctx context.Context, e domain.Exercise) error {
	item, err := attributevalue.MarshalMap(e)
	if err != nil {
		return err
	}
	_, err = s.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(s.table),
		Item:      item,
	})
	return err
}
