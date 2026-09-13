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

type mockGetWalletService struct {
	getWalletFunc func(ctx context.Context, ownerID string, walletID string) (wallet.Wallet, error)
}

func (m *mockGetWalletService) CreateWallet(ctx context.Context, ownerID string, name string, currency string, balance float64, walletType string) error {
	return nil
}

func (m *mockGetWalletService) ListWallets(ctx context.Context, ownerID string, nextToken string, limit int32) ([]wallet.Wallet, string, error) {
	return nil, "", nil
}

func (m *mockGetWalletService) GetWallet(ctx context.Context, ownerID string, walletID string) (wallet.Wallet, error) {
	if m.getWalletFunc != nil {
		return m.getWalletFunc(ctx, ownerID, walletID)
	}
	return wallet.Wallet{}, nil
}

func (m *mockGetWalletService) EditWallet(ctx context.Context, ownerID string, walletID string, changedFields map[string]any) error {
	return nil
}

func (m *mockGetWalletService) SetActiveWallet(ctx context.Context, ownerID string, walletID string, isActive bool) error {
	return nil
}

func TestGetWalletLambdaHandleSuccess(t *testing.T) {
	createdAt := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	updatedAt := time.Date(2026, 1, 3, 4, 5, 6, 0, time.UTC)

	expected := wallet.WalletRead{
		ID:         "wallet-123",
		Name:       "Main Wallet",
		WalletType: wallet.WalletTypeDebit,
		Currency:   "THB",
		OwnerID:    "1",
		CreatedAt:  createdAt,
		UpdatedAt:  updatedAt,
	}

	handler := &getWalletLambda{
		service: &mockGetWalletService{
			getWalletFunc: func(ctx context.Context, ownerID string, walletID string) (wallet.Wallet, error) {
				assert.Equal(t, "1", ownerID)
				assert.Equal(t, "wallet-123", walletID)
				return wallet.Wallet{
					ID:        "wallet-123",
					OwnerID:   "1",
					Name:      "Main Wallet",
					Type:      wallet.WalletTypeDebit,
					Currency:  "THB",
					CreatedAt: createdAt,
					UpdatedAt: updatedAt,
				}, nil
			},
		},
	}

	response, err := handler.Handle(context.Background(), events.APIGatewayV2HTTPRequest{
		PathParameters: map[string]string{"wallet_id": "wallet-123"},
	})
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, response.StatusCode)

	var actual wallet.WalletRead
	err = json.Unmarshal([]byte(response.Body), &actual)
	require.NoError(t, err)
	assert.Equal(t, expected, actual)
}

func TestGetWalletLambdaHandleMissingWalletID(t *testing.T) {
	handler := &getWalletLambda{service: &mockGetWalletService{}}

	response, err := handler.Handle(context.Background(), events.APIGatewayV2HTTPRequest{})
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, response.StatusCode)

	var payload errorResponse
	err = json.Unmarshal([]byte(response.Body), &payload)
	require.NoError(t, err)
	assert.Equal(t, "Wallet ID is required", payload.Message)
}

func TestGetWalletLambdaHandleWalletNotFound(t *testing.T) {
	handler := &getWalletLambda{
		service: &mockGetWalletService{
			getWalletFunc: func(ctx context.Context, ownerID string, walletID string) (wallet.Wallet, error) {
				return wallet.Wallet{}, wallet.ErrWalletNotFound
			},
		},
	}

	response, err := handler.Handle(context.Background(), events.APIGatewayV2HTTPRequest{
		PathParameters: map[string]string{"wallet_id": "wallet-404"},
	})
	require.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, response.StatusCode)

	var payload errorResponse
	err = json.Unmarshal([]byte(response.Body), &payload)
	require.NoError(t, err)
	assert.Equal(t, "Wallet not found", payload.Message)
}

func TestGetWalletLambdaHandleServiceError(t *testing.T) {
	handler := &getWalletLambda{
		service: &mockGetWalletService{
			getWalletFunc: func(ctx context.Context, ownerID string, walletID string) (wallet.Wallet, error) {
				return wallet.Wallet{}, errors.New("database unavailable")
			},
		},
	}

	response, err := handler.Handle(context.Background(), events.APIGatewayV2HTTPRequest{
		PathParameters: map[string]string{"wallet_id": "wallet-err"},
	})
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, response.StatusCode)

	var payload errorResponse
	err = json.Unmarshal([]byte(response.Body), &payload)
	require.NoError(t, err)
	assert.Equal(t, "database unavailable", payload.Message)
}

func TestGetWalletLambdaHandleEmptyWalletResult(t *testing.T) {
	handler := &getWalletLambda{
		service: &mockGetWalletService{
			getWalletFunc: func(ctx context.Context, ownerID string, walletID string) (wallet.Wallet, error) {
				return wallet.Wallet{}, nil
			},
		},
	}

	response, err := handler.Handle(context.Background(), events.APIGatewayV2HTTPRequest{
		PathParameters: map[string]string{"wallet_id": "wallet-empty"},
	})
	require.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, response.StatusCode)

	var payload errorResponse
	err = json.Unmarshal([]byte(response.Body), &payload)
	require.NoError(t, err)
	assert.Equal(t, "Wallet not found", payload.Message)
}

var _ wallet.Service = (*mockGetWalletService)(nil)
