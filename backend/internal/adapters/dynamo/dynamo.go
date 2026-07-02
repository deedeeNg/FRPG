// Package dynamo implements domain.Repository on DynamoDB (local or AWS).
package dynamo

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	"frpg-backend/internal/domain"
)

// api is the slice of the DynamoDB client this repo needs (easy to fake).
type api interface {
	GetItem(ctx context.Context, in *dynamodb.GetItemInput, opts ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error)
	PutItem(ctx context.Context, in *dynamodb.PutItemInput, opts ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error)
}

// Repository is a domain.Repository backed by DynamoDB.
type Repository struct {
	client api
	table  string
}

// New builds a Dynamo repository around a DynamoDB client.
func New(client *dynamodb.Client, table string) *Repository {
	return &Repository{client: client, table: table}
}

func (r *Repository) GetByEmail(ctx context.Context, email string) (domain.User, error) {
	out, err := r.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(r.table),
		Key: map[string]types.AttributeValue{
			"email": &types.AttributeValueMemberS{Value: email},
		},
	})
	if err != nil {
		return domain.User{}, err
	}
	if out.Item == nil {
		return domain.User{}, domain.ErrNotFound
	}
	var u domain.User
	if err := attributevalue.UnmarshalMap(out.Item, &u); err != nil {
		return domain.User{}, err
	}
	return u, nil
}

func (r *Repository) Put(ctx context.Context, u domain.User) error {
	if u.Email == "" {
		return errors.New("user email is required")
	}
	item, err := attributevalue.MarshalMap(u)
	if err != nil {
		return err
	}
	_, err = r.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(r.table),
		Item:      item,
	})
	return err
}
