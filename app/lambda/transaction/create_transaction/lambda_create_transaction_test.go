package main

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/tonytkl/satang/clients"

	"github.com/aws/aws-lambda-go/events"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tonytkl/satang/category"
	"github.com/tonytkl/satang/transaction"
	"github.com/tonytkl/satang/wallet"
)

// MockTransactionService is a mock implementation of TransactionService
type MockTransactionService struct {
	CreateTransactionFunc        func(ctx context.Context, walletID, walletName, categoryID, categoryName, description, currency, imageURL, transactionType string, amount float64, date time.Time, ownerID string) error
	PrepareCreateTransactionFunc func(walletID, walletName, categoryID, categoryName, description, currency, imageURL, transactionType string, amount float64, date time.Time, ownerID string) (transaction.Transaction, clients.WriteOp, error)
	GetTransactionFunc           func(ctx context.Context, transactionID string) (*transaction.Transaction, error)
}

func (m *MockTransactionService) CreateTransaction(ctx context.Context, walletID, walletName, categoryID, categoryName, description, currency, imageURL, transactionType string, amount float64, date time.Time, ownerID string) error {
	if m.CreateTransactionFunc != nil {
		return m.CreateTransactionFunc(ctx, walletID, walletName, categoryID, categoryName, description, currency, imageURL, transactionType, amount, date, ownerID)
	}
	return nil
}

func (m *MockTransactionService) GetTransaction(ctx context.Context, transactionID string, ownerID string) (*transaction.Transaction, error) {
	if m.GetTransactionFunc != nil {
		return m.GetTransactionFunc(ctx, transactionID)
	}
	return nil, nil
}

func (m *MockTransactionService) GetTransactionsBetweenPeriod(ctx context.Context, ownerID string, fromDate time.Time, toDate time.Time, limit int32, cursor string) ([]transaction.Transaction, string, error) {
	return nil, "", nil
}

func (m *MockTransactionService) ListTransactions(ctx context.Context, ownerID string, fromDate time.Time, toDate time.Time, limit int32, cursor string) ([]transaction.Transaction, string, error) {
	return nil, "", nil
}

func (m *MockTransactionService) ListTransactionsOfCategory(ctx context.Context, ownerID string, fromDate time.Time, toDate time.Time, limit int32, cursor string, categoryID string) ([]transaction.Transaction, string, error) {
	return nil, "", nil
}

func (m *MockTransactionService) ListTransactionsOfWallet(ctx context.Context, ownerID string, fromDate time.Time, toDate time.Time, limit int32, cursor string, wallet string) ([]transaction.Transaction, string, error) {
	return nil, "", nil
}

func (m *MockTransactionService) EditTransaction(ctx context.Context, ownerID string, transactionID string, changedFields map[string]any) error {
	return nil
}

func (m *MockTransactionService) DeleteTransaction(ctx context.Context, ownerID string, transactionID string) error {
	return nil
}

func TestHandle_ValidPayload(t *testing.T) {
	payload := createTransactionRequest{
		WalletID:    "wallet-123",
		CategoryID:  "cat-456",
		Description: "Lunch",
		Currency:    "USD",
		ImageURL:    "https://example.com/image.jpg",
		Type:        "expense",
		Amount:      25.50,
		Date:        "2025-05-17",
	}

	body, err := json.Marshal(payload)
	require.NoError(t, err)

	mockService := &MockTransactionService{
		CreateTransactionFunc: func(ctx context.Context, walletID, walletName, categoryID, categoryName, description, currency, imageURL, transactionType string, amount float64, date time.Time, ownerID string) error {
			return nil
		},
	}

	handler := newMockTestHandler(mockService)
	request := events.APIGatewayV2HTTPRequest{
		Body: string(body),
	}

	response, err := handler.Handle(context.Background(), request)

	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, response.StatusCode)
	var created transaction.TransactionRead
	require.NoError(t, json.Unmarshal([]byte(response.Body), &created))
	assert.NotEmpty(t, created.ID)
	assert.Equal(t, payload.WalletID, created.WalletID)
	assert.Equal(t, payload.CategoryID, created.CategoryID)
	assert.Equal(t, transaction.TransactionTypeExpense, created.Type)
	assert.Equal(t, payload.Amount, created.Amount)
}

func TestHandle_ReturnsCreatedTransaction(t *testing.T) {
	description := "Lunch"
	imageURL := "https://example.com/image.jpg"
	date := time.Date(2025, time.May, 17, 0, 0, 0, 0, time.UTC)
	createdAt := time.Date(2025, time.May, 18, 12, 0, 0, 0, time.UTC)
	updatedAt := createdAt.Add(time.Minute)
	created := transaction.Transaction{
		ID:           "tx-123",
		WalletID:     "wallet-123",
		WalletName:   "Main",
		CategoryID:   "cat-456",
		CategoryName: "Food",
		Description:  &description,
		Currency:     "USD",
		ImageURL:     &imageURL,
		Type:         transaction.TransactionTypeExpense,
		Amount:       25.50,
		Date:         date,
		OwnerID:      "1",
		CreatedAt:    createdAt,
		UpdatedAt:    updatedAt,
	}
	service := &MockTransactionService{
		PrepareCreateTransactionFunc: func(walletID, walletName, categoryID, categoryName, description, currency, imageURL, txType string, amount float64, date time.Time, ownerID string) (transaction.Transaction, clients.WriteOp, error) {
			return created, clients.WriteOp{Kind: clients.WritePut, Table: "items", Item: &created}, nil
		},
	}
	writer := &atomicWriter{}
	handler := newMockTestHandler(service)
	handler.writer = writer

	response, err := handler.Handle(context.Background(), events.APIGatewayV2HTTPRequest{
		Body: `{"walletId":"wallet-123","categoryId":"cat-456","description":"Lunch","currency":"USD","imageURL":"https://example.com/image.jpg","type":"expense","amount":25.5,"date":"2025-05-17"}`,
	})

	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, response.StatusCode)
	var got transaction.TransactionRead
	require.NoError(t, json.Unmarshal([]byte(response.Body), &got))
	assert.Equal(t, transaction.TransactionRead{
		ID:           "tx-123",
		WalletID:     "wallet-123",
		WalletName:   "Main",
		CategoryID:   "cat-456",
		CategoryName: "Food",
		Description:  &description,
		Currency:     "USD",
		ImageURL:     &imageURL,
		Type:         transaction.TransactionTypeExpense,
		Amount:       25.50,
		Date:         date,
		OwnerID:      "1",
		CreatedAt:    createdAt,
		UpdatedAt:    updatedAt,
	}, got)
	require.Len(t, writer.ops, 3)
	assert.Equal(t, &created, writer.ops[2].Item)
}

func TestHandle_InvalidCreatedTransactionDoesNotWrite(t *testing.T) {
	writer := &atomicWriter{}
	service := &MockTransactionService{
		PrepareCreateTransactionFunc: func(string, string, string, string, string, string, string, string, float64, time.Time, string) (transaction.Transaction, clients.WriteOp, error) {
			return transaction.Transaction{}, clients.WriteOp{Kind: clients.WritePut}, nil
		},
	}
	handler := newMockTestHandler(service)
	handler.writer = writer

	response, err := handler.Handle(context.Background(), events.APIGatewayV2HTTPRequest{
		Body: `{"walletId":"w","categoryId":"c","currency":"THB","type":"expense","amount":1,"date":"2025-05-17"}`,
	})

	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, response.StatusCode)
	assert.Contains(t, response.Body, "transaction response schema validation failed")
	assert.Empty(t, writer.ops)
}

func TestHandle_InvalidJSONPayload(t *testing.T) {
	mockService := &MockTransactionService{}
	handler := newMockTestHandler(mockService)
	request := events.APIGatewayV2HTTPRequest{
		Body: "invalid json",
	}

	response, err := handler.Handle(context.Background(), request)

	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, response.StatusCode)

	var errResp errorResponse
	err = json.Unmarshal([]byte(response.Body), &errResp)
	require.NoError(t, err)
	assert.Equal(t, "Invalid JSON payload", errResp.Message)
}

func TestHandle_MissingWalletID(t *testing.T) {
	payload := createTransactionRequest{
		WalletID:   "", // missing
		CategoryID: "cat-456",
		Currency:   "USD",
		Type:       "expense",
		Amount:     25.50,
		Date:       "2025-05-17",
	}

	body, err := json.Marshal(payload)
	require.NoError(t, err)

	mockService := &MockTransactionService{}
	handler := newMockTestHandler(mockService)
	request := events.APIGatewayV2HTTPRequest{
		Body: string(body),
	}

	response, err := handler.Handle(context.Background(), request)

	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, response.StatusCode)

	var errResp errorResponse
	err = json.Unmarshal([]byte(response.Body), &errResp)
	require.NoError(t, err)
	assert.Equal(t, "walletId is required", errResp.Message)
}

func TestHandle_MissingCategoryID(t *testing.T) {
	payload := createTransactionRequest{
		WalletID:   "wallet-123",
		CategoryID: "", // missing
		Currency:   "USD",
		Type:       "expense",
		Amount:     25.50,
		Date:       "2025-05-17",
	}

	body, err := json.Marshal(payload)
	require.NoError(t, err)

	mockService := &MockTransactionService{}
	handler := newMockTestHandler(mockService)
	request := events.APIGatewayV2HTTPRequest{
		Body: string(body),
	}

	response, err := handler.Handle(context.Background(), request)

	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, response.StatusCode)

	var errResp errorResponse
	err = json.Unmarshal([]byte(response.Body), &errResp)
	require.NoError(t, err)
	assert.Equal(t, "categoryId is required", errResp.Message)
}

func TestHandle_MissingCurrency(t *testing.T) {
	payload := createTransactionRequest{
		WalletID:   "wallet-123",
		CategoryID: "cat-456",
		Currency:   "", // missing
		Type:       "expense",
		Amount:     25.50,
		Date:       "2025-05-17",
	}

	body, err := json.Marshal(payload)
	require.NoError(t, err)

	mockService := &MockTransactionService{}
	handler := newMockTestHandler(mockService)
	request := events.APIGatewayV2HTTPRequest{
		Body: string(body),
	}

	response, err := handler.Handle(context.Background(), request)

	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, response.StatusCode)

	var errResp errorResponse
	err = json.Unmarshal([]byte(response.Body), &errResp)
	require.NoError(t, err)
	assert.Equal(t, "currency is required", errResp.Message)
}

func TestHandle_MissingType(t *testing.T) {
	payload := createTransactionRequest{
		WalletID:   "wallet-123",
		CategoryID: "cat-456",
		Currency:   "USD",
		Type:       "", // missing
		Amount:     25.50,
		Date:       "2025-05-17",
	}

	body, err := json.Marshal(payload)
	require.NoError(t, err)

	mockService := &MockTransactionService{}
	handler := newMockTestHandler(mockService)
	request := events.APIGatewayV2HTTPRequest{
		Body: string(body),
	}

	response, err := handler.Handle(context.Background(), request)

	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, response.StatusCode)

	var errResp errorResponse
	err = json.Unmarshal([]byte(response.Body), &errResp)
	require.NoError(t, err)
	assert.Equal(t, "type is required", errResp.Message)
}

func TestHandle_MissingAmount(t *testing.T) {
	payload := createTransactionRequest{
		WalletID:   "wallet-123",
		CategoryID: "cat-456",
		Currency:   "USD",
		Type:       "expense",
		Amount:     0, // missing
		Date:       "2025-05-17",
	}

	body, err := json.Marshal(payload)
	require.NoError(t, err)

	mockService := &MockTransactionService{}
	handler := newMockTestHandler(mockService)
	request := events.APIGatewayV2HTTPRequest{
		Body: string(body),
	}

	response, err := handler.Handle(context.Background(), request)

	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, response.StatusCode)

	var errResp errorResponse
	err = json.Unmarshal([]byte(response.Body), &errResp)
	require.NoError(t, err)
	assert.Equal(t, "amount is required", errResp.Message)
}

func TestHandle_MissingDate(t *testing.T) {
	payload := createTransactionRequest{
		WalletID:   "wallet-123",
		CategoryID: "cat-456",
		Currency:   "USD",
		Type:       "expense",
		Amount:     25.50,
		Date:       "", // missing
	}

	body, err := json.Marshal(payload)
	require.NoError(t, err)

	mockService := &MockTransactionService{}
	handler := newMockTestHandler(mockService)
	request := events.APIGatewayV2HTTPRequest{
		Body: string(body),
	}

	response, err := handler.Handle(context.Background(), request)

	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, response.StatusCode)

	var errResp errorResponse
	err = json.Unmarshal([]byte(response.Body), &errResp)
	require.NoError(t, err)
	assert.Equal(t, "date is required", errResp.Message)
}

func TestHandle_InvalidDateFormat(t *testing.T) {
	payload := createTransactionRequest{
		WalletID:   "wallet-123",
		CategoryID: "cat-456",
		Currency:   "USD",
		Type:       "expense",
		Amount:     25.50,
		Date:       "invalid-date", // invalid format
	}

	body, err := json.Marshal(payload)
	require.NoError(t, err)

	mockService := &MockTransactionService{}
	handler := newMockTestHandler(mockService)
	request := events.APIGatewayV2HTTPRequest{
		Body: string(body),
	}

	response, err := handler.Handle(context.Background(), request)

	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, response.StatusCode)

	var errResp errorResponse
	err = json.Unmarshal([]byte(response.Body), &errResp)
	require.NoError(t, err)
	assert.Equal(t, "Date must be RFC3339 or YYYY-MM-DD", errResp.Message)
}

func TestHandle_ServiceError(t *testing.T) {
	payload := createTransactionRequest{
		WalletID:   "wallet-123",
		CategoryID: "cat-456",
		Currency:   "USD",
		Type:       "expense",
		Amount:     25.50,
		Date:       "2025-05-17",
	}

	body, err := json.Marshal(payload)
	require.NoError(t, err)

	mockService := &MockTransactionService{
		CreateTransactionFunc: func(ctx context.Context, walletID, walletName, categoryID, categoryName, description, currency, imageURL, transactionType string, amount float64, date time.Time, ownerID string) error {
			return errors.New("database error")
		},
	}

	handler := newMockTestHandler(mockService)
	request := events.APIGatewayV2HTTPRequest{
		Body: string(body),
	}

	response, err := handler.Handle(context.Background(), request)

	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, response.StatusCode)

	var errResp errorResponse
	err = json.Unmarshal([]byte(response.Body), &errResp)
	require.NoError(t, err)
	assert.Equal(t, "database error", errResp.Message)
}

func TestHandle_WalletNotFoundReturnsNotFound(t *testing.T) {
	writer := &atomicWriter{}
	handler := &createTransactionLambda{
		txService:       atomicTransactionService{},
		walletService:   atomicWalletService{err: wallet.ErrWalletNotFound},
		categoryService: atomicCategoryService{},
		writer:          writer,
		tableName:       "items",
	}

	response, err := handler.Handle(context.Background(), events.APIGatewayV2HTTPRequest{
		Body: `{"walletId":"missing-wallet","categoryId":"category-1","currency":"THB","type":"expense","amount":10,"date":"2025-05-17"}`,
	})

	require.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, response.StatusCode)
	assert.Contains(t, response.Body, "No wallet match given query")
	assert.Empty(t, writer.ops)
}

func TestHandle_CategoryNotFoundReturnsNotFound(t *testing.T) {
	writer := &atomicWriter{}
	handler := &createTransactionLambda{
		txService:       atomicTransactionService{},
		walletService:   atomicWalletService{},
		categoryService: atomicCategoryService{err: category.ErrCategoryNotFound},
		writer:          writer,
		tableName:       "items",
	}

	response, err := handler.Handle(context.Background(), events.APIGatewayV2HTTPRequest{
		Body: `{"walletId":"wallet-1","categoryId":"missing-category","currency":"THB","type":"expense","amount":10,"date":"2025-05-17"}`,
	})

	require.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, response.StatusCode)
	assert.Contains(t, response.Body, "No category match given query")
	assert.Empty(t, writer.ops)
}

func TestHandle_ServiceDeadline(t *testing.T) {
	service := &MockTransactionService{
		CreateTransactionFunc: func(context.Context, string, string, string, string, string, string, string, string, float64, time.Time, string) error {
			return context.DeadlineExceeded
		},
	}
	response, err := newMockTestHandler(service).Handle(context.Background(), events.APIGatewayV2HTTPRequest{
		Body: `{"walletId":"w","walletName":"wallet","categoryId":"c","categoryName":"category","currency":"THB","type":"expense","amount":1,"date":"2025-05-17"}`,
	})
	require.NoError(t, err)
	assert.Equal(t, http.StatusGatewayTimeout, response.StatusCode)
}

func TestHandle_InvalidTransactionType(t *testing.T) {
	response, err := newTestHandler(transaction.NewService(nil)).Handle(context.Background(), events.APIGatewayV2HTTPRequest{
		Body: `{"walletId":"w","walletName":"wallet","categoryId":"c","categoryName":"category","currency":"THB","type":"invalid","amount":1,"date":"2025-05-17"}`,
	})
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, response.StatusCode)
}

func TestHandle_ValidDateInRFC3339Format(t *testing.T) {
	payload := createTransactionRequest{
		WalletID:   "wallet-123",
		CategoryID: "cat-456",
		Currency:   "USD",
		Type:       "expense",
		Amount:     25.50,
		Date:       "2025-05-17T10:30:00Z",
	}

	body, err := json.Marshal(payload)
	require.NoError(t, err)

	mockService := &MockTransactionService{
		CreateTransactionFunc: func(ctx context.Context, walletID, walletName, categoryID, categoryName, description, currency, imageURL, transactionType string, amount float64, date time.Time, ownerID string) error {
			return nil
		},
	}

	handler := newMockTestHandler(mockService)
	request := events.APIGatewayV2HTTPRequest{
		Body: string(body),
	}

	response, err := handler.Handle(context.Background(), request)

	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, response.StatusCode)
}

func TestValidatePayload_AllFieldsValid(t *testing.T) {
	payload := createTransactionRequest{
		WalletID:   "wallet-123",
		CategoryID: "cat-456",
		Currency:   "USD",
		Type:       "expense",
		Amount:     25.50,
		Date:       "2025-05-17",
	}

	err := validatePayload(payload)
	assert.NoError(t, err)
}

func TestValidatePayload_WhitespaceIsInvalid(t *testing.T) {
	tests := []struct {
		name     string
		payload  createTransactionRequest
		expected string
	}{
		{
			name: "whitespace walletId",
			payload: createTransactionRequest{
				WalletID:   "   ",
				CategoryID: "cat-456",
				Currency:   "USD",
				Type:       "expense",
				Amount:     25.50,
				Date:       "2025-05-17",
			},
			expected: "walletId is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validatePayload(tt.payload)
			require.Error(t, err)
			assert.Equal(t, tt.expected, err.Error())
		})
	}
}

func (m *MockTransactionService) PrepareCreateTransaction(walletID string, walletName string, categoryID string, categoryName string, description string, currency string, imageURL string, txType string, amount float64, date time.Time, ownerID string) (transaction.Transaction, clients.WriteOp, error) {
	if m.PrepareCreateTransactionFunc != nil {
		return m.PrepareCreateTransactionFunc(walletID, walletName, categoryID, categoryName, description, currency, imageURL, txType, amount, date, ownerID)
	}
	if m.CreateTransactionFunc != nil {
		if err := m.CreateTransactionFunc(context.Background(), walletID, walletName, categoryID, categoryName, description, currency, imageURL, txType, amount, date, ownerID); err != nil {
			return transaction.Transaction{}, clients.WriteOp{}, err
		}
	}
	created := transaction.NewTransaction(walletID, walletName, categoryID, categoryName, description, currency, imageURL, transaction.TransactionType(strings.ToUpper(txType)), amount, date, ownerID)
	return *created, clients.WriteOp{Kind: clients.WritePut, Table: "items", Item: created}, nil
}

type mockTransactionServiceAdapter struct {
	transaction.Service
	mock *MockTransactionService
}

func (adapter mockTransactionServiceAdapter) PrepareCreateTransaction(walletID, walletName, categoryID, categoryName, description, currency, imageURL, txType string, amount float64, date time.Time, ownerID string) (transaction.Transaction, clients.WriteOp, error) {
	return adapter.mock.PrepareCreateTransaction(walletID, walletName, categoryID, categoryName, description, currency, imageURL, txType, amount, date, ownerID)
}

func newTestHandler(service transaction.Service) *createTransactionLambda {
	return &createTransactionLambda{
		txService:       service,
		walletService:   atomicWalletService{},
		categoryService: atomicCategoryService{},
		writer:          &atomicWriter{},
		tableName:       "items",
	}
}

func newMockTestHandler(service *MockTransactionService) *createTransactionLambda {
	return newTestHandler(mockTransactionServiceAdapter{mock: service})
}
