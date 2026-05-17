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
	"github.com/tonytkl/satang/model"
)

// MockTransactionService is a mock implementation of TransactionService
type MockTransactionService struct {
	CreateTransactionFunc func(ctx context.Context, walletID, walletName, categoryID, categoryName, description, currency, imageURL, transactionType string, amount float64, date time.Time, ownerID string) error
	GetTransactionFunc    func(ctx context.Context, transactionID string) (*model.Transaction, error)
}

func (m *MockTransactionService) CreateTransaction(ctx context.Context, walletID, walletName, categoryID, categoryName, description, currency, imageURL, transactionType string, amount float64, date time.Time, ownerID string) error {
	if m.CreateTransactionFunc != nil {
		return m.CreateTransactionFunc(ctx, walletID, walletName, categoryID, categoryName, description, currency, imageURL, transactionType, amount, date, ownerID)
	}
	return nil
}

func (m *MockTransactionService) GetTransaction(ctx context.Context, transactionID string) (*model.Transaction, error) {
	if m.GetTransactionFunc != nil {
		return m.GetTransactionFunc(ctx, transactionID)
	}
	return nil, nil
}

func TestHandle_ValidPayload(t *testing.T) {
	payload := createTransactionRequest{
		WalletID:     "wallet-123",
		WalletName:   "Main Wallet",
		CategoryID:   "cat-456",
		CategoryName: "Food",
		Description:  "Lunch",
		Currency:     "USD",
		ImageURL:     "https://example.com/image.jpg",
		Type:         "expense",
		Amount:       25.50,
		Date:         "2025-05-17",
	}

	body, err := json.Marshal(payload)
	require.NoError(t, err)

	mockService := &MockTransactionService{
		CreateTransactionFunc: func(ctx context.Context, walletID, walletName, categoryID, categoryName, description, currency, imageURL, transactionType string, amount float64, date time.Time, ownerID string) error {
			return nil
		},
	}

	handler := &createTransactionLambda{service: mockService}
	request := events.APIGatewayV2HTTPRequest{
		Body: string(body),
	}

	response, err := handler.Handle(context.Background(), request)

	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, response.StatusCode)
}

func TestHandle_InvalidJSONPayload(t *testing.T) {
	mockService := &MockTransactionService{}
	handler := &createTransactionLambda{service: mockService}
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
		WalletID:     "", // missing
		WalletName:   "Main Wallet",
		CategoryID:   "cat-456",
		CategoryName: "Food",
		Currency:     "USD",
		Type:         "expense",
		Amount:       25.50,
		Date:         "2025-05-17",
	}

	body, err := json.Marshal(payload)
	require.NoError(t, err)

	mockService := &MockTransactionService{}
	handler := &createTransactionLambda{service: mockService}
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

func TestHandle_MissingWalletName(t *testing.T) {
	payload := createTransactionRequest{
		WalletID:     "wallet-123",
		WalletName:   "", // missing
		CategoryID:   "cat-456",
		CategoryName: "Food",
		Currency:     "USD",
		Type:         "expense",
		Amount:       25.50,
		Date:         "2025-05-17",
	}

	body, err := json.Marshal(payload)
	require.NoError(t, err)

	mockService := &MockTransactionService{}
	handler := &createTransactionLambda{service: mockService}
	request := events.APIGatewayV2HTTPRequest{
		Body: string(body),
	}

	response, err := handler.Handle(context.Background(), request)

	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, response.StatusCode)

	var errResp errorResponse
	err = json.Unmarshal([]byte(response.Body), &errResp)
	require.NoError(t, err)
	assert.Equal(t, "walletName is required", errResp.Message)
}

func TestHandle_MissingCategoryID(t *testing.T) {
	payload := createTransactionRequest{
		WalletID:     "wallet-123",
		WalletName:   "Main Wallet",
		CategoryID:   "", // missing
		CategoryName: "Food",
		Currency:     "USD",
		Type:         "expense",
		Amount:       25.50,
		Date:         "2025-05-17",
	}

	body, err := json.Marshal(payload)
	require.NoError(t, err)

	mockService := &MockTransactionService{}
	handler := &createTransactionLambda{service: mockService}
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

func TestHandle_MissingCategoryName(t *testing.T) {
	payload := createTransactionRequest{
		WalletID:     "wallet-123",
		WalletName:   "Main Wallet",
		CategoryID:   "cat-456",
		CategoryName: "", // missing
		Currency:     "USD",
		Type:         "expense",
		Amount:       25.50,
		Date:         "2025-05-17",
	}

	body, err := json.Marshal(payload)
	require.NoError(t, err)

	mockService := &MockTransactionService{}
	handler := &createTransactionLambda{service: mockService}
	request := events.APIGatewayV2HTTPRequest{
		Body: string(body),
	}

	response, err := handler.Handle(context.Background(), request)

	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, response.StatusCode)

	var errResp errorResponse
	err = json.Unmarshal([]byte(response.Body), &errResp)
	require.NoError(t, err)
	assert.Equal(t, "categoryName is required", errResp.Message)
}

func TestHandle_MissingCurrency(t *testing.T) {
	payload := createTransactionRequest{
		WalletID:     "wallet-123",
		WalletName:   "Main Wallet",
		CategoryID:   "cat-456",
		CategoryName: "Food",
		Currency:     "", // missing
		Type:         "expense",
		Amount:       25.50,
		Date:         "2025-05-17",
	}

	body, err := json.Marshal(payload)
	require.NoError(t, err)

	mockService := &MockTransactionService{}
	handler := &createTransactionLambda{service: mockService}
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
		WalletID:     "wallet-123",
		WalletName:   "Main Wallet",
		CategoryID:   "cat-456",
		CategoryName: "Food",
		Currency:     "USD",
		Type:         "", // missing
		Amount:       25.50,
		Date:         "2025-05-17",
	}

	body, err := json.Marshal(payload)
	require.NoError(t, err)

	mockService := &MockTransactionService{}
	handler := &createTransactionLambda{service: mockService}
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
		WalletID:     "wallet-123",
		WalletName:   "Main Wallet",
		CategoryID:   "cat-456",
		CategoryName: "Food",
		Currency:     "USD",
		Type:         "expense",
		Amount:       0, // missing
		Date:         "2025-05-17",
	}

	body, err := json.Marshal(payload)
	require.NoError(t, err)

	mockService := &MockTransactionService{}
	handler := &createTransactionLambda{service: mockService}
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
		WalletID:     "wallet-123",
		WalletName:   "Main Wallet",
		CategoryID:   "cat-456",
		CategoryName: "Food",
		Currency:     "USD",
		Type:         "expense",
		Amount:       25.50,
		Date:         "", // missing
	}

	body, err := json.Marshal(payload)
	require.NoError(t, err)

	mockService := &MockTransactionService{}
	handler := &createTransactionLambda{service: mockService}
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
		WalletID:     "wallet-123",
		WalletName:   "Main Wallet",
		CategoryID:   "cat-456",
		CategoryName: "Food",
		Currency:     "USD",
		Type:         "expense",
		Amount:       25.50,
		Date:         "invalid-date", // invalid format
	}

	body, err := json.Marshal(payload)
	require.NoError(t, err)

	mockService := &MockTransactionService{}
	handler := &createTransactionLambda{service: mockService}
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
		WalletID:     "wallet-123",
		WalletName:   "Main Wallet",
		CategoryID:   "cat-456",
		CategoryName: "Food",
		Currency:     "USD",
		Type:         "expense",
		Amount:       25.50,
		Date:         "2025-05-17",
	}

	body, err := json.Marshal(payload)
	require.NoError(t, err)

	mockService := &MockTransactionService{
		CreateTransactionFunc: func(ctx context.Context, walletID, walletName, categoryID, categoryName, description, currency, imageURL, transactionType string, amount float64, date time.Time, ownerID string) error {
			return errors.New("database error")
		},
	}

	handler := &createTransactionLambda{service: mockService}
	request := events.APIGatewayV2HTTPRequest{
		Body: string(body),
	}

	response, err := handler.Handle(context.Background(), request)

	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, response.StatusCode)

	var errResp errorResponse
	err = json.Unmarshal([]byte(response.Body), &errResp)
	require.NoError(t, err)
	assert.Equal(t, "database error", errResp.Message)
}

func TestHandle_ValidDateInRFC3339Format(t *testing.T) {
	payload := createTransactionRequest{
		WalletID:     "wallet-123",
		WalletName:   "Main Wallet",
		CategoryID:   "cat-456",
		CategoryName: "Food",
		Currency:     "USD",
		Type:         "expense",
		Amount:       25.50,
		Date:         "2025-05-17T10:30:00Z",
	}

	body, err := json.Marshal(payload)
	require.NoError(t, err)

	mockService := &MockTransactionService{
		CreateTransactionFunc: func(ctx context.Context, walletID, walletName, categoryID, categoryName, description, currency, imageURL, transactionType string, amount float64, date time.Time, ownerID string) error {
			return nil
		},
	}

	handler := &createTransactionLambda{service: mockService}
	request := events.APIGatewayV2HTTPRequest{
		Body: string(body),
	}

	response, err := handler.Handle(context.Background(), request)

	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, response.StatusCode)
}

func TestValidatePayload_AllFieldsValid(t *testing.T) {
	payload := createTransactionRequest{
		WalletID:     "wallet-123",
		WalletName:   "Main Wallet",
		CategoryID:   "cat-456",
		CategoryName: "Food",
		Currency:     "USD",
		Type:         "expense",
		Amount:       25.50,
		Date:         "2025-05-17",
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
				WalletID:     "   ",
				WalletName:   "Main Wallet",
				CategoryID:   "cat-456",
				CategoryName: "Food",
				Currency:     "USD",
				Type:         "expense",
				Amount:       25.50,
				Date:         "2025-05-17",
			},
			expected: "walletId is required",
		},
		{
			name: "whitespace walletName",
			payload: createTransactionRequest{
				WalletID:     "wallet-123",
				WalletName:   "   ",
				CategoryID:   "cat-456",
				CategoryName: "Food",
				Currency:     "USD",
				Type:         "expense",
				Amount:       25.50,
				Date:         "2025-05-17",
			},
			expected: "walletName is required",
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
