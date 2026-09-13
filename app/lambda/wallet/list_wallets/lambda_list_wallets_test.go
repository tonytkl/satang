package main

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
	"github.com/tonytkl/satang/wallet"
)

type mockWalletService struct {
	listWalletsFunc func(ctx context.Context, ownerID string, nextToken string, limit int32) ([]wallet.Wallet, string, error)
}

func (m *mockWalletService) CreateWallet(ctx context.Context, ownerID string, name string, currency string, balance float64, walletType string) error {
	return nil
}

func (m *mockWalletService) ListWallets(ctx context.Context, ownerID string, nextToken string, limit int32) ([]wallet.Wallet, string, error) {
	if m.listWalletsFunc != nil {
		return m.listWalletsFunc(ctx, ownerID, nextToken, limit)
	}
	return nil, "", nil
}

func (m *mockWalletService) GetWallet(ctx context.Context, ownerID string, walletID string) (wallet.Wallet, error) {
	return wallet.Wallet{}, nil
}

func (m *mockWalletService) EditWallet(ctx context.Context, ownerID string, walletID string, changedFields map[string]any) error {
	return nil
}

func (m *mockWalletService) SetActiveWallet(ctx context.Context, ownerID string, walletID string, isActive bool) error {
	return nil
}

func TestListWalletsLambdaHandleSuccess(t *testing.T) {
	expectedNextToken := "next-token-1"
	createdAt := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	updatedAt := time.Date(2026, 1, 3, 4, 5, 6, 0, time.UTC)

	handler := &listWalletsLambda{
		service: &mockWalletService{
			listWalletsFunc: func(ctx context.Context, ownerID string, nextToken string, limit int32) ([]wallet.Wallet, string, error) {
				assert.Equal(t, "1", ownerID)
				assert.Equal(t, expectedNextToken, nextToken)
				assert.Equal(t, int32(10), limit)

				return []wallet.Wallet{{
					ID:        "wallet-1",
					Name:      "Main Wallet",
					Type:      wallet.WalletTypeDebit,
					Currency:  "THB",
					OwnerID:   "1",
					CreatedAt: createdAt,
					UpdatedAt: updatedAt,
				}}, "next-token-2", nil
			},
		},
	}

	request := events.APIGatewayV2HTTPRequest{
		QueryStringParameters: map[string]string{
			"nextToken": expectedNextToken,
			"limit":     "10",
		},
	}

	response, err := handler.Handle(context.Background(), request)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, response.StatusCode)

	var payload listWalletsResponse
	err = json.Unmarshal([]byte(response.Body), &payload)
	require.NoError(t, err)
	require.Len(t, payload.Wallets, 1)
	assert.Equal(t, "wallet-1", payload.Wallets[0].ID)
	assert.Equal(t, "Main Wallet", payload.Wallets[0].Name)
	assert.Equal(t, wallet.WalletTypeDebit, payload.Wallets[0].WalletType)
	assert.Equal(t, "THB", payload.Wallets[0].Currency)
	assert.Equal(t, "1", payload.Wallets[0].OwnerID)
	assert.True(t, createdAt.Equal(payload.Wallets[0].CreatedAt))
	assert.True(t, updatedAt.Equal(payload.Wallets[0].UpdatedAt))
	assert.Equal(t, "next-token-2", payload.NextToken)
}

func TestListWalletsLambdaHandleInvalidLimit(t *testing.T) {
	handler := &listWalletsLambda{service: &mockWalletService{}}

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

func TestListWalletsLambdaHandleServiceError(t *testing.T) {
	handler := &listWalletsLambda{
		service: &mockWalletService{
			listWalletsFunc: func(ctx context.Context, ownerID string, nextToken string, limit int32) ([]wallet.Wallet, string, error) {
				return nil, "", errors.New("database unavailable")
			},
		},
	}

	response, err := handler.Handle(context.Background(), events.APIGatewayV2HTTPRequest{})

	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, response.StatusCode)

	var payload errorResponse
	err = json.Unmarshal([]byte(response.Body), &payload)
	require.NoError(t, err)
	assert.Equal(t, "database unavailable", payload.Message)
}

var _ wallet.Service = (*mockWalletService)(nil)