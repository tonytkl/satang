package repository

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tonytkl/satang/clients"
	"github.com/tonytkl/satang/wallet"
)

// mockDynamoDB is a test double for clients.DynamoDBClient.
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
	return "", nil
}

func (m *mockDynamoDB) ScanItems(ctx context.Context, table string, filterExpression string, expressionValues map[string]any, out any) error {
	if m.scanItemsFn != nil {
		return m.scanItemsFn(ctx, table, filterExpression, expressionValues, out)
	}
	return nil
}

// newWalletRepo creates a BaseRepository[*wallet.Wallet] backed by the given mock.
func newWalletRepo(db clients.DynamoDBClient) BaseRepository[*wallet.Wallet] {
	return NewBaseRepository(db, "wallets", "WALLET", func() *wallet.Wallet { return &wallet.Wallet{} })
}

// ---------------------------------------------------------------------------
// Save
// ---------------------------------------------------------------------------

func TestBaseRepositorySave_SetsKeysAndTimestamps(t *testing.T) {
	db := &mockDynamoDB{
		putItemFn: func(_ context.Context, table string, item any) error {
			require.Equal(t, "wallets", table)

			w, ok := item.(*wallet.Wallet)
			require.True(t, ok)

			assert.Equal(t, "USER#owner-1", w.PK)
			assert.Equal(t, "WALLET#wallet-1", w.SK)
			assert.False(t, w.CreatedAt.IsZero(), "CreatedAt should be set")
			assert.False(t, w.UpdatedAt.IsZero(), "UpdatedAt should be set")
			return nil
		},
	}

	repo := newWalletRepo(db)
	w := &wallet.Wallet{ID: "wallet-1", OwnerID: "owner-1", Name: "Cash", Currency: "THB"}

	err := repo.Save(context.Background(), w)
	require.NoError(t, err)
}

func TestBaseRepositorySave_PreservesExistingTimestamps(t *testing.T) {
	createdAt := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	updatedAt := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)

	db := &mockDynamoDB{
		putItemFn: func(_ context.Context, _ string, item any) error {
			w := item.(*wallet.Wallet)
			assert.Equal(t, createdAt, w.CreatedAt)
			assert.Equal(t, updatedAt, w.UpdatedAt)
			return nil
		},
	}

	repo := newWalletRepo(db)
	w := &wallet.Wallet{ID: "wallet-1", OwnerID: "owner-1", CreatedAt: createdAt, UpdatedAt: updatedAt}

	err := repo.Save(context.Background(), w)
	require.NoError(t, err)
}

func TestBaseRepositorySave_PropagatesDynamoDBError(t *testing.T) {
	db := &mockDynamoDB{
		putItemFn: func(_ context.Context, _ string, _ any) error {
			return errors.New("dynamodb unavailable")
		},
	}

	repo := newWalletRepo(db)
	w := &wallet.Wallet{ID: "wallet-1", OwnerID: "owner-1"}

	err := repo.Save(context.Background(), w)
	require.EqualError(t, err, "dynamodb unavailable")
}

// ---------------------------------------------------------------------------
// Get
// ---------------------------------------------------------------------------

func TestBaseRepositoryGet_Success(t *testing.T) {
	db := &mockDynamoDB{
		getItemFn: func(_ context.Context, table string, key map[string]any, out any) error {
			require.Equal(t, "wallets", table)
			assert.Equal(t, "USER#owner-1", key["PK"])
			assert.Equal(t, "WALLET#wallet-1", key["SK"])

			w, ok := out.(wallet.Wallet)
			require.True(t, ok)
			w.ID = "wallet-1"
			w.OwnerID = "owner-1"
			w.Name = "Cash"
			return nil
		},
	}

	repo := newWalletRepo(db)

	got, err := repo.Get(context.Background(), "owner-1", "wallet-1")
	require.NoError(t, err)
	require.NotNil(t, got)
}

func TestBaseRepositoryGet_EmptyOwnerIDReturnsError(t *testing.T) {
	repo := newWalletRepo(&mockDynamoDB{})

	_, err := repo.Get(context.Background(), "", "wallet-1")
	require.EqualError(t, err, "owner ID is required")
}

func TestBaseRepositoryGet_EmptyItemIDReturnsError(t *testing.T) {
	repo := newWalletRepo(&mockDynamoDB{})

	_, err := repo.Get(context.Background(), "owner-1", "")
	require.EqualError(t, err, "item ID is required")
}

func TestBaseRepositoryGet_PropagatesDynamoDBError(t *testing.T) {
	db := &mockDynamoDB{
		getItemFn: func(_ context.Context, _ string, _ map[string]any, _ any) error {
			return errors.New("connection refused")
		},
	}

	repo := newWalletRepo(db)

	_, err := repo.Get(context.Background(), "owner-1", "wallet-1")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Error on Get")
	assert.Contains(t, err.Error(), "connection refused")
}

// ---------------------------------------------------------------------------
// List
// ---------------------------------------------------------------------------

func TestBaseRepositoryList_Success(t *testing.T) {
	db := &mockDynamoDB{
		queryItemsWithPaginationFn: func(_ context.Context, table string, keyConditionExpression string, expressionValues map[string]any, _, _ string, limit int32, nextToken string, out any) (string, error) {
			require.Equal(t, "wallets", table)
			assert.Equal(t, "PK = :ownerPK AND begins_with(SK, :modelPrefix)", keyConditionExpression)
			assert.Equal(t, "USER#owner-1", expressionValues[":ownerPK"])
			assert.Equal(t, "WALLET#", expressionValues[":modelPrefix"])
			assert.Equal(t, int32(10), limit)
			assert.Equal(t, "", nextToken)

			dest, ok := out.(*[]*wallet.Wallet)
			require.True(t, ok)
			*dest = []*wallet.Wallet{
				{ID: "wallet-1"},
				{ID: "wallet-2"},
			}
			return "next-page-token", nil
		},
	}

	repo := newWalletRepo(db)

	items, token, err := repo.List(context.Background(), "owner-1", "", 10)
	require.NoError(t, err)
	assert.Equal(t, "next-page-token", token)
	require.Len(t, items, 2)
	assert.Equal(t, "wallet-1", items[0].ID)
	assert.Equal(t, "wallet-2", items[1].ID)
}

func TestBaseRepositoryList_EmptyOwnerIDReturnsError(t *testing.T) {
	repo := newWalletRepo(&mockDynamoDB{})

	_, _, err := repo.List(context.Background(), "", "", 10)
	require.EqualError(t, err, "owner ID is required")
}

func TestBaseRepositoryList_PropagatesDynamoDBError(t *testing.T) {
	db := &mockDynamoDB{
		queryItemsWithPaginationFn: func(_ context.Context, _ string, _ string, _ map[string]any, _ string, _ string, _ int32, _ string, _ any) (string, error) {
			return "", errors.New("query failed")
		},
	}

	repo := newWalletRepo(db)

	_, _, err := repo.List(context.Background(), "owner-1", "", 10)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Errors on List")
	assert.Contains(t, err.Error(), "query failed")
}

// ---------------------------------------------------------------------------
// Update
// ---------------------------------------------------------------------------

func TestBaseRepositoryUpdate_Success(t *testing.T) {
	db := &mockDynamoDB{
		updateItemFn: func(_ context.Context, table string, key map[string]any, updateExpression string, expressionValues map[string]any, expressionNames map[string]string, conditionExpression string) error {
			require.Equal(t, "wallets", table)
			assert.Equal(t, "USER#owner-1", key["PK"])
			assert.Equal(t, "WALLET#wallet-1", key["SK"])
			assert.Contains(t, updateExpression, "SET")
			assert.Equal(t, "attribute_exists(PK) AND attribute_exists(SK) AND ID = :id", conditionExpression)
			assert.Equal(t, "wallet-1", expressionValues[":id"])
			return nil
		},
	}

	repo := newWalletRepo(db)
	w := &wallet.Wallet{
		ID:      "wallet-1",
		OwnerID: "owner-1",
		Name:    "Updated Name",
	}

	err := repo.Update(context.Background(), "owner-1", "wallet-1", w)
	require.NoError(t, err)
}

func TestBaseRepositoryUpdate_EmptyOwnerIDReturnsError(t *testing.T) {
	repo := newWalletRepo(&mockDynamoDB{})
	w := &wallet.Wallet{ID: "wallet-1", OwnerID: "owner-1"}

	err := repo.Update(context.Background(), "", "wallet-1", w)
	require.EqualError(t, err, "owner ID is required")
}

func TestBaseRepositoryUpdate_EmptyItemIDReturnsError(t *testing.T) {
	repo := newWalletRepo(&mockDynamoDB{})
	w := &wallet.Wallet{ID: "wallet-1", OwnerID: "owner-1"}

	err := repo.Update(context.Background(), "owner-1", "", w)
	require.EqualError(t, err, "item ID is required")
}

func TestBaseRepositoryUpdate_NilItemReturnsError(t *testing.T) {
	repo := newWalletRepo(&mockDynamoDB{})

	err := repo.Update(context.Background(), "owner-1", "wallet-1", nil)
	require.EqualError(t, err, "Update payload is required")
}

func TestBaseRepositoryUpdate_PropagatesDynamoDBError(t *testing.T) {
	db := &mockDynamoDB{
		updateItemFn: func(_ context.Context, _ string, _ map[string]any, _ string, _ map[string]any, _ map[string]string, _ string) error {
			return errors.New("conditional check failed")
		},
	}

	repo := newWalletRepo(db)
	w := &wallet.Wallet{ID: "wallet-1", OwnerID: "owner-1", Name: "Cash"}

	err := repo.Update(context.Background(), "owner-1", "wallet-1", w)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "update item")
	assert.Contains(t, err.Error(), "conditional check failed")
}

// ---------------------------------------------------------------------------
// Delete
// ---------------------------------------------------------------------------

func TestBaseRepositoryDelete_Success(t *testing.T) {
	db := &mockDynamoDB{
		deleteItemFn: func(_ context.Context, table string, key map[string]any) error {
			require.Equal(t, "wallets", table)
			assert.Equal(t, "USER#owner-1", key["PK"])
			assert.Equal(t, "WALLET#wallet-1", key["SK"])
			return nil
		},
	}

	repo := newWalletRepo(db)

	err := repo.Delete(context.Background(), "owner-1", "wallet-1")
	require.NoError(t, err)
}

func TestBaseRepositoryDelete_EmptyOwnerIDReturnsError(t *testing.T) {
	repo := newWalletRepo(&mockDynamoDB{})

	err := repo.Delete(context.Background(), "", "wallet-1")
	require.EqualError(t, err, "owner ID is required")
}

func TestBaseRepositoryDelete_EmptyItemIDReturnsError(t *testing.T) {
	repo := newWalletRepo(&mockDynamoDB{})

	err := repo.Delete(context.Background(), "owner-1", "")
	require.EqualError(t, err, "item ID is required")
}

func TestBaseRepositoryDelete_PropagatesDynamoDBError(t *testing.T) {
	db := &mockDynamoDB{
		deleteItemFn: func(_ context.Context, _ string, _ map[string]any) error {
			return errors.New("delete failed")
		},
	}

	repo := newWalletRepo(db)

	err := repo.Delete(context.Background(), "owner-1", "wallet-1")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Error on Delete")
	assert.Contains(t, err.Error(), "delete failed")
}
