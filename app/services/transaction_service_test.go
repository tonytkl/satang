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
	createFn              func(ctx context.Context, transaction *model.Transaction) error
	getByKeyFn            func(ctx context.Context, id string, ownerID string) (*model.Transaction, error)
	listWithinDateRangeFn func(ctx context.Context, ownerID string, fromDate time.Time, toDate time.Time, limit int32, nextToken string) ([]model.Transaction, string, error)
}

var _ repositories.TransactionRepository = (*mockTransactionRepository)(nil)

func (m *mockTransactionRepository) Create(ctx context.Context, transaction *model.Transaction) error {
	if m.createFn != nil {
		return m.createFn(ctx, transaction)
	}
	return nil
}

func (m *mockTransactionRepository) GetByKey(ctx context.Context, id string, ownerID string) (*model.Transaction, error) {
	if m.getByKeyFn != nil {
		return m.getByKeyFn(ctx, id, ownerID)
	}
	return nil, nil
}

func (m *mockTransactionRepository) ListByGSI(ctx context.Context, indexName string, indexPartitionKeyPrefix string, targetID string, ownerID string, fromDate *time.Time, toDate *time.Time) ([]model.Transaction, error) {
	return nil, nil
}

func (m *mockTransactionRepository) ListWithinDateRange(ctx context.Context, ownerID string, fromDate time.Time, toDate time.Time, limit int32, nextToken string) ([]model.Transaction, string, error) {
	if m.listWithinDateRangeFn != nil {
		return m.listWithinDateRangeFn(ctx, ownerID, fromDate, toDate, limit, nextToken)
	}
	return nil, "", nil
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
		PK:         "USER#1",
		ID:         "tx-1",
		WalletID:   "wallet-1",
		CategoryID: "category-1",
		Amount:     100.0,
		Currency:   "USD",
		Type:       model.TransactionTypeExpense,
		OwnerID:    "user-1",
	}

	mock := &mockTransactionRepository{
		getByKeyFn: func(ctx context.Context, id string, ownerID string) (*model.Transaction, error) {
			assert.Equal(t, "tx-1", id)
			return expectedTx, nil
		},
	}

	service := NewTransactionService(mock)
	ctx := context.Background()

	tx, err := service.GetTransaction(ctx, "tx-1", "1")
	require.NoError(t, err)
	assert.Equal(t, expectedTx, tx)
}

func TestGetTransactionEmptyID(t *testing.T) {
	mock := &mockTransactionRepository{}
	service := NewTransactionService(mock)
	ctx := context.Background()

	tx, err := service.GetTransaction(ctx, "", "")
	require.Error(t, err)
	assert.Nil(t, tx)
}

func TestGetTransactionRepositoryError(t *testing.T) {
	expectedErr := errors.New("database error")
	mock := &mockTransactionRepository{
		getByKeyFn: func(ctx context.Context, id string, ownerID string) (*model.Transaction, error) {
			return nil, expectedErr
		},
	}

	service := NewTransactionService(mock)
	ctx := context.Background()

	tx, err := service.GetTransaction(ctx, "tx-1", "1")
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

func TestGetTransactionsBetweenPeriodSuccess(t *testing.T) {
	from := time.Date(2026, time.April, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, time.April, 30, 0, 0, 0, 0, time.UTC)

	mock := &mockTransactionRepository{
		listWithinDateRangeFn: func(ctx context.Context, ownerID string, fromDate time.Time, toDate time.Time, limit int32, nextToken string) ([]model.Transaction, string, error) {
			assert.Equal(t, "user-1", ownerID)
			assert.Equal(t, from, fromDate)
			assert.Equal(t, to, toDate)
			assert.Equal(t, int32(25), limit)
			assert.Equal(t, "token-1", nextToken)

			return []model.Transaction{{ID: "tx-1"}}, "token-2", nil
		},
	}

	service := NewTransactionService(mock)
	txs, token, err := service.GetTransactionsBetweenPeriod(context.Background(), "user-1", from, to, 25, "token-1")
	require.NoError(t, err)
	assert.Len(t, txs, 1)
	assert.Equal(t, "tx-1", txs[0].ID)
	assert.Equal(t, "token-2", token)
}

func TestGetTransactionsBetweenPeriodValidation(t *testing.T) {
	service := NewTransactionService(&mockTransactionRepository{})
	from := time.Date(2026, time.April, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, time.April, 30, 0, 0, 0, 0, time.UTC)

	_, _, err := service.GetTransactionsBetweenPeriod(context.Background(), "", from, to, 10, "")
	require.EqualError(t, err, "owner ID is required")

	_, _, err = service.GetTransactionsBetweenPeriod(context.Background(), "user-1", time.Time{}, to, 10, "")
	require.EqualError(t, err, "from date and to date are required")

	_, _, err = service.GetTransactionsBetweenPeriod(context.Background(), "user-1", from, to, -1, "")
	require.EqualError(t, err, "limit must be greater than or equal to 0")
}
