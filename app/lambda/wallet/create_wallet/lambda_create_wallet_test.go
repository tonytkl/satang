package main

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tonytkl/satang/wallet"
)

// MockWalletService is a mock implementation of wallet.Service
// used to validate the lambda request flow.
type MockWalletService struct {
	CreateWalletFunc func(ctx context.Context, ownerID string, name string, currency string, balance float64, walletType string) error
}

func (m *MockWalletService) CreateWallet(ctx context.Context, ownerID string, name string, currency string, balance float64, walletType string) error {
	if m.CreateWalletFunc != nil {
		return m.CreateWalletFunc(ctx, ownerID, name, currency, balance, walletType)
	}
	return nil
}

func (m *MockWalletService) ListWallets(ctx context.Context, ownerID string, nextToken string, limit int32) ([]wallet.Wallet, string, error) {
	return nil, "", nil
}

func (m *MockWalletService) GetWallet(ctx context.Context, ownerID string, walletID string) (wallet.Wallet, error) {
	return wallet.Wallet{}, nil
}

func (m *MockWalletService) EditWallet(ctx context.Context, ownerID string, walletID string, changedFields map[string]any) error {
	return nil
}

func (m *MockWalletService) SetActiveWallet(ctx context.Context, ownerID string, walletID string, isActive bool) error {
	return nil
}

func TestHandle_ValidPayload(t *testing.T) {
	payload := createWalletRequest{
		Name:       "Main Wallet",
		Currency:   "USD",
		Balance:    250.0,
		WalletType: "debit",
	}

	body, err := json.Marshal(payload)
	require.NoError(t, err)

	mockService := &MockWalletService{
		CreateWalletFunc: func(ctx context.Context, ownerID string, name string, currency string, balance float64, walletType string) error {
			assert.Equal(t, "1", ownerID)
			assert.Equal(t, "Main Wallet", name)
			assert.Equal(t, "USD", currency)
			assert.Equal(t, 250.0, balance)
			assert.Equal(t, "debit", walletType)
			return nil
		},
	}

	handler := &createWalletLambda{service: mockService}
	response, err := handler.Handle(context.Background(), events.APIGatewayV2HTTPRequest{Body: string(body)})

	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, response.StatusCode)
}

func TestHandle_InvalidJSONPayload(t *testing.T) {
	handler := &createWalletLambda{service: &MockWalletService{}}
	response, err := handler.Handle(context.Background(), events.APIGatewayV2HTTPRequest{Body: "invalid json"})

	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, response.StatusCode)

	var errResp errorResponse
	err = json.Unmarshal([]byte(response.Body), &errResp)
	require.NoError(t, err)
	assert.Equal(t, "Invalid JSON payload", errResp.Message)
}

func TestHandle_MissingWalletType(t *testing.T) {
	payload := createWalletRequest{
		Name:     "Main Wallet",
		Currency: "USD",
		Balance:  25.5,
	}

	body, err := json.Marshal(payload)
	require.NoError(t, err)

	handler := &createWalletLambda{service: &MockWalletService{}}
	response, err := handler.Handle(context.Background(), events.APIGatewayV2HTTPRequest{Body: string(body)})

	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, response.StatusCode)

	var errResp errorResponse
	err = json.Unmarshal([]byte(response.Body), &errResp)
	require.NoError(t, err)
	assert.Equal(t, "type is required", errResp.Message)
}

func TestHandle_ServiceErrorReturnsBadRequest(t *testing.T) {
	payload := createWalletRequest{
		Name:       "Main Wallet",
		Currency:   "USD",
		Balance:    25.5,
		WalletType: "debit",
	}

	body, err := json.Marshal(payload)
	require.NoError(t, err)

	mockService := &MockWalletService{
		CreateWalletFunc: func(ctx context.Context, ownerID string, name string, currency string, balance float64, walletType string) error {
			return assert.AnError
		},
	}

	handler := &createWalletLambda{service: mockService}
	response, err := handler.Handle(context.Background(), events.APIGatewayV2HTTPRequest{Body: string(body)})

	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, response.StatusCode)

	var errResp errorResponse
	err = json.Unmarshal([]byte(response.Body), &errResp)
	require.NoError(t, err)
	assert.Equal(t, assert.AnError.Error(), errResp.Message)
}

var _ wallet.Service = (*MockWalletService)(nil)

func TestValidatePayload_EmptyWalletType(t *testing.T) {
	err := validatePayload(createWalletRequest{WalletType: "   "})
	require.Error(t, err)
	assert.Equal(t, "type is required", err.Error())
}

func TestValidatePayload_ValidWalletType(t *testing.T) {
	err := validatePayload(createWalletRequest{WalletType: "credit"})
	require.NoError(t, err)
}
