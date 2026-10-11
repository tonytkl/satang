package main

import (
	"context"
	"encoding/json"
	"github.com/tonytkl/satang/clients"
	"net/http"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tonytkl/satang/wallet"
)

type mockWalletService struct {
	editWalletFunc func(ctx context.Context, ownerID string, walletID string, changedFields map[string]any) (wallet.Wallet, error)
}

func (service *mockWalletService) CreateWallet(context.Context, string, string, string, float64, string) (wallet.Wallet, error) {
	return wallet.Wallet{}, nil
}

func (service *mockWalletService) ListWallets(context.Context, string, string, int32) ([]wallet.Wallet, string, error) {
	return nil, "", nil
}

func (service *mockWalletService) GetWallet(context.Context, string, string) (wallet.Wallet, error) {
	return wallet.Wallet{}, nil
}

func (service *mockWalletService) EditWallet(ctx context.Context, ownerID string, walletID string, changedFields map[string]any) (wallet.Wallet, error) {
	if service.editWalletFunc != nil {
		return service.editWalletFunc(ctx, ownerID, walletID, changedFields)
	}
	return wallet.Wallet{}, nil
}

func (service *mockWalletService) SetActiveWallet(context.Context, string, string, bool) error {
	return nil
}

var _ wallet.Service = (*mockWalletService)(nil)

func TestHandle_ValidPayload(t *testing.T) {
	handler := &editWalletLambda{service: &mockWalletService{
		editWalletFunc: func(_ context.Context, ownerID string, walletID string, changedFields map[string]any) (wallet.Wallet, error) {
			assert.Equal(t, "1", ownerID)
			assert.Equal(t, "wallet-1", walletID)
			assert.Equal(t, map[string]any{"Name": "Travel", "Type": "credit"}, changedFields)
			return wallet.Wallet{ID: walletID, OwnerID: ownerID, Name: "Travel", Type: wallet.WalletTypeCredit}, nil
		},
	}}

	response, err := handler.Handle(context.Background(), events.APIGatewayV2HTTPRequest{
		PathParameters: map[string]string{"wallet_id": " wallet-1 "},
		Body:           `{"name":"Travel","walletType":"credit"}`,
	})

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, response.StatusCode)
	assert.Equal(t, "application/json", response.Headers["Content-Type"])

	var result wallet.WalletRead
	require.NoError(t, json.Unmarshal([]byte(response.Body), &result))
	assert.Equal(t, "wallet-1", result.ID)
	assert.Equal(t, "Travel", result.Name)
	assert.Equal(t, wallet.WalletTypeCredit, result.WalletType)
}

func TestHandle_MissingWalletID(t *testing.T) {
	handler := &editWalletLambda{service: &mockWalletService{}}
	response, err := handler.Handle(context.Background(), events.APIGatewayV2HTTPRequest{
		PathParameters: map[string]string{"wallet_id": "   "},
		Body:           `{}`,
	})

	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, response.StatusCode)
	assertErrorMessage(t, response, "Wallet ID is required")
}

func TestHandle_InvalidJSONPayload(t *testing.T) {
	handler := &editWalletLambda{service: &mockWalletService{}}
	response, err := handler.Handle(context.Background(), events.APIGatewayV2HTTPRequest{
		PathParameters: map[string]string{"wallet_id": "wallet-1"},
		Body:           `{invalid`,
	})

	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, response.StatusCode)
	assertErrorMessage(t, response, "Invalid JSON payload")
}

func TestHandle_PartialPayload(t *testing.T) {
	handler := &editWalletLambda{service: &mockWalletService{
		editWalletFunc: func(_ context.Context, _ string, _ string, changedFields map[string]any) (wallet.Wallet, error) {
			assert.Equal(t, map[string]any{"Name": "Travel"}, changedFields)
			return wallet.Wallet{ID: "wallet-1", Name: "Travel"}, nil
		},
	}}
	response, err := handler.Handle(context.Background(), events.APIGatewayV2HTTPRequest{
		PathParameters: map[string]string{"wallet_id": "wallet-1"},
		Body:           `{"name":"Travel"}`,
	})

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, response.StatusCode)
}

func TestHandle_ServiceError(t *testing.T) {
	handler := &editWalletLambda{service: &mockWalletService{
		editWalletFunc: func(context.Context, string, string, map[string]any) (wallet.Wallet, error) {
			return wallet.Wallet{}, assert.AnError
		},
	}}
	response, err := handler.Handle(context.Background(), events.APIGatewayV2HTTPRequest{
		PathParameters: map[string]string{"wallet_id": "wallet-1"},
		Body:           `{"name":"Travel"}`,
	})

	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, response.StatusCode)
	assertErrorMessage(t, response, assert.AnError.Error())
}

func TestHandle_WalletNotFound(t *testing.T) {
	handler := &editWalletLambda{service: &mockWalletService{
		editWalletFunc: func(context.Context, string, string, map[string]any) (wallet.Wallet, error) {
			return wallet.Wallet{}, wallet.ErrWalletNotFound
		},
	}}
	response, err := handler.Handle(context.Background(), events.APIGatewayV2HTTPRequest{
		PathParameters: map[string]string{"wallet_id": "wallet-404"},
		Body:           `{"name":"Travel"}`,
	})

	require.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, response.StatusCode)
	assertErrorMessage(t, response, "Wallet not found")
}

func assertErrorMessage(t *testing.T, response events.APIGatewayV2HTTPResponse, expected string) {
	t.Helper()
	var result errorResponse
	require.NoError(t, json.Unmarshal([]byte(response.Body), &result))
	assert.Equal(t, expected, result.Message)
}

func (m *mockWalletService) PrepareCreateWallet(ownerID string, name string, currency string, balance float64, walletType string) (wallet.Wallet, clients.WriteOp, error) {
	return wallet.Wallet{}, clients.WriteOp{}, nil
}
