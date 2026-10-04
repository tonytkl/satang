package main

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tonytkl/satang/transaction"
	"github.com/tonytkl/satang/wallet"
)

// MockWalletService is a mock implementation of wallet.Service
// used to validate the lambda request flow.
type MockWalletService struct {
	CreateWalletFunc func(ctx context.Context, ownerID string, name string, currency string, balance float64, walletType string) (wallet.Wallet, error)
}

func (m *MockWalletService) CreateWallet(ctx context.Context, ownerID string, name string, currency string, balance float64, walletType string) (wallet.Wallet, error) {
	if m.CreateWalletFunc != nil {
		return m.CreateWalletFunc(ctx, ownerID, name, currency, balance, walletType)
	}
	return wallet.Wallet{}, nil
}

func (m *MockWalletService) ListWallets(ctx context.Context, ownerID string, nextToken string, limit int32) ([]wallet.Wallet, string, error) {
	return nil, "", nil
}

func (m *MockWalletService) GetWallet(ctx context.Context, ownerID string, walletID string) (wallet.Wallet, error) {
	return wallet.Wallet{}, nil
}

func (m *MockWalletService) EditWallet(ctx context.Context, ownerID string, walletID string, changedFields map[string]any) (wallet.Wallet, error) {
	return wallet.Wallet{}, nil
}

func (m *MockWalletService) SetActiveWallet(ctx context.Context, ownerID string, walletID string, isActive bool) error {
	return nil
}

type MockTransactionService struct {
	transaction.Service
	CreateTransactionFunc func(ctx context.Context, walletID string, walletName string, categoryID string, categoryName string, description string, currency string, imageURL string, txType string, amount float64, date time.Time, ownerID string) error
}

func (m *MockTransactionService) CreateTransaction(ctx context.Context, walletID string, walletName string, categoryID string, categoryName string, description string, currency string, imageURL string, txType string, amount float64, date time.Time, ownerID string) error {
	if m.CreateTransactionFunc != nil {
		return m.CreateTransactionFunc(ctx, walletID, walletName, categoryID, categoryName, description, currency, imageURL, txType, amount, date, ownerID)
	}
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
		CreateWalletFunc: func(ctx context.Context, ownerID string, name string, currency string, balance float64, walletType string) (wallet.Wallet, error) {
			assert.Equal(t, "1", ownerID)
			assert.Equal(t, "Main Wallet", name)
			assert.Equal(t, "USD", currency)
			assert.Equal(t, 250.0, balance)
			assert.Equal(t, "debit", walletType)
			return wallet.Wallet{
				ID:       "wallet-1",
				OwnerID:  ownerID,
				Name:     name,
				Currency: currency,
				Balance:  balance,
				Type:     wallet.WalletTypeDebit,
			}, nil
		},
	}

	handler := &createWalletLambda{walletService: mockService, transactionService: &MockTransactionService{}}
	response, err := handler.Handle(context.Background(), events.APIGatewayV2HTTPRequest{Body: string(body)})

	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, response.StatusCode)
}

func TestHandle_InvalidJSONPayload(t *testing.T) {
	handler := &createWalletLambda{walletService: &MockWalletService{}, transactionService: &MockTransactionService{}}
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

	handler := &createWalletLambda{walletService: &MockWalletService{}, transactionService: &MockTransactionService{}}
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
		CreateWalletFunc: func(ctx context.Context, ownerID string, name string, currency string, balance float64, walletType string) (wallet.Wallet, error) {
			return wallet.Wallet{}, assert.AnError
		},
	}

	handler := &createWalletLambda{walletService: mockService, transactionService: &MockTransactionService{}}
	response, err := handler.Handle(context.Background(), events.APIGatewayV2HTTPRequest{Body: string(body)})

	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, response.StatusCode)

	var errResp errorResponse
	err = json.Unmarshal([]byte(response.Body), &errResp)
	require.NoError(t, err)
	assert.Equal(t, assert.AnError.Error(), errResp.Message)
}

var _ wallet.Service = (*MockWalletService)(nil)

func TestHandle_NonZeroBalanceCreatesInitialTransaction(t *testing.T) {
	body, err := json.Marshal(createWalletRequest{Name: "Main Wallet", Balance: 250.0, WalletType: "debit"})
	require.NoError(t, err)

	walletService := &MockWalletService{
		CreateWalletFunc: func(ctx context.Context, ownerID string, name string, currency string, balance float64, walletType string) (wallet.Wallet, error) {
			return wallet.Wallet{ID: "wallet-1", OwnerID: ownerID, Name: name, Currency: "THB", Balance: balance, Type: wallet.WalletTypeDebit}, nil
		},
	}

	calls := 0
	transactionService := &MockTransactionService{
		CreateTransactionFunc: func(ctx context.Context, walletID string, walletName string, categoryID string, categoryName string, description string, currency string, imageURL string, txType string, amount float64, date time.Time, ownerID string) error {
			calls++
			assert.Equal(t, "wallet-1", walletID)
			assert.Equal(t, "Main Wallet", walletName)
			assert.Equal(t, "THB", currency)
			assert.Equal(t, string(transaction.TransactionTypeIncome), txType)
			assert.Equal(t, 250.0, amount)
			assert.Equal(t, "1", ownerID)
			return nil
		},
	}

	handler := &createWalletLambda{walletService: walletService, transactionService: transactionService}
	response, err := handler.Handle(context.Background(), events.APIGatewayV2HTTPRequest{Body: string(body)})

	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, response.StatusCode)
	assert.Equal(t, 1, calls)
}

func TestHandle_ZeroBalanceSkipsInitialTransaction(t *testing.T) {
	body, err := json.Marshal(createWalletRequest{Name: "Main Wallet", WalletType: "debit"})
	require.NoError(t, err)

	calls := 0
	transactionService := &MockTransactionService{
		CreateTransactionFunc: func(ctx context.Context, walletID string, walletName string, categoryID string, categoryName string, description string, currency string, imageURL string, txType string, amount float64, date time.Time, ownerID string) error {
			calls++
			return nil
		},
	}

	handler := &createWalletLambda{walletService: &MockWalletService{}, transactionService: transactionService}
	response, err := handler.Handle(context.Background(), events.APIGatewayV2HTTPRequest{Body: string(body)})

	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, response.StatusCode)
	assert.Equal(t, 0, calls)
}

func TestValidatePayload_EmptyWalletType(t *testing.T) {
	err := validatePayload(createWalletRequest{WalletType: "   "})
	require.Error(t, err)
	assert.Equal(t, "type is required", err.Error())
}

func TestValidatePayload_ValidWalletType(t *testing.T) {
	err := validatePayload(createWalletRequest{WalletType: "credit"})
	require.NoError(t, err)
}
