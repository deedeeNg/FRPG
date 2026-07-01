package users

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// dynamoAPI is the slice of the DynamoDB client this repo needs (easy to fake).
type dynamoAPI interface {
	GetItem(ctx context.Context, in *dynamodb.GetItemInput, opts ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error)
	PutItem(ctx context.Context, in *dynamodb.PutItemInput, opts ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error)
}

// Dynamo is a Repository backed by DynamoDB (local or AWS).
type Dynamo struct {
	client dynamoAPI
	table  string
}

// NewDynamo builds a Dynamo repository around a DynamoDB client.
func NewDynamo(client *dynamodb.Client, table string) *Dynamo {
	return &Dynamo{client: client, table: table}
}

func (r *Dynamo) GetByEmail(ctx context.Context, email string) (User, error) {
	out, err := r.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(r.table),
		Key: map[string]types.AttributeValue{
			"email": &types.AttributeValueMemberS{Value: email},
		},
	})
	if err != nil {
		return User{}, err
	}
	if out.Item == nil {
		return User{}, ErrNotFound
	}
	var u User
	if err := attributevalue.UnmarshalMap(out.Item, &u); err != nil {
		return User{}, err
	}
	return u, nil
}

func (r *Dynamo) Put(ctx context.Context, u User) error {
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
