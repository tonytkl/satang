package main

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/tonytkl/satang/clients"
	"net/http"
	"testing"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tonytkl/satang/transaction"
	"github.com/tonytkl/satang/utils"
)

type mockTransactionService struct {
	getTransactionsBetweenPeriodFunc func(ctx context.Context, ownerID string, fromDate time.Time, toDate time.Time, limit int32, cursor string) ([]transaction.Transaction, string, error)
}

func TestListTransactionsLambdaHandleServiceDeadline(t *testing.T) {
	handler := &listTransactionsLambda{service: &mockTransactionService{
		getTransactionsBetweenPeriodFunc: func(context.Context, string, time.Time, time.Time, int32, string) ([]transaction.Transaction, string, error) {
			return nil, "", context.DeadlineExceeded
		},
	}}
	response, err := handler.Handle(context.Background(), events.APIGatewayV2HTTPRequest{})
	require.NoError(t, err)
	assert.Equal(t, http.StatusGatewayTimeout, response.StatusCode)
}

func TestListTransactionsLambdaHandleInvalidRangeAndLimit(t *testing.T) {
	handler := &listTransactionsLambda{service: transaction.NewService(nil)}
	for _, query := range []map[string]string{
		{"fromDate": "2025-06-01", "toDate": "2025-05-01"},
		{"limit": "-1"},
	} {
		response, err := handler.Handle(context.Background(), events.APIGatewayV2HTTPRequest{QueryStringParameters: query})
		require.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, response.StatusCode)
	}
}

func (m *mockTransactionService) CreateTransaction(ctx context.Context, walletID, walletName, categoryID, categoryName, description, currency, imageURL, txType string, amount float64, date time.Time, ownerID string) error {
	return nil
}

func (m *mockTransactionService) GetTransaction(ctx context.Context, transactionID string, ownerID string) (*transaction.Transaction, error) {
	return nil, nil
}

func (m *mockTransactionService) GetTransactionsBetweenPeriod(ctx context.Context, ownerID string, fromDate time.Time, toDate time.Time, limit int32, cursor string) ([]transaction.Transaction, string, error) {
	if m.getTransactionsBetweenPeriodFunc != nil {
		return m.getTransactionsBetweenPeriodFunc(ctx, ownerID, fromDate, toDate, limit, cursor)
	}
	return nil, "", nil
}

func (m *mockTransactionService) ListTransactions(ctx context.Context, ownerID string, fromDate time.Time, toDate time.Time, limit int32, cursor string) ([]transaction.Transaction, string, error) {
	if m.getTransactionsBetweenPeriodFunc != nil {
		return m.getTransactionsBetweenPeriodFunc(ctx, ownerID, fromDate, toDate, limit, cursor)
	}
	return nil, "", nil
}

func (m *mockTransactionService) ListTransactionsOfCategory(ctx context.Context, ownerID string, fromDate time.Time, toDate time.Time, limit int32, cursor string, categoryID string) ([]transaction.Transaction, string, error) {
	return nil, "", nil
}

func (m *mockTransactionService) ListTransactionsOfWallet(ctx context.Context, ownerID string, fromDate time.Time, toDate time.Time, limit int32, cursor string, wallet string) ([]transaction.Transaction, string, error) {
	return nil, "", nil
}

func (m *mockTransactionService) EditTransaction(ctx context.Context, ownerID string, transactionID string, changedFields map[string]any) error {
	return nil
}

func (m *mockTransactionService) DeleteTransaction(ctx context.Context, ownerID string, transactionID string) error {
	return nil
}

func TestListTransactionsLambdaHandleSuccess(t *testing.T) {
	expectedFromDate := time.Date(2025, 5, 1, 0, 0, 0, 0, time.UTC)
	expectedToDate := time.Date(2025, 5, 31, 0, 0, 0, 0, time.UTC)
	expectedLimit := int32(10)
	expectedNextCursor := "next-token-1"

	handler := &listTransactionsLambda{
		service: &mockTransactionService{
			getTransactionsBetweenPeriodFunc: func(ctx context.Context, ownerID string, fromDate time.Time, toDate time.Time, limit int32, cursor string) ([]transaction.Transaction, string, error) {
				assert.Equal(t, "1", ownerID)
				assert.True(t, expectedFromDate.Equal(fromDate))
				assert.True(t, expectedToDate.Equal(toDate))
				assert.Equal(t, expectedLimit, limit)
				assert.Equal(t, expectedNextCursor, cursor)

				return []transaction.Transaction{{
					ID:           "tx-1",
					WalletID:     "wallet-1",
					WalletName:   "Cash",
					Type:         transaction.TransactionTypeExpense,
					Amount:       120.50,
					Currency:     "THB",
					CategoryID:   "cat-1",
					CategoryName: "Food",
					Date:         expectedFromDate,
					OwnerID:      "1",
				}}, "next-token-2", nil
			},
		},
	}

	request := events.APIGatewayV2HTTPRequest{
		QueryStringParameters: map[string]string{
			"fromDate": "2025-05-01",
			"toDate":   "2025-05-31",
			"limit":    "10",
			"cursor":   expectedNextCursor,
		},
	}

	response, err := handler.Handle(context.Background(), request)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, response.StatusCode)

	var payload listTransactionsResponse
	err = json.Unmarshal([]byte(response.Body), &payload)
	require.NoError(t, err)
	require.Len(t, payload.Transactions, 1)
	assert.Equal(t, "tx-1", payload.Transactions[0].ID)
	assert.Equal(t, "wallet-1", payload.Transactions[0].WalletID)
	assert.Equal(t, transaction.TransactionTypeExpense, payload.Transactions[0].Type)
	assert.Equal(t, "next-token-2", payload.NextCursor)
}

func TestListTransactionsLambdaHandleInvalidLimit(t *testing.T) {
	handler := &listTransactionsLambda{service: &mockTransactionService{}}

	response, err := handler.Handle(context.Background(), events.APIGatewayV2HTTPRequest{
		QueryStringParameters: map[string]string{"limit": "abc"},
	})

	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, response.StatusCode)

	var payload errorResponse
	err = json.Unmarshal([]byte(response.Body), &payload)
	require.NoError(t, err)
	assert.Equal(t, "limit must be a valid integer", payload.Message)
}

func TestListTransactionsLambdaHandleInvalidFromDate(t *testing.T) {
	handler := &listTransactionsLambda{service: &mockTransactionService{}}

	response, err := handler.Handle(context.Background(), events.APIGatewayV2HTTPRequest{
		QueryStringParameters: map[string]string{"fromDate": "invalid-date"},
	})

	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, response.StatusCode)

	var payload errorResponse
	err = json.Unmarshal([]byte(response.Body), &payload)
	require.NoError(t, err)
	assert.Contains(t, payload.Message, "invalid")
}

func TestListTransactionsLambdaHandleServiceError(t *testing.T) {
	handler := &listTransactionsLambda{
		service: &mockTransactionService{
			getTransactionsBetweenPeriodFunc: func(ctx context.Context, ownerID string, fromDate time.Time, toDate time.Time, limit int32, cursor string) ([]transaction.Transaction, string, error) {
				return nil, "", errors.New("database unavailable")
			},
		},
	}

	response, err := handler.Handle(context.Background(), events.APIGatewayV2HTTPRequest{
		QueryStringParameters: map[string]string{
			"fromDate": "2025-05-01",
			"toDate":   "2025-05-31",
		},
	})

	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, response.StatusCode)

	var payload errorResponse
	err = json.Unmarshal([]byte(response.Body), &payload)
	require.NoError(t, err)
	assert.Equal(t, "database unavailable", payload.Message)
}

func TestListTransactionsLambdaHandleInvalidCursor(t *testing.T) {
	handler := &listTransactionsLambda{
		service: &mockTransactionService{
			getTransactionsBetweenPeriodFunc: func(_ context.Context, _ string, _ time.Time, _ time.Time, _ int32, cursor string) ([]transaction.Transaction, string, error) {
				assert.Equal(t, "random", cursor)
				return nil, "", utils.ValidationError(errors.New("decode cursor failed"))
			},
		},
	}

	response, err := handler.Handle(context.Background(), events.APIGatewayV2HTTPRequest{
		QueryStringParameters: map[string]string{"cursor": "random"},
	})

	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, response.StatusCode)
}

func TestListTransactionsLambdaHandleInvalidTransactionSchema(t *testing.T) {
	handler := &listTransactionsLambda{
		service: &mockTransactionService{
			getTransactionsBetweenPeriodFunc: func(ctx context.Context, ownerID string, fromDate time.Time, toDate time.Time, limit int32, cursor string) ([]transaction.Transaction, string, error) {
				return []transaction.Transaction{{
					ID:         "",
					WalletID:   "wallet-1",
					Type:       transaction.TransactionTypeExpense,
					Amount:     10,
					Currency:   "THB",
					CategoryID: "cat-1",
					Date:       time.Date(2025, 5, 1, 0, 0, 0, 0, time.UTC),
					OwnerID:    "1",
				}}, "", nil
			},
		},
	}

	response, err := handler.Handle(context.Background(), events.APIGatewayV2HTTPRequest{
		QueryStringParameters: map[string]string{
			"fromDate": "2025-05-01",
			"toDate":   "2025-05-31",
		},
	})

	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, response.StatusCode)

	var payload errorResponse
	err = json.Unmarshal([]byte(response.Body), &payload)
	require.NoError(t, err)
	assert.Contains(t, payload.Message, "transaction response schema validation failed")
	assert.Contains(t, payload.Message, "id is required")
}

func (m *mockTransactionService) PrepareCreateTransaction(walletID string, walletName string, categoryID string, categoryName string, description string, currency string, imageURL string, txType string, amount float64, date time.Time, ownerID string) (clients.WriteOp, error) {
	return clients.WriteOp{}, nil
}
