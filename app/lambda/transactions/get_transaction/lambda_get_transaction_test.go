package main

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tonytkl/satang/models"
	"github.com/tonytkl/satang/repositories"
	"github.com/tonytkl/satang/schemas"
)

type mockTransactionService struct {
	getTransactionFunc func(ctx context.Context, transactionID string) (*models.Transaction, error)
}

func (m *mockTransactionService) GetTransaction(ctx context.Context, transactionID string, ownerID string) (*models.Transaction, error) {
	return m.getTransactionFunc(ctx, transactionID)
}

func (m *mockTransactionService) CreateTransaction(ctx context.Context, walletID, walletName, categoryID, categoryName, description, currency, imageURL, txType string, amount float64, date time.Time, ownerID string) error {
	return nil
}

func (m *mockTransactionService) GetTransactionsBetweenPeriod(ctx context.Context, ownerID string, fromDate time.Time, toDate time.Time, limit int32, nextToken string) ([]models.Transaction, string, error) {
	return nil, "", nil
}

func TestGetTransactionLambda_Handle(t *testing.T) {
	sampleTransaction := &models.Transaction{
		ID:           "tx123",
		WalletID:     "wallet1",
		WalletName:   "Main Wallet",
		Type:         models.TransactionTypeExpense,
		Amount:       100.0,
		Currency:     "USD",
		CategoryID:   "cat1",
		CategoryName: "Food",
		Date:         time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		OwnerID:      "user1",
		CreatedAt:    time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt:    time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
	}
	sampleSchema := schemas.TransactionSchemas{
		ID:           sampleTransaction.ID,
		WalletID:     sampleTransaction.WalletID,
		WalletName:   sampleTransaction.WalletName,
		Type:         sampleTransaction.Type,
		Amount:       sampleTransaction.Amount,
		Currency:     sampleTransaction.Currency,
		CategoryID:   sampleTransaction.CategoryID,
		CategoryName: sampleTransaction.CategoryName,
		Date:         sampleTransaction.Date,
		OwnerID:      sampleTransaction.OwnerID,
		CreatedAt:    sampleTransaction.CreatedAt,
		UpdatedAt:    sampleTransaction.UpdatedAt,
	}

	tests := []struct {
		name           string
		transactionID  string
		service        *mockTransactionService
		wantStatusCode int
		wantSchema     *schemas.TransactionSchemas
		wantBody       string
	}{
		{
			name:          "success",
			transactionID: "tx123",
			service: &mockTransactionService{
				getTransactionFunc: func(ctx context.Context, transactionID string) (*models.Transaction, error) {
					return sampleTransaction, nil
				},
			},
			wantStatusCode: 200,
			wantSchema:     &sampleSchema,
		},
		{
			name:          "not found",
			transactionID: "notfound",
			service: &mockTransactionService{
				getTransactionFunc: func(ctx context.Context, transactionID string) (*models.Transaction, error) {
					return nil, repositories.ErrTransactionNotFound
				},
			},
			wantStatusCode: 404,
			wantBody:       "transaction not found",
		},
		{
			name:          "service error",
			transactionID: "err",
			service: &mockTransactionService{
				getTransactionFunc: func(ctx context.Context, transactionID string) (*models.Transaction, error) {
					return nil, errors.New("db error")
				},
			},
			wantStatusCode: 400,
			wantBody:       "db error",
		},
		{
			name:          "missing transaction_id",
			transactionID: "",
			service: &mockTransactionService{
				getTransactionFunc: func(ctx context.Context, transactionID string) (*models.Transaction, error) {
					return nil, errors.New("Transaction ID is required")
				},
			},
			wantStatusCode: 400,
			wantBody:       "Transaction ID is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := &getTransactionLambda{service: tt.service}
			req := events.APIGatewayV2HTTPRequest{
				PathParameters: map[string]string{"transaction_id": tt.transactionID},
			}
			resp, err := handler.Handle(context.Background(), req)
			require.NoError(t, err)
			assert.Equal(t, tt.wantStatusCode, resp.StatusCode)
			if tt.wantSchema != nil {
				var got schemas.TransactionSchemas
				require.NoError(t, json.Unmarshal([]byte(resp.Body), &got))
				assert.Equal(t, *tt.wantSchema, got)
			}
			if tt.wantBody != "" {
				assert.Contains(t, resp.Body, tt.wantBody)
			}
		})
	}
}
