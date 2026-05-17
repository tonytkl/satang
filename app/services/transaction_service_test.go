package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tonytkl/satang/model"
	"github.com/tonytkl/satang/repositories"
)

// mockTransactionRepository implements repositories.TransactionRepository for testing
type mockTransactionRepository struct {
	createFn   func(ctx context.Context, transaction *model.Transaction) error
	getByKeyFn func(ctx context.Context, id string) (*model.Transaction, error)
}

var _ repositories.TransactionRepository = (*mockTransactionRepository)(nil)

func (m *mockTransactionRepository) Create(ctx context.Context, transaction *model.Transaction) error {
	if m.createFn != nil {
		return m.createFn(ctx, transaction)
	}
	return nil
}

func (m *mockTransactionRepository) GetByKey(ctx context.Context, id string) (*model.Transaction, error) {
	if m.getByKeyFn != nil {
		return m.getByKeyFn(ctx, id)
	}
	return nil, nil
}

func (m *mockTransactionRepository) ListByGSI(ctx context.Context, indexName string, indexPartitionKeyPrefix string, targetID string, ownerID string, fromDate *time.Time, toDate *time.Time) ([]model.Transaction, error) {
	return nil, nil
}

func (m *mockTransactionRepository) ListWithinDateRange(ctx context.Context, ownerID string, fromDate time.Time, toDate time.Time) ([]model.Transaction, error) {
	return nil, nil
}

func (m *mockTransactionRepository) Update(ctx context.Context, ownerID string, transactionDate string, transactionID string, transaction *model.Transaction) error {
	return nil
}

func (m *mockTransactionRepository) Delete(ctx context.Context, ownerID string, transactionDate string, transactionID string) error {
	return nil
}

func TestCreateTransactionSuccess(t *testing.T) {
	mock := &mockTransactionRepository{
		createFn: func(ctx context.Context, transaction *model.Transaction) error {
			assert.Equal(t, 100.0, transaction.Amount)
			assert.Equal(t, "USD", transaction.Currency)
			assert.Equal(t, "wallet-1", transaction.WalletID)
			assert.Equal(t, "category-1", transaction.CategoryID)
			assert.Equal(t, model.TransactionTypeExpense, transaction.Type)
			return nil
		},
	}

	service := NewTransactionService(mock)
	ctx := context.Background()
	testDate := time.Date(2026, time.April, 15, 10, 0, 0, 0, time.UTC)

	err := service.CreateTransaction(
		ctx,
		"wallet-1",
		"My Wallet",
		"category-1",
		"Groceries",
		"Weekly grocery shopping",
		"USD",
		"https://example.com/image.png",
		"expense",
		100.0,
		testDate,
		"user-1",
	)
	require.NoError(t, err)
}

func TestCreateTransactionInvalidType(t *testing.T) {
	mock := &mockTransactionRepository{}
	service := NewTransactionService(mock)
	ctx := context.Background()
	testDate := time.Date(2026, time.April, 15, 10, 0, 0, 0, time.UTC)

	err := service.CreateTransaction(
		ctx,
		"wallet-1",
		"My Wallet",
		"category-1",
		"Groceries",
		"Weekly grocery shopping",
		"USD",
		"https://example.com/image.png",
		"invalid",
		100.0,
		testDate,
		"user-1",
	)
	require.Error(t, err)
}

func TestCreateTransactionMissingAmount(t *testing.T) {
	mock := &mockTransactionRepository{}
	service := NewTransactionService(mock)
	ctx := context.Background()
	testDate := time.Date(2026, time.April, 15, 10, 0, 0, 0, time.UTC)

	err := service.CreateTransaction(
		ctx,
		"wallet-1",
		"My Wallet",
		"category-1",
		"Groceries",
		"Weekly grocery shopping",
		"USD",
		"https://example.com/image.png",
		"expense",
		0, // amount is 0
		testDate,
		"user-1",
	)
	require.Error(t, err)
}

func TestCreateTransactionMissingCurrency(t *testing.T) {
	mock := &mockTransactionRepository{}
	service := NewTransactionService(mock)
	ctx := context.Background()
	testDate := time.Date(2026, time.April, 15, 10, 0, 0, 0, time.UTC)

	err := service.CreateTransaction(
		ctx,
		"wallet-1",
		"My Wallet",
		"category-1",
		"Groceries",
		"Weekly grocery shopping",
		"", // empty currency
		"https://example.com/image.png",
		"expense",
		100.0,
		testDate,
		"user-1",
	)
	require.Error(t, err)
}

func TestCreateTransactionMissingWalletID(t *testing.T) {
	mock := &mockTransactionRepository{}
	service := NewTransactionService(mock)
	ctx := context.Background()
	testDate := time.Date(2026, time.April, 15, 10, 0, 0, 0, time.UTC)

	err := service.CreateTransaction(
		ctx,
		"", // empty wallet ID
		"My Wallet",
		"category-1",
		"Groceries",
		"Weekly grocery shopping",
		"USD",
		"https://example.com/image.png",
		"expense",
		100.0,
		testDate,
		"user-1",
	)
	require.Error(t, err)
}

func TestCreateTransactionMissingCategoryID(t *testing.T) {
	mock := &mockTransactionRepository{}
	service := NewTransactionService(mock)
	ctx := context.Background()
	testDate := time.Date(2026, time.April, 15, 10, 0, 0, 0, time.UTC)

	err := service.CreateTransaction(
		ctx,
		"wallet-1",
		"My Wallet",
		"", // empty category ID
		"Groceries",
		"Weekly grocery shopping",
		"USD",
		"https://example.com/image.png",
		"expense",
		100.0,
		testDate,
		"user-1",
	)
	require.Error(t, err)
}

func TestCreateTransactionMissingDate(t *testing.T) {
	mock := &mockTransactionRepository{}
	service := NewTransactionService(mock)
	ctx := context.Background()
	zeroDate := time.Time{} // zero value date

	err := service.CreateTransaction(
		ctx,
		"wallet-1",
		"My Wallet",
		"category-1",
		"Groceries",
		"Weekly grocery shopping",
		"USD",
		"https://example.com/image.png",
		"expense",
		100.0,
		zeroDate,
		"user-1",
	)
	require.Error(t, err)
}

func TestCreateTransactionRepositoryError(t *testing.T) {
	expectedErr := errors.New("database error")
	mock := &mockTransactionRepository{
		createFn: func(ctx context.Context, transaction *model.Transaction) error {
			return expectedErr
		},
	}

	service := NewTransactionService(mock)
	ctx := context.Background()
	testDate := time.Date(2026, time.April, 15, 10, 0, 0, 0, time.UTC)

	err := service.CreateTransaction(
		ctx,
		"wallet-1",
		"My Wallet",
		"category-1",
		"Groceries",
		"Weekly grocery shopping",
		"USD",
		"https://example.com/image.png",
		"expense",
		100.0,
		testDate,
		"user-1",
	)
	require.ErrorIs(t, err, expectedErr)
}

func TestGetTransactionSuccess(t *testing.T) {
	expectedTx := &model.Transaction{
		ID:         "tx-1",
		WalletID:   "wallet-1",
		CategoryID: "category-1",
		Amount:     100.0,
		Currency:   "USD",
		Type:       model.TransactionTypeExpense,
		OwnerID:    "user-1",
	}

	mock := &mockTransactionRepository{
		getByKeyFn: func(ctx context.Context, id string) (*model.Transaction, error) {
			assert.Equal(t, "tx-1", id)
			return expectedTx, nil
		},
	}

	service := NewTransactionService(mock)
	ctx := context.Background()

	tx, err := service.GetTransaction(ctx, "tx-1")
	require.NoError(t, err)
	assert.Equal(t, expectedTx, tx)
}

func TestGetTransactionEmptyID(t *testing.T) {
	mock := &mockTransactionRepository{}
	service := NewTransactionService(mock)
	ctx := context.Background()

	tx, err := service.GetTransaction(ctx, "")
	require.Error(t, err)
	assert.Nil(t, tx)
}

func TestGetTransactionRepositoryError(t *testing.T) {
	expectedErr := errors.New("database error")
	mock := &mockTransactionRepository{
		getByKeyFn: func(ctx context.Context, id string) (*model.Transaction, error) {
			return nil, expectedErr
		},
	}

	service := NewTransactionService(mock)
	ctx := context.Background()

	tx, err := service.GetTransaction(ctx, "tx-1")
	require.ErrorIs(t, err, expectedErr)
	assert.Nil(t, tx)
}

func TestCreateTransactionAllTypes(t *testing.T) {
	testCases := []struct {
		name     string
		txType   string
		wantType model.TransactionType
	}{
		{"income", "income", model.TransactionTypeIncome},
		{"Income uppercase", "INCOME", model.TransactionTypeIncome},
		{"expense", "expense", model.TransactionTypeExpense},
		{"Expense uppercase", "EXPENSE", model.TransactionTypeExpense},
		{"transfer", "transfer", model.TransactionTypeTransfer},
		{"Transfer uppercase", "TRANSFER", model.TransactionTypeTransfer},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mock := &mockTransactionRepository{
				createFn: func(ctx context.Context, transaction *model.Transaction) error {
					assert.Equal(t, tc.wantType, transaction.Type)
					return nil
				},
			}

			service := NewTransactionService(mock)
			ctx := context.Background()
			testDate := time.Date(2026, time.April, 15, 10, 0, 0, 0, time.UTC)

			err := service.CreateTransaction(
				ctx,
				"wallet-1",
				"My Wallet",
				"category-1",
				"Test",
				"description",
				"USD",
				"https://example.com/image.png",
				tc.txType,
				100.0,
				testDate,
				"user-1",
			)
			require.NoError(t, err)
		})
	}
}
