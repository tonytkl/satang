package repositories

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tonytkl/satang/clients"
	"github.com/tonytkl/satang/model"
)

type mockDynamoDB struct {
	putItemFn    func(ctx context.Context, table string, item any) error
	updateItemFn func(ctx context.Context, table string, key map[string]any, updateExpression string, expressionValues map[string]any, conditionExpression string) error
	getItemFn    func(ctx context.Context, table string, key map[string]any, out any) error
	deleteItemFn func(ctx context.Context, table string, key map[string]any) error
	queryItemsFn func(ctx context.Context, table string, keyConditionExpression string, expressionValues map[string]any, indexName string, filterExpression string, out any) error
	scanItemsFn  func(ctx context.Context, table string, filterExpression string, expressionValues map[string]any, out any) error
}

var _ clients.DynamoDBClient = (*mockDynamoDB)(nil)

func (m *mockDynamoDB) PutItem(ctx context.Context, table string, item any) error {
	if m.putItemFn != nil {
		return m.putItemFn(ctx, table, item)
	}
	return nil
}

func (m *mockDynamoDB) UpdateItem(ctx context.Context, table string, key map[string]any, updateExpression string, expressionValues map[string]any, conditionExpression string) error {
	if m.updateItemFn != nil {
		return m.updateItemFn(ctx, table, key, updateExpression, expressionValues, conditionExpression)
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

func (m *mockDynamoDB) ScanItems(ctx context.Context, table string, filterExpression string, expressionValues map[string]any, out any) error {
	if m.scanItemsFn != nil {
		return m.scanItemsFn(ctx, table, filterExpression, expressionValues, out)
	}
	return nil
}

func TestTransactionRepositoryCreateSuccess(t *testing.T) {
	mock := &mockDynamoDB{
		putItemFn: func(_ context.Context, table string, item any) error {
			require.Equal(t, "transactions", table)

			tx, ok := item.(*model.Transaction)
			require.True(t, ok, "item type = %T, want *model.Transaction", item)

			assert.Equal(t, "USER#user-1", tx.PK)
			assert.Equal(t, "TX#2026-04-15#tx-1", tx.SK)
			assert.Equal(t, "TX_CATEGORY#cat-1", tx.GSI_ByCategoryPK)
			assert.Equal(t, "TX#2026-04-15#tx-1", tx.GSI_ByCategorySK)
			assert.Equal(t, "TX_WALLET#wallet-1", tx.GSI_ByWalletPK)
			assert.Equal(t, "TX#2026-04-15#tx-1", tx.GSI_ByWalletSK)
			assert.Equal(t, "TX_ID#tx-1", tx.GSI_ByTransactionID)
			assert.Equal(t, "TX#2026-04-15#tx-1", tx.GSI_ByTransactionSK)
			assert.False(t, tx.CreatedAt.IsZero())
			assert.False(t, tx.UpdatedAt.IsZero())

			return nil
		},
	}

	repo := NewTransactionRepository(mock, "transactions")
	tx := &model.Transaction{
		ID:         "tx-1",
		OwnerID:    "user-1",
		WalletID:   "wallet-1",
		CategoryID: "cat-1",
		Amount:     100,
		Currency:   "THB",
		Date:       time.Date(2026, 4, 15, 10, 0, 0, 0, time.UTC),
	}

	err := repo.Create(context.Background(), tx)
	require.NoError(t, err)
}

func TestTransactionRepositoryListByGSISuccess(t *testing.T) {
	mock := &mockDynamoDB{
		queryItemsFn: func(_ context.Context, table string, keyConditionExpression string, expressionValues map[string]any, indexName string, filterExpression string, out any) error {
			require.Equal(t, "transactions", table)
			require.Equal(t, "GSI1", indexName)
			require.Equal(t, "GSI_PK = :indexPK AND GSI_SK BETWEEN :from AND :to", keyConditionExpression)
			require.Equal(t, "PK = :ownerPK", filterExpression)
			assert.Equal(t, "TX_CATEGORY#cat-1", expressionValues[":indexPK"])
			assert.Equal(t, "USER#user-1", expressionValues[":ownerPK"])
			assert.Equal(t, "TX#2026-04-01#", expressionValues[":from"])
			assert.Equal(t, "TX#2026-04-30#", expressionValues[":to"])

			dst, ok := out.(*[]model.Transaction)
			require.True(t, ok, "out type = %T, want *[]model.Transaction", out)
			*dst = []model.Transaction{{ID: "tx-1"}}

			return nil
		},
	}

	repo := NewTransactionRepository(mock, "transactions")
	from := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 4, 30, 0, 0, 0, 0, time.UTC)

	got, err := repo.ListByGSI(context.Background(), "GSI1", "TX_CATEGORY", "cat-1", "user-1", &from, &to)
	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, "tx-1", got[0].ID)
}

func TestTransactionRepositoryListByGSIErrors(t *testing.T) {
	repo := NewTransactionRepository(&mockDynamoDB{}, "transactions")

	_, err := repo.ListByGSI(context.Background(), "", "TX_CATEGORY", "cat-1", "", nil, nil)
	require.EqualError(t, err, "index name, index partition key prefix, and target ID are required")

	_, err = repo.ListByGSI(context.Background(), "BAD_INDEX", "TX_CATEGORY", "cat-1", "", nil, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported index name")

	from := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)
	_, err = repo.ListByGSI(context.Background(), "GSI1", "TX_CATEGORY", "cat-1", "", &from, &to)
	require.EqualError(t, err, "from date must not be after to date")
}

func TestTransactionRepositoryListByGSINotFound(t *testing.T) {
	mock := &mockDynamoDB{
		queryItemsFn: func(_ context.Context, _ string, _ string, _ map[string]any, _ string, _ string, out any) error {
			dst := out.(*[]model.Transaction)
			*dst = []model.Transaction{}
			return nil
		},
	}

	repo := NewTransactionRepository(mock, "transactions")

	got, err := repo.ListByGSI(context.Background(), "GSI3", "TX_ID", "tx-1", "", nil, nil)
	require.NoError(t, err)
	assert.Len(t, got, 0)
}

func TestTransactionRepositoryListWithinDateRangeSuccess(t *testing.T) {
	mock := &mockDynamoDB{
		queryItemsFn: func(_ context.Context, _ string, keyConditionExpression string, expressionValues map[string]any, indexName string, filterExpression string, out any) error {
			require.Equal(t, "PK = :pk AND SK BETWEEN :from AND :to", keyConditionExpression)
			require.Empty(t, indexName)
			require.Empty(t, filterExpression)
			assert.Equal(t, "USER#user-1", expressionValues[":pk"])
			assert.Equal(t, "TX#2026-04-01#", expressionValues[":from"])
			assert.Equal(t, "TX#2026-04-30#", expressionValues[":to"])

			dst := out.(*[]model.Transaction)
			*dst = []model.Transaction{{ID: "tx-1"}, {ID: "tx-2"}}
			return nil
		},
	}

	repo := NewTransactionRepository(mock, "transactions")
	from := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 4, 30, 0, 0, 0, 0, time.UTC)

	got, err := repo.ListWithinDateRange(context.Background(), "user-1", from, to)
	require.NoError(t, err)
	assert.Len(t, got, 2)
}

func TestTransactionRepositoryGetByKeySuccess(t *testing.T) {
	mock := &mockDynamoDB{
		queryItemsFn: func(_ context.Context, _ string, keyConditionExpression string, expressionValues map[string]any, indexName string, filterExpression string, out any) error {
			require.Equal(t, "GSI3", indexName)
			require.Equal(t, "GSI3_PK = :indexPK", keyConditionExpression)
			require.Empty(t, filterExpression)
			assert.Equal(t, "TX_ID#tx-1", expressionValues[":indexPK"])

			dst := out.(*[]model.Transaction)
			*dst = []model.Transaction{{ID: "tx-1"}}
			return nil
		},
	}

	repo := NewTransactionRepository(mock, "transactions")

	got, err := repo.GetByKey(context.Background(), "tx-1")
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "tx-1", got.ID)
}

func TestTransactionRepositoryUpdateSuccess(t *testing.T) {
	desc := "new description"
	image := "https://example.com/image.png"

	mock := &mockDynamoDB{
		updateItemFn: func(_ context.Context, table string, key map[string]any, updateExpression string, expressionValues map[string]any, conditionExpression string) error {
			require.Equal(t, "transactions", table)
			assert.Equal(t, "USER#user-1", key["PK"])
			assert.Equal(t, "TX#2026-04-20#tx-1", key["SK"])

			wantCond := "attribute_exists(PK) AND attribute_exists(SK) AND ID = :transactionID"
			require.Equal(t, wantCond, conditionExpression)

			assert.Contains(t, updateExpression, "SET WalletID = :walletID")

			assert.Equal(t, "wallet-2", expressionValues[":walletID"])
			assert.Equal(t, "TX_CATEGORY#cat-2", expressionValues[":gsiCategoryPK"])
			assert.Equal(t, "TX_WALLET#wallet-2", expressionValues[":gsiWalletPK"])
			assert.Equal(t, "tx-1", expressionValues[":transactionID"])

			updatedAt, ok := expressionValues[":updatedAt"].(time.Time)
			require.True(t, ok)
			assert.False(t, updatedAt.IsZero())

			return nil
		},
	}

	repo := NewTransactionRepository(mock, "transactions")
	err := repo.Update(context.Background(), "user-1", "2026-04-20", "tx-1", &model.Transaction{
		WalletID:     "wallet-2",
		WalletName:   "Cash",
		Amount:       999,
		Currency:     "THB",
		CategoryID:   "cat-2",
		CategoryName: "Food",
		Description:  &desc,
		ImageURL:     &image,
	})
	require.NoError(t, err)
}

func TestTransactionRepositoryDeleteSuccess(t *testing.T) {
	mock := &mockDynamoDB{
		deleteItemFn: func(_ context.Context, table string, key map[string]any) error {
			require.Equal(t, "transactions", table)
			assert.Equal(t, "USER#user-1", key["PK"])
			assert.Equal(t, "TX#2026-04-20#tx-1", key["SK"])
			return nil
		},
	}

	repo := NewTransactionRepository(mock, "transactions")
	err := repo.Delete(context.Background(), "user-1", "2026-04-20", "tx-1")
	require.NoError(t, err)
}

func TestTransactionRepositoryUpdateAndDeleteErrorPaths(t *testing.T) {
	repo := NewTransactionRepository(&mockDynamoDB{}, "transactions")

	err := repo.Update(context.Background(), "", "2026-04-20", "tx-1", &model.Transaction{})
	require.EqualError(t, err, "owner ID is required")

	err = repo.Delete(context.Background(), "user-1", "bad-date", "tx-1")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid transaction date format")
}

func TestTransactionRepositoryDBErrorWrapping(t *testing.T) {
	dbErr := errors.New("dynamodb failed")

	mock := &mockDynamoDB{
		updateItemFn: func(_ context.Context, _ string, _ map[string]any, _ string, _ map[string]any, _ string) error {
			return dbErr
		},
		deleteItemFn: func(_ context.Context, _ string, _ map[string]any) error {
			return dbErr
		},
		queryItemsFn: func(_ context.Context, _ string, _ string, _ map[string]any, _ string, _ string, _ any) error {
			return dbErr
		},
	}

	repo := NewTransactionRepository(mock, "transactions")

	err := repo.Update(context.Background(), "user-1", "2026-04-20", "tx-1", &model.Transaction{
		WalletID:   "wallet-1",
		CategoryID: "cat-1",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "update transaction")
	assert.ErrorIs(t, err, dbErr)

	_, err = repo.ListByGSI(context.Background(), "GSI3", "TX_ID", "tx-1", "", nil, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "query transaction by ID")
	assert.ErrorIs(t, err, dbErr)

	err = repo.Delete(context.Background(), "user-1", "2026-04-20", "tx-1")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "delete transaction")
	assert.ErrorIs(t, err, dbErr)

	_, err = repo.ListWithinDateRange(context.Background(), "", time.Now(), time.Now())
	require.EqualError(t, err, "owner ID is required")

	_, err = repo.GetByKey(context.Background(), "")
	require.EqualError(t, err, "ID is required")

	assert.NotEmpty(t, fmt.Sprintf("%v", ErrTransactionNotFound))
}
