package clients

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testTransaction struct {
	ID     string `dynamodbav:"id"`
	UserID string `dynamodbav:"user_id"`
	Amount int    `dynamodbav:"amount"`
}

func TestDynamoDBGetItem(t *testing.T) {
	client := newTestClient(t, func(t *testing.T, writer http.ResponseWriter, request *http.Request, payload map[string]any) {
		assert.Equal(t, "DynamoDB_20120810.GetItem", request.Header.Get("X-Amz-Target"))
		assert.Equal(t, "transactions", payload["TableName"])

		key := payload["Key"].(map[string]any)
		assert.Equal(t, "txn-1", key["id"].(map[string]any)["S"])

		writeJSON(t, writer, map[string]any{
			"Item": map[string]any{
				"id":      map[string]any{"S": "txn-1"},
				"user_id": map[string]any{"S": "user-1"},
				"amount":  map[string]any{"N": "42"},
			},
		})
	})

	var got testTransaction
	err := client.GetItem(context.Background(), "transactions", map[string]any{"id": "txn-1"}, &got)
	require.NoError(t, err)
	assert.Equal(t, testTransaction{ID: "txn-1", UserID: "user-1", Amount: 42}, got)
}

func TestDynamoDBGetItemNotFound(t *testing.T) {
	client := newTestClient(t, func(t *testing.T, writer http.ResponseWriter, request *http.Request, payload map[string]any) {
		assert.Equal(t, "DynamoDB_20120810.GetItem", request.Header.Get("X-Amz-Target"))
		assert.Equal(t, "transactions", payload["TableName"])
		writeJSON(t, writer, map[string]any{})
	})

	var got testTransaction
	err := client.GetItem(context.Background(), "transactions", map[string]any{"id": "missing"}, &got)
	require.ErrorIs(t, err, ErrItemNotFound)
}

func TestDynamoDBDeleteItem(t *testing.T) {
	client := newTestClient(t, func(t *testing.T, writer http.ResponseWriter, request *http.Request, payload map[string]any) {
		assert.Equal(t, "DynamoDB_20120810.DeleteItem", request.Header.Get("X-Amz-Target"))
		assert.Equal(t, "transactions", payload["TableName"])

		key := payload["Key"].(map[string]any)
		assert.Equal(t, "txn-1", key["id"].(map[string]any)["S"])

		writeJSON(t, writer, map[string]any{})
	})

	err := client.DeleteItem(context.Background(), "transactions", map[string]any{"id": "txn-1"})
	require.NoError(t, err)
}

func TestDynamoDBUpdateItem(t *testing.T) {
	client := newTestClient(t, func(t *testing.T, writer http.ResponseWriter, request *http.Request, payload map[string]any) {
		assert.Equal(t, "DynamoDB_20120810.UpdateItem", request.Header.Get("X-Amz-Target"))
		assert.Equal(t, "transactions", payload["TableName"])
		assert.Equal(t, "SET amount = :amount", payload["UpdateExpression"])
		assert.Equal(t, "attribute_exists(id)", payload["ConditionExpression"])

		key := payload["Key"].(map[string]any)
		assert.Equal(t, "txn-1", key["id"].(map[string]any)["S"])

		values := payload["ExpressionAttributeValues"].(map[string]any)
		assert.Equal(t, "99", values[":amount"].(map[string]any)["N"])

		writeJSON(t, writer, map[string]any{})
	})

	err := client.UpdateItem(
		context.Background(),
		"transactions",
		map[string]any{"id": "txn-1"},
		"SET amount = :amount",
		map[string]any{":amount": 99},
		"attribute_exists(id)",
	)
	require.NoError(t, err)
}

func TestDynamoDBUpdateItemWithoutOptionalFields(t *testing.T) {
	client := newTestClient(t, func(t *testing.T, writer http.ResponseWriter, request *http.Request, payload map[string]any) {
		assert.Equal(t, "DynamoDB_20120810.UpdateItem", request.Header.Get("X-Amz-Target"))
		assert.Equal(t, "transactions", payload["TableName"])
		assert.Equal(t, "SET amount = amount + :increment", payload["UpdateExpression"])

		key := payload["Key"].(map[string]any)
		assert.Equal(t, "txn-2", key["id"].(map[string]any)["S"])

		assert.NotContains(t, payload, "ExpressionAttributeValues")
		assert.NotContains(t, payload, "ConditionExpression")

		writeJSON(t, writer, map[string]any{})
	})

	err := client.UpdateItem(
		context.Background(),
		"transactions",
		map[string]any{"id": "txn-2"},
		"SET amount = amount + :increment",
		map[string]any{},
		"",
	)
	require.NoError(t, err)
}

func TestDynamoDBQueryItems(t *testing.T) {
	client := newTestClient(t, func(t *testing.T, writer http.ResponseWriter, request *http.Request, payload map[string]any) {
		assert.Equal(t, "DynamoDB_20120810.Query", request.Header.Get("X-Amz-Target"))
		assert.Equal(t, "transactions", payload["TableName"])
		assert.Equal(t, "user_id = :user_id", payload["KeyConditionExpression"])
		assert.Equal(t, "user-index", payload["IndexName"])

		values := payload["ExpressionAttributeValues"].(map[string]any)
		assert.Equal(t, "user-1", values[":user_id"].(map[string]any)["S"])

		writeJSON(t, writer, map[string]any{
			"Items": []map[string]any{
				{
					"id":      map[string]any{"S": "txn-1"},
					"user_id": map[string]any{"S": "user-1"},
					"amount":  map[string]any{"N": "42"},
				},
				{
					"id":      map[string]any{"S": "txn-2"},
					"user_id": map[string]any{"S": "user-1"},
					"amount":  map[string]any{"N": "100"},
				},
			},
		})
	})

	var got []testTransaction
	err := client.QueryItems(
		context.Background(),
		"transactions",
		"user_id = :user_id",
		map[string]any{":user_id": "user-1"},
		"user-index",
		"",
		&got,
	)
	require.NoError(t, err)

	want := []testTransaction{
		{ID: "txn-1", UserID: "user-1", Amount: 42},
		{ID: "txn-2", UserID: "user-1", Amount: 100},
	}
	assert.Equal(t, want, got)
}

func TestDynamoDBScanItems(t *testing.T) {
	client := newTestClient(t, func(t *testing.T, writer http.ResponseWriter, request *http.Request, payload map[string]any) {
		assert.Equal(t, "DynamoDB_20120810.Scan", request.Header.Get("X-Amz-Target"))
		assert.Equal(t, "transactions", payload["TableName"])
		assert.Equal(t, "user_id = :user_id", payload["FilterExpression"])

		values := payload["ExpressionAttributeValues"].(map[string]any)
		assert.Equal(t, "user-1", values[":user_id"].(map[string]any)["S"])

		writeJSON(t, writer, map[string]any{
			"Items": []map[string]any{
				{
					"id":      map[string]any{"S": "txn-1"},
					"user_id": map[string]any{"S": "user-1"},
					"amount":  map[string]any{"N": "42"},
				},
			},
		})
	})

	var got []testTransaction
	err := client.ScanItems(
		context.Background(),
		"transactions",
		"user_id = :user_id",
		map[string]any{":user_id": "user-1"},
		&got,
	)
	require.NoError(t, err)

	want := []testTransaction{{ID: "txn-1", UserID: "user-1", Amount: 42}}
	assert.Equal(t, want, got)
}

func newTestClient(t *testing.T, handler func(t *testing.T, writer http.ResponseWriter, request *http.Request, payload map[string]any)) *DynamoDB {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "application/x-amz-json-1.0")

		defer request.Body.Close()

		payload := map[string]any{}
		err := json.NewDecoder(request.Body).Decode(&payload)
		require.NoError(t, err)

		handler(t, writer, request, payload)
	}))
	t.Cleanup(server.Close)

	t.Setenv("ENVIRONMENT", "local")
	t.Setenv("AWS_DYNAMODB_ENDPOINT", server.URL)

	return NewDynamoDBClientWithConfig(aws.Config{
		Region:      "us-east-1",
		HTTPClient:  server.Client(),
		Credentials: aws.NewCredentialsCache(credentials.NewStaticCredentialsProvider("test", "test", "test")),
	})
}

func writeJSON(t *testing.T, writer http.ResponseWriter, body map[string]any) {
	t.Helper()

	err := json.NewEncoder(writer).Encode(body)
	require.NoError(t, err)
}

func equalTransactions(got, want []testTransaction) bool {
	if len(got) != len(want) {
		return false
	}

	for index := range got {
		if got[index] != want[index] {
			return false
		}
	}

	return true
}
