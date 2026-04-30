package clients

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	helper "github.com/tonytkl/satang/tests/utils"
)

type testTransaction struct {
	ID     string `dynamodbav:"id"`
	UserID string `dynamodbav:"user_id"`
	Amount int    `dynamodbav:"amount"`
}

func TestDynamoDBGetItem(t *testing.T) {
	client := newTestClient(t, func(t *testing.T, writer http.ResponseWriter, request *http.Request, payload map[string]any) {
		helper.AssertHeader(t, request, "X-Amz-Target", "DynamoDB_20120810.GetItem")
		helper.AssertEqual(t, payload["TableName"], "transactions")

		key := payload["Key"].(map[string]any)
		helper.AssertEqual(t, key["id"].(map[string]any)["S"], "txn-1")

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
	if err != nil {
		t.Fatalf("GetItem returned error: %v", err)
	}

	if got != (testTransaction{ID: "txn-1", UserID: "user-1", Amount: 42}) {
		t.Fatalf("GetItem returned unexpected item: %#v", got)
	}
}

func TestDynamoDBGetItemNotFound(t *testing.T) {
	client := newTestClient(t, func(t *testing.T, writer http.ResponseWriter, request *http.Request, payload map[string]any) {
		helper.AssertHeader(t, request, "X-Amz-Target", "DynamoDB_20120810.GetItem")
		helper.AssertEqual(t, payload["TableName"], "transactions")
		writeJSON(t, writer, map[string]any{})
	})

	var got testTransaction
	err := client.GetItem(context.Background(), "transactions", map[string]any{"id": "missing"}, &got)
	if err != ErrItemNotFound {
		t.Fatalf("GetItem returned %v, want %v", err, ErrItemNotFound)
	}
}

func TestDynamoDBDeleteItem(t *testing.T) {
	client := newTestClient(t, func(t *testing.T, writer http.ResponseWriter, request *http.Request, payload map[string]any) {
		helper.AssertHeader(t, request, "X-Amz-Target", "DynamoDB_20120810.DeleteItem")
		helper.AssertEqual(t, payload["TableName"], "transactions")

		key := payload["Key"].(map[string]any)
		helper.AssertEqual(t, key["id"].(map[string]any)["S"], "txn-1")

		writeJSON(t, writer, map[string]any{})
	})

	err := client.DeleteItem(context.Background(), "transactions", map[string]any{"id": "txn-1"})
	if err != nil {
		t.Fatalf("DeleteItem returned error: %v", err)
	}
}

func TestDynamoDBUpdateItem(t *testing.T) {
	client := newTestClient(t, func(t *testing.T, writer http.ResponseWriter, request *http.Request, payload map[string]any) {
		helper.AssertHeader(t, request, "X-Amz-Target", "DynamoDB_20120810.UpdateItem")
		helper.AssertEqual(t, payload["TableName"], "transactions")
		helper.AssertEqual(t, payload["UpdateExpression"], "SET amount = :amount")
		helper.AssertEqual(t, payload["ConditionExpression"], "attribute_exists(id)")

		key := payload["Key"].(map[string]any)
		helper.AssertEqual(t, key["id"].(map[string]any)["S"], "txn-1")

		values := payload["ExpressionAttributeValues"].(map[string]any)
		helper.AssertEqual(t, values[":amount"].(map[string]any)["N"], "99")

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
	if err != nil {
		t.Fatalf("UpdateItem returned error: %v", err)
	}
}

func TestDynamoDBUpdateItemWithoutOptionalFields(t *testing.T) {
	client := newTestClient(t, func(t *testing.T, writer http.ResponseWriter, request *http.Request, payload map[string]any) {
		helper.AssertHeader(t, request, "X-Amz-Target", "DynamoDB_20120810.UpdateItem")
		helper.AssertEqual(t, payload["TableName"], "transactions")
		helper.AssertEqual(t, payload["UpdateExpression"], "SET amount = amount + :increment")

		key := payload["Key"].(map[string]any)
		helper.AssertEqual(t, key["id"].(map[string]any)["S"], "txn-2")

		if _, ok := payload["ExpressionAttributeValues"]; ok {
			t.Fatalf("ExpressionAttributeValues should be omitted when empty")
		}
		if _, ok := payload["ConditionExpression"]; ok {
			t.Fatalf("ConditionExpression should be omitted when empty")
		}

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
	if err != nil {
		t.Fatalf("UpdateItem returned error: %v", err)
	}
}

func TestDynamoDBQueryItems(t *testing.T) {
	client := newTestClient(t, func(t *testing.T, writer http.ResponseWriter, request *http.Request, payload map[string]any) {
		helper.AssertHeader(t, request, "X-Amz-Target", "DynamoDB_20120810.Query")
		helper.AssertEqual(t, payload["TableName"], "transactions")
		helper.AssertEqual(t, payload["KeyConditionExpression"], "user_id = :user_id")
		helper.AssertEqual(t, payload["IndexName"], "user-index")

		values := payload["ExpressionAttributeValues"].(map[string]any)
		helper.AssertEqual(t, values[":user_id"].(map[string]any)["S"], "user-1")

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
	if err != nil {
		t.Fatalf("QueryItems returned error: %v", err)
	}

	want := []testTransaction{
		{ID: "txn-1", UserID: "user-1", Amount: 42},
		{ID: "txn-2", UserID: "user-1", Amount: 100},
	}
	if !equalTransactions(got, want) {
		t.Fatalf("QueryItems returned unexpected items: %#v", got)
	}
}

func TestDynamoDBScanItems(t *testing.T) {
	client := newTestClient(t, func(t *testing.T, writer http.ResponseWriter, request *http.Request, payload map[string]any) {
		helper.AssertHeader(t, request, "X-Amz-Target", "DynamoDB_20120810.Scan")
		helper.AssertEqual(t, payload["TableName"], "transactions")
		helper.AssertEqual(t, payload["FilterExpression"], "user_id = :user_id")

		values := payload["ExpressionAttributeValues"].(map[string]any)
		helper.AssertEqual(t, values[":user_id"].(map[string]any)["S"], "user-1")

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
	if err != nil {
		t.Fatalf("ScanItems returned error: %v", err)
	}

	want := []testTransaction{{ID: "txn-1", UserID: "user-1", Amount: 42}}
	if !equalTransactions(got, want) {
		t.Fatalf("ScanItems returned unexpected items: %#v", got)
	}
}

func newTestClient(t *testing.T, handler func(t *testing.T, writer http.ResponseWriter, request *http.Request, payload map[string]any)) *DynamoDB {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "application/x-amz-json-1.0")

		defer request.Body.Close()

		payload := map[string]any{}
		if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
			t.Fatalf("decode request body: %v", err)
		}

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

	if err := json.NewEncoder(writer).Encode(body); err != nil {
		t.Fatalf("encode response body: %v", err)
	}
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
