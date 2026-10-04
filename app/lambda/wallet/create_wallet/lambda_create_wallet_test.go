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
	"github.com/tonytkl/satang/clients"
	"github.com/tonytkl/satang/transaction"
	"github.com/tonytkl/satang/wallet"
)

// MockWalletService is a mock implementation of wallet.Service
// used to validate the lambda request flow.
type MockWalletService struct {
	PrepareCreateWalletFunc func(ownerID string, name string, currency string, balance float64, walletType string) (wallet.Wallet, clients.WriteOp, error)
}

func (m *MockWalletService) CreateWallet(ctx context.Context, ownerID string, name string, currency string, balance float64, walletType string) (wallet.Wallet, error) {
	return wallet.Wallet{}, nil
}

func (m *MockWalletService) PrepareCreateWallet(ownerID string, name string, currency string, balance float64, walletType string) (wallet.Wallet, clients.WriteOp, error) {
	if m.PrepareCreateWalletFunc != nil {
		return m.PrepareCreateWalletFunc(ownerID, name, currency, balance, walletType)
	}
	return wallet.Wallet{}, clients.WriteOp{Table: "wallet"}, nil
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
	PrepareCreateTransactionFunc func(walletID string, walletName string, categoryID string, categoryName string, description string, currency string, imageURL string, txType string, amount float64, date time.Time, ownerID string) (clients.WriteOp, error)
}

func (m *MockTransactionService) PrepareCreateTransaction(walletID string, walletName string, categoryID string, categoryName string, description string, currency string, imageURL string, txType string, amount float64, date time.Time, ownerID string) (clients.WriteOp, error) {
	if m.PrepareCreateTransactionFunc != nil {
		return m.PrepareCreateTransactionFunc(walletID, walletName, categoryID, categoryName, description, currency, imageURL, txType, amount, date, ownerID)
	}
	return clients.WriteOp{Table: "transaction"}, nil
}

type MockWriter struct {
	ops []clients.WriteOp
	err error
}

func (m *MockWriter) TransactWrite(ctx context.Context, ops ...clients.WriteOp) error {
	m.ops = ops
	return m.err
}

func newHandler(walletService wallet.Service, transactionService transaction.Service, writer clients.TransactionalWriter) *createWalletLambda {
	return &createWalletLambda{walletService: walletService, transactionService: transactionService, writer: writer}
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
		PrepareCreateWalletFunc: func(ownerID string, name string, currency string, balance float64, walletType string) (wallet.Wallet, clients.WriteOp, error) {
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
			}, clients.WriteOp{Table: "wallet"}, nil
		},
	}

	handler := newHandler(mockService, &MockTransactionService{}, &MockWriter{})
	response, err := handler.Handle(context.Background(), events.APIGatewayV2HTTPRequest{Body: string(body)})

	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, response.StatusCode)
}

func TestHandle_InvalidJSONPayload(t *testing.T) {
	handler := newHandler(&MockWalletService{}, &MockTransactionService{}, &MockWriter{})
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

	handler := newHandler(&MockWalletService{}, &MockTransactionService{}, &MockWriter{})
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
		PrepareCreateWalletFunc: func(ownerID string, name string, currency string, balance float64, walletType string) (wallet.Wallet, clients.WriteOp, error) {
			return wallet.Wallet{}, clients.WriteOp{}, assert.AnError
		},
	}

	writer := &MockWriter{}
	handler := newHandler(mockService, &MockTransactionService{}, writer)
	response, err := handler.Handle(context.Background(), events.APIGatewayV2HTTPRequest{Body: string(body)})

	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, response.StatusCode)

	var errResp errorResponse
	err = json.Unmarshal([]byte(response.Body), &errResp)
	require.NoError(t, err)
	assert.Equal(t, assert.AnError.Error(), errResp.Message)
	assert.Nil(t, writer.ops)
}

var _ wallet.Service = (*MockWalletService)(nil)

func TestHandle_NonZeroBalanceWritesWalletAndTransactionAtomically(t *testing.T) {
	body, err := json.Marshal(createWalletRequest{Name: "Main Wallet", Balance: 250.0, WalletType: "debit"})
	require.NoError(t, err)

	walletService := &MockWalletService{
		PrepareCreateWalletFunc: func(ownerID string, name string, currency string, balance float64, walletType string) (wallet.Wallet, clients.WriteOp, error) {
			return wallet.Wallet{ID: "wallet-1", OwnerID: ownerID, Name: name, Currency: "THB", Balance: balance, Type: wallet.WalletTypeDebit}, clients.WriteOp{Table: "wallet"}, nil
		},
	}

	calls := 0
	transactionService := &MockTransactionService{
		PrepareCreateTransactionFunc: func(walletID string, walletName string, categoryID string, categoryName string, description string, currency string, imageURL string, txType string, amount float64, date time.Time, ownerID string) (clients.WriteOp, error) {
			calls++
			assert.Equal(t, "wallet-1", walletID)
			assert.Equal(t, "Main Wallet", walletName)
			assert.Equal(t, "THB", currency)
			assert.Equal(t, string(transaction.TransactionTypeIncome), txType)
			assert.Equal(t, 250.0, amount)
			assert.Equal(t, "1", ownerID)
			return clients.WriteOp{Table: "transaction"}, nil
		},
	}

	writer := &MockWriter{}
	response, err := newHandler(walletService, transactionService, writer).Handle(context.Background(), events.APIGatewayV2HTTPRequest{Body: string(body)})

	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, response.StatusCode)
	assert.Equal(t, 1, calls)
	require.Len(t, writer.ops, 2)
	assert.Equal(t, "wallet", writer.ops[0].Table)
	assert.Equal(t, "transaction", writer.ops[1].Table)
}

func TestHandle_ZeroBalanceWritesOnlyWallet(t *testing.T) {
	body, err := json.Marshal(createWalletRequest{Name: "Main Wallet", WalletType: "debit"})
	require.NoError(t, err)

	calls := 0
	transactionService := &MockTransactionService{
		PrepareCreateTransactionFunc: func(walletID string, walletName string, categoryID string, categoryName string, description string, currency string, imageURL string, txType string, amount float64, date time.Time, ownerID string) (clients.WriteOp, error) {
			calls++
			return clients.WriteOp{}, nil
		},
	}

	writer := &MockWriter{}
	response, err := newHandler(&MockWalletService{}, transactionService, writer).Handle(context.Background(), events.APIGatewayV2HTTPRequest{Body: string(body)})

	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, response.StatusCode)
	assert.Equal(t, 0, calls)
	assert.Len(t, writer.ops, 1)
}

func TestHandle_TransactionPrepareErrorWritesNothing(t *testing.T) {
	body, err := json.Marshal(createWalletRequest{Name: "Main Wallet", Balance: 10, WalletType: "debit"})
	require.NoError(t, err)

	transactionService := &MockTransactionService{
		PrepareCreateTransactionFunc: func(walletID string, walletName string, categoryID string, categoryName string, description string, currency string, imageURL string, txType string, amount float64, date time.Time, ownerID string) (clients.WriteOp, error) {
			return clients.WriteOp{}, assert.AnError
		},
	}

	writer := &MockWriter{}
	response, err := newHandler(&MockWalletService{}, transactionService, writer).Handle(context.Background(), events.APIGatewayV2HTTPRequest{Body: string(body)})

	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, response.StatusCode)
	assert.Nil(t, writer.ops)
}

func TestHandle_CommitErrorReturnsInternalServerError(t *testing.T) {
	body, err := json.Marshal(createWalletRequest{Name: "Main Wallet", WalletType: "debit"})
	require.NoError(t, err)

	writer := &MockWriter{err: assert.AnError}
	response, err := newHandler(&MockWalletService{}, &MockTransactionService{}, writer).Handle(context.Background(), events.APIGatewayV2HTTPRequest{Body: string(body)})

	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, response.StatusCode)
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
