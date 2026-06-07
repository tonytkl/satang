package getlisttransactions

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tonytkl/satang/model"
)

type mockTransactionService struct {
	getTransactionsBetweenPeriodFunc func(ctx context.Context, ownerID string, fromDate time.Time, toDate time.Time, limit int32, nextToken string) ([]model.Transaction, string, error)
}

func (m *mockTransactionService) CreateTransaction(ctx context.Context, walletID, walletName, categoryID, categoryName, description, currency, imageURL, txType string, amount float64, date time.Time, ownerID string) error {
	return nil
}

func (m *mockTransactionService) GetTransaction(ctx context.Context, transactionID string, ownerID string) (*model.Transaction, error) {
	return nil, nil
}

func (m *mockTransactionService) GetTransactionsBetweenPeriod(ctx context.Context, ownerID string, fromDate time.Time, toDate time.Time, limit int32, nextToken string) ([]model.Transaction, string, error) {
	if m.getTransactionsBetweenPeriodFunc != nil {
		return m.getTransactionsBetweenPeriodFunc(ctx, ownerID, fromDate, toDate, limit, nextToken)
	}
	return nil, "", nil
}

func TestGetListTransactionsLambdaHandleSuccess(t *testing.T) {
	expectedFromDate := time.Date(2025, 5, 1, 0, 0, 0, 0, time.UTC)
	expectedToDate := time.Date(2025, 5, 31, 0, 0, 0, 0, time.UTC)
	expectedLimit := int32(10)
	expectedNextToken := "next-token-1"

	handler := &getListTransactionsLambda{
		service: &mockTransactionService{
			getTransactionsBetweenPeriodFunc: func(ctx context.Context, ownerID string, fromDate time.Time, toDate time.Time, limit int32, nextToken string) ([]model.Transaction, string, error) {
				assert.Equal(t, "1", ownerID)
				assert.True(t, expectedFromDate.Equal(fromDate))
				assert.True(t, expectedToDate.Equal(toDate))
				assert.Equal(t, expectedLimit, limit)
				assert.Equal(t, expectedNextToken, nextToken)

				return []model.Transaction{{
					ID:           "tx-1",
					WalletID:     "wallet-1",
					WalletName:   "Cash",
					Type:         model.TransactionTypeExpense,
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
			"fromDate":  "2025-05-01",
			"toDate":    "2025-05-31",
			"limit":     "10",
			"nextToken": expectedNextToken,
		},
	}

	response, err := handler.Handle(context.Background(), request)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, response.StatusCode)

	var payload getListTransactionResponse
	err = json.Unmarshal([]byte(response.Body), &payload)
	require.NoError(t, err)
	require.Len(t, payload.Transactions, 1)
	assert.Equal(t, "tx-1", payload.Transactions[0].ID)
	assert.Equal(t, "next-token-2", payload.NextToken)
}

func TestGetListTransactionsLambdaHandleInvalidLimit(t *testing.T) {
	handler := &getListTransactionsLambda{service: &mockTransactionService{}}

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

func TestGetListTransactionsLambdaHandleInvalidFromDate(t *testing.T) {
	handler := &getListTransactionsLambda{service: &mockTransactionService{}}

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

func TestGetListTransactionsLambdaHandleServiceError(t *testing.T) {
	handler := &getListTransactionsLambda{
		service: &mockTransactionService{
			getTransactionsBetweenPeriodFunc: func(ctx context.Context, ownerID string, fromDate time.Time, toDate time.Time, limit int32, nextToken string) ([]model.Transaction, string, error) {
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
