package transaction

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockTransactionRepository implements TransactionRepository for testing.
type mockTransactionRepository struct {
	createTransactionFn          func(ctx context.Context, transaction *Transaction) error
	getTransactionFn             func(ctx context.Context, ownerID string, transactionID string) (*Transaction, error)
	listTransactionsOfSubModelFn func(ctx context.Context, subModelName string, targetID string, ownerID string, fromDate time.Time, toDate time.Time, nextToken string, limit int32) ([]Transaction, string, error)
	editTransactionFn            func(ctx context.Context, ownerID string, transactionID string, changedFields map[string]any) error
	deleteTransactionFn          func(ctx context.Context, ownerID string, transactionID string) error
}

var _ TransactionRepository = (*mockTransactionRepository)(nil)

func (m *mockTransactionRepository) CreateTransaction(ctx context.Context, transaction *Transaction) error {
	if m.createTransactionFn != nil {
		return m.createTransactionFn(ctx, transaction)
	}
	return nil
}

func (m *mockTransactionRepository) GetTransaction(ctx context.Context, ownerID string, transactionID string) (*Transaction, error) {
	if m.getTransactionFn != nil {
		return m.getTransactionFn(ctx, ownerID, transactionID)
	}
	return nil, nil
}

func (m *mockTransactionRepository) EditTransaction(ctx context.Context, ownerID string, transactionID string, changedFields map[string]any) error {
	if m.editTransactionFn != nil {
		return m.editTransactionFn(ctx, ownerID, transactionID, changedFields)
	}
	return nil
}

func (m *mockTransactionRepository) DeleteTransaction(ctx context.Context, ownerID string, transactionID string) error {
	if m.deleteTransactionFn != nil {
		return m.deleteTransactionFn(ctx, ownerID, transactionID)
	}
	return nil
}

func (m *mockTransactionRepository) ListTransactionsOfSubModel(ctx context.Context, subModelName string, targetID string, ownerID string, fromDate time.Time, toDate time.Time, nextToken string, limit int32) ([]Transaction, string, error) {
	if m.listTransactionsOfSubModelFn != nil {
		return m.listTransactionsOfSubModelFn(ctx, subModelName, targetID, ownerID, fromDate, toDate, nextToken, limit)
	}
	return nil, "", nil
}

func TestCreateTransactionSuccess(t *testing.T) {
	mock := &mockTransactionRepository{
		createTransactionFn: func(ctx context.Context, transaction *Transaction) error {
			assert.Equal(t, 100.0, transaction.Amount)
			assert.Equal(t, "USD", transaction.Currency)
			assert.Equal(t, "wallet-1", transaction.WalletID)
			assert.Equal(t, "category-1", transaction.CategoryID)
			assert.Equal(t, TransactionTypeExpense, transaction.Type)
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
	service := NewTransactionService(&mockTransactionRepository{})
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
	service := NewTransactionService(&mockTransactionRepository{})
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
		0,
		testDate,
		"user-1",
	)
	require.Error(t, err)
}

func TestCreateTransactionMissingCurrency(t *testing.T) {
	service := NewTransactionService(&mockTransactionRepository{})
	ctx := context.Background()
	testDate := time.Date(2026, time.April, 15, 10, 0, 0, 0, time.UTC)

	err := service.CreateTransaction(
		ctx,
		"wallet-1",
		"My Wallet",
		"category-1",
		"Groceries",
		"Weekly grocery shopping",
		"",
		"https://example.com/image.png",
		"expense",
		100.0,
		testDate,
		"user-1",
	)
	require.Error(t, err)
}

func TestCreateTransactionMissingWalletID(t *testing.T) {
	service := NewTransactionService(&mockTransactionRepository{})
	ctx := context.Background()
	testDate := time.Date(2026, time.April, 15, 10, 0, 0, 0, time.UTC)

	err := service.CreateTransaction(
		ctx,
		"",
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
	service := NewTransactionService(&mockTransactionRepository{})
	ctx := context.Background()
	testDate := time.Date(2026, time.April, 15, 10, 0, 0, 0, time.UTC)

	err := service.CreateTransaction(
		ctx,
		"wallet-1",
		"My Wallet",
		"",
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
	service := NewTransactionService(&mockTransactionRepository{})
	ctx := context.Background()
	zeroDate := time.Time{}

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
		createTransactionFn: func(ctx context.Context, transaction *Transaction) error {
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
	expectedTx := &Transaction{
		PK:         "USER#1",
		ID:         "tx-1",
		WalletID:   "wallet-1",
		CategoryID: "category-1",
		Amount:     100.0,
		Currency:   "USD",
		Type:       TransactionTypeExpense,
		OwnerID:    "user-1",
	}

	mock := &mockTransactionRepository{
		getTransactionFn: func(ctx context.Context, ownerID string, transactionID string) (*Transaction, error) {
			assert.Equal(t, "user-1", ownerID)
			assert.Equal(t, "tx-1", transactionID)
			return expectedTx, nil
		},
	}

	service := NewTransactionService(mock)
	ctx := context.Background()

	tx, err := service.GetTransaction(ctx, "tx-1", "user-1")
	require.NoError(t, err)
	assert.Equal(t, expectedTx, tx)
}

func TestGetTransactionEmptyID(t *testing.T) {
	service := NewTransactionService(&mockTransactionRepository{})
	ctx := context.Background()

	tx, err := service.GetTransaction(ctx, "", "")
	require.Error(t, err)
	assert.Nil(t, tx)
}

func TestGetTransactionRepositoryError(t *testing.T) {
	expectedErr := errors.New("database error")
	mock := &mockTransactionRepository{
		getTransactionFn: func(ctx context.Context, ownerID string, transactionID string) (*Transaction, error) {
			return nil, expectedErr
		},
	}

	service := NewTransactionService(mock)
	ctx := context.Background()

	tx, err := service.GetTransaction(ctx, "tx-1", "user-1")
	require.ErrorIs(t, err, expectedErr)
	assert.Nil(t, tx)
}

func TestCreateTransactionAllTypes(t *testing.T) {
	testCases := []struct {
		name     string
		txType   string
		wantType TransactionType
	}{
		{"income", "income", TransactionTypeIncome},
		{"Income uppercase", "INCOME", TransactionTypeIncome},
		{"expense", "expense", TransactionTypeExpense},
		{"Expense uppercase", "EXPENSE", TransactionTypeExpense},
		{"transfer", "transfer", TransactionTypeTransfer},
		{"Transfer uppercase", "TRANSFER", TransactionTypeTransfer},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mock := &mockTransactionRepository{
				createTransactionFn: func(ctx context.Context, transaction *Transaction) error {
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

func TestListTransactionsDefaultsAndDelegatesToRepository(t *testing.T) {
	from := time.Date(2026, time.April, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, time.April, 30, 0, 0, 0, 0, time.UTC)

	mock := &mockTransactionRepository{
		listTransactionsOfSubModelFn: func(ctx context.Context, subModelName string, targetID string, ownerID string, fromDate time.Time, toDate time.Time, nextToken string, limit int32) ([]Transaction, string, error) {
			assert.Equal(t, "date", subModelName)
			assert.Equal(t, "", targetID)
			assert.Equal(t, "user-1", ownerID)
			assert.Equal(t, from, fromDate)
			assert.Equal(t, to, toDate)
			assert.Equal(t, int32(25), limit)
			assert.Equal(t, "token-1", nextToken)
			return []Transaction{{ID: "tx-1"}}, "token-2", nil
		},
	}

	service := NewTransactionService(mock)
	txs, token, err := service.ListTransactions(context.Background(), "user-1", from, to, 25, "token-1")
	require.NoError(t, err)
	assert.Len(t, txs, 1)
	assert.Equal(t, "tx-1", txs[0].ID)
	assert.Equal(t, "token-2", token)
}

func TestListTransactionsValidation(t *testing.T) {
	service := NewTransactionService(&mockTransactionRepository{})
	from := time.Date(2026, time.April, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, time.April, 30, 0, 0, 0, 0, time.UTC)

	_, _, err := service.ListTransactions(context.Background(), "", from, to, 10, "")
	require.EqualError(t, err, "owner ID is required")

	_, _, err = service.ListTransactions(context.Background(), "user-1", time.Time{}, to, 10, "")
	require.NoError(t, err)

	_, _, err = service.ListTransactions(context.Background(), "user-1", from, to, -1, "")
	require.EqualError(t, err, "limit must be greater than or equal to 0")
}

func TestListTransactionsOfCategoryDelegatesToRepository(t *testing.T) {
	from := time.Date(2026, time.April, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, time.April, 30, 0, 0, 0, 0, time.UTC)

	mock := &mockTransactionRepository{
		listTransactionsOfSubModelFn: func(ctx context.Context, subModelName string, targetID string, ownerID string, fromDate time.Time, toDate time.Time, nextToken string, limit int32) ([]Transaction, string, error) {
			assert.Equal(t, "category", subModelName)
			assert.Equal(t, "category-1", targetID)
			assert.Equal(t, "user-1", ownerID)
			assert.Equal(t, from, fromDate)
			assert.Equal(t, to, toDate)
			assert.Equal(t, int32(15), limit)
			assert.Equal(t, "token-1", nextToken)
			return []Transaction{{ID: "tx-2"}}, "token-3", nil
		},
	}

	service := NewTransactionService(mock)
	txs, token, err := service.ListTransactionsOfCategory(context.Background(), "user-1", from, to, 15, "token-1", "category-1")
	require.NoError(t, err)
	assert.Len(t, txs, 1)
	assert.Equal(t, "tx-2", txs[0].ID)
	assert.Equal(t, "token-3", token)
}

func TestListTransactionsOfWalletDelegatesToRepository(t *testing.T) {
	from := time.Date(2026, time.April, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, time.April, 30, 0, 0, 0, 0, time.UTC)

	mock := &mockTransactionRepository{
		listTransactionsOfSubModelFn: func(ctx context.Context, subModelName string, targetID string, ownerID string, fromDate time.Time, toDate time.Time, nextToken string, limit int32) ([]Transaction, string, error) {
			assert.Equal(t, "wallet", subModelName)
			assert.Equal(t, "wallet-1", targetID)
			assert.Equal(t, "user-1", ownerID)
			assert.Equal(t, from, fromDate)
			assert.Equal(t, to, toDate)
			assert.Equal(t, int32(20), limit)
			assert.Equal(t, "token-2", nextToken)
			return []Transaction{{ID: "tx-3"}}, "token-4", nil
		},
	}

	service := NewTransactionService(mock)
	txs, token, err := service.ListTransactionsOfWallet(context.Background(), "user-1", from, to, 20, "token-2", "wallet-1")
	require.NoError(t, err)
	assert.Len(t, txs, 1)
	assert.Equal(t, "tx-3", txs[0].ID)
	assert.Equal(t, "token-4", token)
}

func TestEditTransactionRejectsOwnerIDAndNormalizesType(t *testing.T) {
	service := NewTransactionService(&mockTransactionRepository{})

	err := service.EditTransaction(context.Background(), "user-1", "tx-1", map[string]any{"OwnerID": "user-2"})
	require.EqualError(t, err, "Owner ID is not updateable")

	err = service.EditTransaction(context.Background(), "user-1", "tx-1", map[string]any{"Type": "income"})
	require.NoError(t, err)
}

func TestEditTransactionDelegatesToRepository(t *testing.T) {
	mock := &mockTransactionRepository{
		editTransactionFn: func(ctx context.Context, ownerID string, transactionID string, changedFields map[string]any) error {
			assert.Equal(t, "user-1", ownerID)
			assert.Equal(t, "tx-1", transactionID)
			assert.Equal(t, TransactionTypeIncome, changedFields["Type"])
			return nil
		},
	}

	service := NewTransactionService(mock)
	err := service.EditTransaction(context.Background(), "user-1", "tx-1", map[string]any{"Type": "income"})
	require.NoError(t, err)
}

func TestDeleteTransactionDelegatesToRepository(t *testing.T) {
	mock := &mockTransactionRepository{
		deleteTransactionFn: func(ctx context.Context, ownerID string, transactionID string) error {
			assert.Equal(t, "user-1", ownerID)
			assert.Equal(t, "tx-1", transactionID)
			return nil
		},
	}

	service := NewTransactionService(mock)
	err := service.DeleteTransaction(context.Background(), "user-1", "tx-1")
	require.NoError(t, err)
}
