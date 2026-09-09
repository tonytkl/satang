package transaction

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tonytkl/satang/clients"
)

type mockDynamoDB struct {
	putItemFn                  func(ctx context.Context, table string, item any) error
	updateItemFn               func(ctx context.Context, table string, key map[string]any, updateExpression string, expressionValues map[string]any, expressionNames map[string]string, conditionExpression string) error
	getItemFn                  func(ctx context.Context, table string, key map[string]any, out any) error
	deleteItemFn               func(ctx context.Context, table string, key map[string]any) error
	queryItemsFn               func(ctx context.Context, table string, keyConditionExpression string, expressionValues map[string]any, indexName string, filterExpression string, out any) error
	queryItemsWithPaginationFn func(ctx context.Context, table string, keyConditionExpression string, expressionValues map[string]any, indexName string, filterExpression string, limit int32, nextToken string, out any) (string, error)
	scanItemsFn                func(ctx context.Context, table string, filterExpression string, expressionValues map[string]any, out any) error
}

var _ clients.DynamoDBClient = (*mockDynamoDB)(nil)

func (m *mockDynamoDB) PutItem(ctx context.Context, table string, item any) error {
	if m.putItemFn != nil {
		return m.putItemFn(ctx, table, item)
	}
	return nil
}

func (m *mockDynamoDB) UpdateItem(ctx context.Context, table string, key map[string]any, updateExpression string, expressionValues map[string]any, expressionNames map[string]string, conditionExpression string) error {
	if m.updateItemFn != nil {
		return m.updateItemFn(ctx, table, key, updateExpression, expressionValues, expressionNames, conditionExpression)
	}
	return nil
}

func (m *mockDynamoDB) GetItem(ctx context.Context, table string, key map[string]any, out any) error {
	if m.getItemFn != nil {
		return m.getItemFn(ctx, table, key, out)
	}
	return nil
}

func (m *mockDynamoDB) DeleteItem(ctx context.Context, table string, key map[string]any) error {
	if m.deleteItemFn != nil {
		return m.deleteItemFn(ctx, table, key)
	}
	return nil
}

func (m *mockDynamoDB) QueryItems(ctx context.Context, table string, keyConditionExpression string, expressionValues map[string]any, indexName string, filterExpression string, out any) error {
	if m.queryItemsFn != nil {
		return m.queryItemsFn(ctx, table, keyConditionExpression, expressionValues, indexName, filterExpression, out)
	}
	return nil
}

func (m *mockDynamoDB) QueryItemsWithPagination(ctx context.Context, table string, keyConditionExpression string, expressionValues map[string]any, indexName string, filterExpression string, limit int32, nextToken string, out any) (string, error) {
	if m.queryItemsWithPaginationFn != nil {
		return m.queryItemsWithPaginationFn(ctx, table, keyConditionExpression, expressionValues, indexName, filterExpression, limit, nextToken, out)
	}

	if m.queryItemsFn != nil {
		if err := m.queryItemsFn(ctx, table, keyConditionExpression, expressionValues, indexName, filterExpression, out); err != nil {
			return "", err
		}
	}

	return "", nil
}

func (m *mockDynamoDB) ScanItems(ctx context.Context, table string, filterExpression string, expressionValues map[string]any, out any) error {
	if m.scanItemsFn != nil {
		return m.scanItemsFn(ctx, table, filterExpression, expressionValues, out)
	}
	return nil
}

func findPlaceholderByAttribute(t *testing.T, expressionNames map[string]string, attribute string) string {
	t.Helper()

	for placeholder, name := range expressionNames {
		if name == attribute {
			return placeholder
		}
	}

	t.Fatalf("attribute %q not found in expression names %#v", attribute, expressionNames)
	return ""
}

func TestTransactionRepositoryCreateSuccess(t *testing.T) {
	mock := &mockDynamoDB{
		putItemFn: func(_ context.Context, table string, item any) error {
			require.Equal(t, "transactions", table)

			tx, ok := item.(*Transaction)
			require.True(t, ok, "item type = %T, want *Transaction", item)

			assert.Equal(t, "USER#user-1", tx.PK)
			assert.Equal(t, "TX#tx-1", tx.SK)
			assert.Equal(t, "USER#user-1", tx.GSI_ByDatePK)
			assert.Equal(t, "TX#2026-04-15#tx-1", tx.GSI_ByDateSK)
			assert.Equal(t, "USER#user-1#TX_CATEGORY#cat-1", tx.GSI_ByCategoryPK)
			assert.Equal(t, "TX#2026-04-15#tx-1", tx.GSI_ByCategorySK)
			assert.Equal(t, "USER#user-1#TX_WALLET#wallet-1", tx.GSI_ByWalletPK)
			assert.Equal(t, "TX#2026-04-15#tx-1", tx.GSI_ByWalletSK)
			assert.False(t, tx.CreatedAt.IsZero())
			assert.False(t, tx.UpdatedAt.IsZero())

			return nil
		},
	}

	repo := NewTransactionRepository(mock, "transactions")
	tx := &Transaction{
		ID:         "tx-1",
		OwnerID:    "user-1",
		WalletID:   "wallet-1",
		CategoryID: "cat-1",
		Amount:     100,
		Currency:   "THB",
		Date:       time.Date(2026, 4, 15, 10, 0, 0, 0, time.UTC),
	}

	err := repo.CreateTransaction(context.Background(), tx)
	require.NoError(t, err)
}

func TestTransactionRepositoryGetTransactionSuccess(t *testing.T) {
	mock := &mockDynamoDB{
		getItemFn: func(_ context.Context, table string, key map[string]any, out any) error {
			require.Equal(t, "transactions", table)
			assert.Equal(t, "USER#user-1", key["PK"])
			assert.Equal(t, "TX#tx-1", key["SK"])

			dst, ok := out.(*Transaction)
			require.True(t, ok, "out type = %T, want *Transaction", out)
			*dst = Transaction{ID: "tx-1", PK: "USER#user-1"}

			return nil
		},
	}

	repo := NewTransactionRepository(mock, "transactions")

	got, err := repo.GetTransaction(context.Background(), "user-1", "tx-1")
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "tx-1", got.ID)
	assert.Equal(t, "USER#user-1", got.PK)
}

func TestTransactionRepositoryGetTransactionErrors(t *testing.T) {
	repo := NewTransactionRepository(&mockDynamoDB{}, "transactions")

	_, err := repo.GetTransaction(context.Background(), "", "tx-1")
	require.EqualError(t, err, "owner ID is required")

	_, err = repo.GetTransaction(context.Background(), "user-1", "")
	require.EqualError(t, err, "item ID is required")
}

func TestTransactionRepositoryListTransactionsOfSubModelDateSuccess(t *testing.T) {
	mock := &mockDynamoDB{
		queryItemsWithPaginationFn: func(_ context.Context, table string, keyConditionExpression string, expressionValues map[string]any, indexName string, filterExpression string, limit int32, nextToken string, out any) (string, error) {
			require.Equal(t, "transactions", table)
			require.Equal(t, "GSI1", indexName)
			require.Equal(t, "GSI1_PK = :indexPK AND GSI1_SK BETWEEN :from AND :to", keyConditionExpression)
			require.Empty(t, filterExpression)
			require.Equal(t, int32(20), limit)
			require.Equal(t, "token-1", nextToken)
			assert.Equal(t, "USER#user-1", expressionValues[":indexPK"])
			assert.Equal(t, "TX#2026-04-01#", expressionValues[":from"])
			assert.Equal(t, "TX#2026-04-30#", expressionValues[":to"])

			dst := out.(*[]Transaction)
			*dst = []Transaction{{ID: "tx-1"}, {ID: "tx-2"}}
			return "token-2", nil
		},
	}

	repo := NewTransactionRepository(mock, "transactions")
	from := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 4, 30, 0, 0, 0, 0, time.UTC)

	got, nextToken, err := repo.ListTransactionsOfSubModel(context.Background(), "date", "", "user-1", from, to, "token-1", 20)
	require.NoError(t, err)
	assert.Len(t, got, 2)
	assert.Equal(t, "token-2", nextToken)
}

func TestTransactionRepositoryListTransactionsOfSubModelCategorySuccess(t *testing.T) {
	mock := &mockDynamoDB{
		queryItemsWithPaginationFn: func(_ context.Context, _ string, keyConditionExpression string, expressionValues map[string]any, indexName string, filterExpression string, limit int32, nextToken string, out any) (string, error) {
			require.Equal(t, "GSI3_PK = :indexPK AND GSI3_SK BETWEEN :from AND :to", keyConditionExpression)
			require.Equal(t, "GSI3", indexName)
			require.Empty(t, filterExpression)
			require.Equal(t, int32(10), limit)
			require.Empty(t, nextToken)
			assert.Equal(t, "USER#user-1#TX_CATEGORY#cat-1", expressionValues[":indexPK"])
			assert.Equal(t, "TX#2026-04-01#", expressionValues[":from"])
			assert.Equal(t, "TX#2026-04-30#", expressionValues[":to"])

			dst := out.(*[]Transaction)
			*dst = []Transaction{{ID: "tx-1"}}
			return "", nil
		},
	}

	repo := NewTransactionRepository(mock, "transactions")
	from := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 4, 30, 0, 0, 0, 0, time.UTC)

	got, nextToken, err := repo.ListTransactionsOfSubModel(context.Background(), "category", "cat-1", "user-1", from, to, "", 10)
	require.NoError(t, err)
	assert.Len(t, got, 1)
	assert.Equal(t, "", nextToken)
}

func TestTransactionRepositoryListTransactionsOfSubModelErrors(t *testing.T) {
	repo := NewTransactionRepository(&mockDynamoDB{}, "transactions")
	from := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 4, 30, 0, 0, 0, 0, time.UTC)

	_, _, err := repo.ListTransactionsOfSubModel(context.Background(), "unsupported", "cat-1", "user-1", from, to, "", 10)
	require.EqualError(t, err, "unsupported model name: unsupported")

	_, _, err = repo.ListTransactionsOfSubModel(context.Background(), "date", "", "user-1", to, from, "", 10)
	require.EqualError(t, err, "from date must not be after to date")
}

func TestTransactionRepositoryEditTransactionSuccess(t *testing.T) {
	desc := "new description"
	updatedDate := time.Date(2026, 4, 20, 12, 0, 0, 0, time.UTC)

	mock := &mockDynamoDB{
		updateItemFn: func(_ context.Context, table string, key map[string]any, updateExpression string, expressionValues map[string]any, expressionNames map[string]string, conditionExpression string) error {
			require.Equal(t, "transactions", table)
			assert.Equal(t, "USER#user-1", key["PK"])
			assert.Equal(t, "TX#tx-1", key["SK"])

			wantCond := "attribute_exists(PK) AND attribute_exists(SK) AND ID = :id"
			require.Equal(t, wantCond, conditionExpression)

			walletNameField := findPlaceholderByAttribute(t, expressionNames, "WalletID")
			walletValueField := strings.Replace(walletNameField, "#n", ":v", 1)
			assert.Contains(t, updateExpression, walletNameField+" = "+walletValueField)
			assert.Equal(t, "wallet-2", expressionValues[walletValueField])

			categoryPKNameField := findPlaceholderByAttribute(t, expressionNames, "GSI_ByCategoryPK")
			categoryPKValueField := strings.Replace(categoryPKNameField, "#n", ":v", 1)
			assert.Equal(t, "USER#user-1#TX_CATEGORY#cat-2", expressionValues[categoryPKValueField])

			walletPKNameField := findPlaceholderByAttribute(t, expressionNames, "GSI_ByWalletPK")
			walletPKValueField := strings.Replace(walletPKNameField, "#n", ":v", 1)
			assert.Equal(t, "USER#user-1#TX_WALLET#wallet-2", expressionValues[walletPKValueField])

			dateSKNameField := findPlaceholderByAttribute(t, expressionNames, "GSI_ByDateSK")
			dateSKValueField := strings.Replace(dateSKNameField, "#n", ":v", 1)
			assert.Equal(t, "TX#2026-04-20#tx-1", expressionValues[dateSKValueField])

			assert.Equal(t, "tx-1", expressionValues[":id"])

			updatedAtNameField := findPlaceholderByAttribute(t, expressionNames, "UpdatedAt")
			updatedAtValueField := strings.Replace(updatedAtNameField, "#n", ":v", 1)
			updatedAt, ok := expressionValues[updatedAtValueField].(time.Time)
			require.True(t, ok)
			assert.False(t, updatedAt.IsZero())

			return nil
		},
	}

	repo := NewTransactionRepository(mock, "transactions")
	err := repo.EditTransaction(context.Background(), "user-1", "tx-1", map[string]any{
		"WalletID":    "wallet-2",
		"CategoryID":  "cat-2",
		"Date":        updatedDate,
		"Description": &desc,
	})
	require.NoError(t, err)
}

func TestTransactionRepositoryEditTransactionDateTypeError(t *testing.T) {
	repo := NewTransactionRepository(&mockDynamoDB{}, "transactions")

	err := repo.EditTransaction(context.Background(), "user-1", "tx-1", map[string]any{
		"Date": "2026-04-20",
	})
	require.EqualError(t, err, "Date must be time.Time")
}

func TestTransactionRepositoryDeleteSuccess(t *testing.T) {
	mock := &mockDynamoDB{
		deleteItemFn: func(_ context.Context, table string, key map[string]any) error {
			require.Equal(t, "transactions", table)
			assert.Equal(t, "USER#user-1", key["PK"])
			assert.Equal(t, "TX#tx-1", key["SK"])
			return nil
		},
	}

	repo := NewTransactionRepository(mock, "transactions")
	err := repo.DeleteTransaction(context.Background(), "user-1", "tx-1")
	require.NoError(t, err)
}

func TestTransactionRepositoryEditAndDeleteErrorPaths(t *testing.T) {
	repo := NewTransactionRepository(&mockDynamoDB{}, "transactions")

	err := repo.EditTransaction(context.Background(), "", "tx-1", map[string]any{"WalletID": "wallet-1"})
	require.EqualError(t, err, "owner ID is required")

	err = repo.DeleteTransaction(context.Background(), "user-1", "")
	require.EqualError(t, err, "item ID is required")
}

func TestTransactionRepositoryDBErrorWrapping(t *testing.T) {
	dbErr := errors.New("dynamodb failed")

	mock := &mockDynamoDB{
		getItemFn: func(_ context.Context, _ string, _ map[string]any, _ any) error {
			return dbErr
		},
		updateItemFn: func(_ context.Context, _ string, _ map[string]any, _ string, _ map[string]any, _ map[string]string, _ string) error {
			return dbErr
		},
		deleteItemFn: func(_ context.Context, _ string, _ map[string]any) error {
			return dbErr
		},
		queryItemsWithPaginationFn: func(_ context.Context, _ string, _ string, _ map[string]any, _ string, _ string, _ int32, _ string, _ any) (string, error) {
			return "", dbErr
		},
	}

	repo := NewTransactionRepository(mock, "transactions")
	from := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 4, 30, 0, 0, 0, 0, time.UTC)

	_, err := repo.GetTransaction(context.Background(), "user-1", "tx-1")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Error on Get")
	assert.ErrorIs(t, err, dbErr)

	err = repo.EditTransaction(context.Background(), "user-1", "tx-1", map[string]any{
		"WalletID":   "wallet-1",
		"CategoryID": "cat-1",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "update item")
	assert.ErrorIs(t, err, dbErr)

	_, _, err = repo.ListTransactionsOfSubModel(context.Background(), "category", "cat-1", "user-1", from, to, "", 10)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "query transaction by ID")
	assert.ErrorIs(t, err, dbErr)

	err = repo.DeleteTransaction(context.Background(), "user-1", "tx-1")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Error on Delete")
	assert.ErrorIs(t, err, dbErr)
}
