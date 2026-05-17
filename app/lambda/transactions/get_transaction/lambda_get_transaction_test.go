package main

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tonytkl/satang/model"
	"github.com/tonytkl/satang/repositories"
)

type mockTransactionService struct {
	getTransactionFunc func(ctx context.Context, transactionID string) (*model.Transaction, error)
}

func (m *mockTransactionService) GetTransaction(ctx context.Context, transactionID string) (*model.Transaction, error) {
	return m.getTransactionFunc(ctx, transactionID)
}

func (m *mockTransactionService) CreateTransaction(ctx context.Context, walletID, walletName, categoryID, categoryName, description, currency, imageURL, txType string, amount float64, date time.Time, ownerID string) error {
	return nil
}

func TestGetTransactionLambda_Handle(t *testing.T) {
	sampleTransaction := &model.Transaction{
		ID:           "tx123",
		WalletID:     "wallet1",
		WalletName:   "Main Wallet",
		Amount:       100.0,
		Currency:     "USD",
		CategoryID:   "cat1",
		CategoryName: "Food",
		OwnerID:      "user1",
	}

	tests := []struct {
		name           string
		transactionID  string
		service        *mockTransactionService
		wantStatusCode int
		wantBody       string
	}{
		{
			name:          "success",
			transactionID: "tx123",
			service: &mockTransactionService{
				getTransactionFunc: func(ctx context.Context, transactionID string) (*model.Transaction, error) {
					return sampleTransaction, nil
				},
			},
			wantStatusCode: 200,
			wantBody:       "\"tx123\"", // Only check for transaction ID in body
		},
		{
			name:          "not found",
			transactionID: "notfound",
			service: &mockTransactionService{
				getTransactionFunc: func(ctx context.Context, transactionID string) (*model.Transaction, error) {
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
				getTransactionFunc: func(ctx context.Context, transactionID string) (*model.Transaction, error) {
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
				getTransactionFunc: func(ctx context.Context, transactionID string) (*model.Transaction, error) {
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
			assert.Contains(t, resp.Body, tt.wantBody)
		})
	}
}
