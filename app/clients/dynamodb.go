// Package clients provides repository-friendly AWS clients for Satang.
//
// This package includes a DynamoDB wrapper that hides low-level AWS SDK
// implementation details and exposes a small, testable interface.
package clients

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/joho/godotenv"
)

// DynamoDBClient defines the public methods supported by the DynamoDB wrapper.
// Repository implementations can depend on this interface so the AWS SDK is
// isolated from higher-level application code.
type DynamoDBClient interface {
	PutItem(ctx context.Context, table string, item any) error
	GetItem(ctx context.Context, table string, key map[string]any, out any) error
	DeleteItem(ctx context.Context, table string, key map[string]any) error
	QueryItems(ctx context.Context, table string, keyConditionExpression string, expressionValues map[string]any, indexName string, out any) error
	ScanItems(ctx context.Context, table string, filterExpression string, expressionValues map[string]any, out any) error
}

// DynamoDB is a thin wrapper around the AWS SDK v2 DynamoDB client.
// It marshals items using the attributevalue package and translates DynamoDB
// operations into repository-friendly method calls.
type DynamoDB struct {
	client *dynamodb.Client
}

// ErrItemNotFound is returned when a GetItem request returns no item.
var ErrItemNotFound = errors.New("dynamodb: item not found")

var loadDotenvOnce sync.Once
var loadDotenvErr error

// NewDynamoDBClient creates a DynamoDB wrapper using the default AWS config.
// It loads credentials and region from the environment, shared config, or IAM.
func NewDynamoDBClient(ctx context.Context) (*DynamoDB, error) {
	if err := loadDotEnv(); err != nil {
		return nil, err
	}

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("load aws config: %w", err)
	}

	return &DynamoDB{client: dynamodb.NewFromConfig(cfg, func(o *dynamodb.Options) {
		if endpoint, ok := localDynamoDBEndpoint(); ok {
			o.BaseEndpoint = aws.String(endpoint)
		}
	})}, nil
}

// NewDynamoDBClientWithConfig creates a DynamoDB wrapper from an existing AWS config.
// Use this when the AWS configuration is created externally or modified before use.
func NewDynamoDBClientWithConfig(cfg aws.Config) *DynamoDB {
	_ = loadDotEnv()

	return &DynamoDB{client: dynamodb.NewFromConfig(cfg, func(o *dynamodb.Options) {
		if endpoint, ok := localDynamoDBEndpoint(); ok {
			o.BaseEndpoint = aws.String(endpoint)
		}
	})}
}

// PutItem writes the provided item into the given table.
// The item may be any Go value that attributevalue.MarshalMap can encode.
func (d *DynamoDB) PutItem(ctx context.Context, table string, item any) error {
	attrMap, err := attributevalue.MarshalMap(item)
	if err != nil {
		return fmt.Errorf("marshal item: %w", err)
	}

	_, err = d.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: &table,
		Item:      attrMap,
	})
	if err != nil {
		return fmt.Errorf("put item: %w", err)
	}

	return nil
}

// GetItem retrieves a single item from DynamoDB and unmarshals it into out.
// The out parameter must be a pointer to the desired destination type.
func (d *DynamoDB) GetItem(ctx context.Context, table string, key map[string]any, out any) error {
	attrKey, err := attributevalue.MarshalMap(key)
	if err != nil {
		return fmt.Errorf("marshal key: %w", err)
	}

	result, err := d.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: &table,
		Key:       attrKey,
	})
	if err != nil {
		return fmt.Errorf("get item: %w", err)
	}

	if len(result.Item) == 0 {
		return ErrItemNotFound
	}

	if err := attributevalue.UnmarshalMap(result.Item, out); err != nil {
		return fmt.Errorf("unmarshal item: %w", err)
	}

	return nil
}

// DeleteItem removes an item from the specified table by its key.
func (d *DynamoDB) DeleteItem(ctx context.Context, table string, key map[string]any) error {
	attrKey, err := attributevalue.MarshalMap(key)
	if err != nil {
		return fmt.Errorf("marshal key: %w", err)
	}

	_, err = d.client.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: &table,
		Key:       attrKey,
	})
	if err != nil {
		return fmt.Errorf("delete item: %w", err)
	}

	return nil
}

// QueryItems runs a DynamoDB Query operation and unmarshals the matching items into out.
// expressionValues should map placeholders like ":pk" to concrete values.
func (d *DynamoDB) QueryItems(ctx context.Context, table, keyConditionExpression string, expressionValues map[string]any, indexName string, out any) error {
	attrValues, err := marshalExpressionValues(expressionValues)
	if err != nil {
		return fmt.Errorf("marshal expression values: %w", err)
	}

	input := &dynamodb.QueryInput{
		TableName:                 &table,
		KeyConditionExpression:    &keyConditionExpression,
		ExpressionAttributeValues: attrValues,
	}
	if indexName != "" {
		input.IndexName = &indexName
	}

	result, err := d.client.Query(ctx, input)
	if err != nil {
		return fmt.Errorf("query items: %w", err)
	}

	if err := attributevalue.UnmarshalListOfMaps(result.Items, out); err != nil {
		return fmt.Errorf("unmarshal query results: %w", err)
	}

	return nil
}

// ScanItems runs a DynamoDB Scan operation and unmarshals the matching items into out.
// If filterExpression is empty, the scan returns all items from the table.
func (d *DynamoDB) ScanItems(ctx context.Context, table, filterExpression string, expressionValues map[string]any, out any) error {
	attrValues, err := marshalExpressionValues(expressionValues)
	if err != nil {
		return fmt.Errorf("marshal expression values: %w", err)
	}

	input := &dynamodb.ScanInput{
		TableName:                 &table,
		ExpressionAttributeValues: attrValues,
	}
	if filterExpression != "" {
		input.FilterExpression = &filterExpression
	}

	result, err := d.client.Scan(ctx, input)
	if err != nil {
		return fmt.Errorf("scan items: %w", err)
	}

	if err := attributevalue.UnmarshalListOfMaps(result.Items, out); err != nil {
		return fmt.Errorf("unmarshal scan results: %w", err)
	}

	return nil
}

func marshalExpressionValues(values map[string]any) (map[string]types.AttributeValue, error) {
	attrValues := make(map[string]types.AttributeValue, len(values))
	for key, value := range values {
		attrValue, err := attributevalue.Marshal(value)
		if err != nil {
			return nil, fmt.Errorf("marshal expression value %s: %w", key, err)
		}
		attrValues[key] = attrValue
	}
	return attrValues, nil
}

func loadDotEnv() error {
	loadDotenvOnce.Do(func() {
		if _, err := os.Stat(".env"); err != nil {
			if os.IsNotExist(err) {
				return
			}

			loadDotenvErr = fmt.Errorf("stat .env: %w", err)
			return
		}

		if err := godotenv.Load(); err != nil {
			loadDotenvErr = fmt.Errorf("load .env: %w", err)
		}
	})

	return loadDotenvErr
}

func localDynamoDBEndpoint() (string, bool) {
	if os.Getenv("ENVIRONMENT") != "local" {
		return "", false
	}

	endpoint := os.Getenv("AWS_DYNAMODB_ENDPOINT")
	if endpoint == "" {
		return "http://localhost:8000", true
	}

	return endpoint, true
}
