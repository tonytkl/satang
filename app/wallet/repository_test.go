package wallet

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tonytkl/satang/clients"
)

type mockWalletDynamoDB struct {
	putItemFn                  func(ctx context.Context, table string, item any) error
	updateItemFn               func(ctx context.Context, table string, key map[string]any, updateExpression string, expressionValues map[string]any, expressionNames map[string]string, conditionExpression string) error
	getItemFn                  func(ctx context.Context, table string, key map[string]any, out any) error
	deleteItemFn               func(ctx context.Context, table string, key map[string]any) error
	queryItemsFn               func(ctx context.Context, table string, keyConditionExpression string, expressionValues map[string]any, indexName string, filterExpression string, out any) error
	queryItemsWithPaginationFn func(ctx context.Context, table string, keyConditionExpression string, expressionValues map[string]any, indexName string, filterExpression string, limit int32, nextToken string, out any) (string, error)
	scanItemsFn                func(ctx context.Context, table string, filterExpression string, expressionValues map[string]any, out any) error
}

var _ clients.DynamoDBClient = (*mockWalletDynamoDB)(nil)

func (m *mockWalletDynamoDB) PutItem(ctx context.Context, table string, item any) error {
	if m.putItemFn != nil {
		return m.putItemFn(ctx, table, item)
	}
	return nil
}

func (m *mockWalletDynamoDB) UpdateItem(ctx context.Context, table string, key map[string]any, updateExpression string, expressionValues map[string]any, expressionNames map[string]string, conditionExpression string) error {
	if m.updateItemFn != nil {
		return m.updateItemFn(ctx, table, key, updateExpression, expressionValues, expressionNames, conditionExpression)
	}
	return nil
}

func (m *mockWalletDynamoDB) GetItem(ctx context.Context, table string, key map[string]any, out any) error {
	if m.getItemFn != nil {
		return m.getItemFn(ctx, table, key, out)
	}
	return nil
}

func (m *mockWalletDynamoDB) DeleteItem(ctx context.Context, table string, key map[string]any) error {
	if m.deleteItemFn != nil {
		return m.deleteItemFn(ctx, table, key)
	}
	return nil
}

func (m *mockWalletDynamoDB) QueryItems(ctx context.Context, table string, keyConditionExpression string, expressionValues map[string]any, indexName string, filterExpression string, out any) error {
	if m.queryItemsFn != nil {
		return m.queryItemsFn(ctx, table, keyConditionExpression, expressionValues, indexName, filterExpression, out)
	}
	return nil
}

func (m *mockWalletDynamoDB) QueryItemsWithPagination(ctx context.Context, table string, keyConditionExpression string, expressionValues map[string]any, indexName string, filterExpression string, limit int32, nextToken string, out any) (string, error) {
	if m.queryItemsWithPaginationFn != nil {
		return m.queryItemsWithPaginationFn(ctx, table, keyConditionExpression, expressionValues, indexName, filterExpression, limit, nextToken, out)
	}
	return "", nil
}

func (m *mockWalletDynamoDB) ScanItems(ctx context.Context, table string, filterExpression string, expressionValues map[string]any, out any) error {
	if m.scanItemsFn != nil {
		return m.scanItemsFn(ctx, table, filterExpression, expressionValues, out)
	}
	return nil
}

func TestWalletRepositoryCreateWallet(t *testing.T) {
	db := &mockWalletDynamoDB{
		putItemFn: func(_ context.Context, table string, item any) error {
			require.Equal(t, "wallets", table)
			wallet, ok := item.(*Wallet)
			require.True(t, ok)
			assert.Equal(t, "USER#owner-1", wallet.PK)
			assert.Equal(t, "WALLET#wallet-1", wallet.SK)
			return nil
		},
	}

	repo := NewWalletRepository(db, "wallets")
	err := repo.CreateWallet(context.Background(), &Wallet{ID: "wallet-1", OwnerID: "owner-1", Name: "Cash"})
	require.NoError(t, err)
}

func TestWalletRepositoryListWallets(t *testing.T) {
	db := &mockWalletDynamoDB{
		queryItemsWithPaginationFn: func(_ context.Context, table string, keyConditionExpression string, expressionValues map[string]any, _, _ string, limit int32, nextToken string, out any) (string, error) {
			require.Equal(t, "wallets", table)
			assert.Equal(t, "PK = :ownerPK AND begins_with(SK, :modelPrefix)", keyConditionExpression)
			assert.Equal(t, "USER#owner-1", expressionValues[":ownerPK"])
			assert.Equal(t, "WALLET#", expressionValues[":modelPrefix"])
			assert.Equal(t, int32(10), limit)
			assert.Equal(t, "tok-1", nextToken)

			dst, ok := out.(*[]*Wallet)
			require.True(t, ok)
			*dst = []*Wallet{{ID: "wallet-1"}}
			return "tok-2", nil
		},
	}

	repo := NewWalletRepository(db, "wallets")
	got, nextToken, err := repo.ListWallets(context.Background(), "owner-1", "tok-1", 10)
	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, "wallet-1", got[0].ID)
	assert.Equal(t, "tok-2", nextToken)
}

func TestWalletRepositoryGetWallet(t *testing.T) {
	db := &mockWalletDynamoDB{
		getItemFn: func(_ context.Context, table string, key map[string]any, out any) error {
			require.Equal(t, "wallets", table)
			assert.Equal(t, "USER#owner-1", key["PK"])
			assert.Equal(t, "WALLET#wallet-1", key["SK"])

			wallet, ok := out.(*Wallet)
			require.True(t, ok)
			wallet.ID = "wallet-1"
			wallet.OwnerID = "owner-1"
			wallet.Name = "Cash"
			return nil
		},
	}

	repo := NewWalletRepository(db, "wallets")
	got, err := repo.GetWallet(context.Background(), "owner-1", "wallet-1")
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "wallet-1", got.ID)
}

func TestWalletRepositoryEditWallet(t *testing.T) {
	db := &mockWalletDynamoDB{
		updateItemFn: func(_ context.Context, table string, key map[string]any, updateExpression string, expressionValues map[string]any, _ map[string]string, conditionExpression string) error {
			require.Equal(t, "wallets", table)
			assert.Equal(t, "USER#owner-1", key["PK"])
			assert.Equal(t, "WALLET#wallet-1", key["SK"])
			assert.Contains(t, updateExpression, "SET")
			assert.Equal(t, "wallet-1", expressionValues[":id"])
			assert.Equal(t, "attribute_exists(PK) AND attribute_exists(SK) AND ID = :id", conditionExpression)
			return nil
		},
	}

	repo := NewWalletRepository(db, "wallets")
	err := repo.EditWallet(context.Background(), "owner-1", "wallet-1", map[string]any{"Name": "Updated"})
	require.NoError(t, err)
}

func TestWalletRepositoryDeleteWallet(t *testing.T) {
	db := &mockWalletDynamoDB{
		deleteItemFn: func(_ context.Context, table string, key map[string]any) error {
			require.Equal(t, "wallets", table)
			assert.Equal(t, "USER#owner-1", key["PK"])
			assert.Equal(t, "WALLET#wallet-1", key["SK"])
			return nil
		},
	}

	repo := NewWalletRepository(db, "wallets")
	err := repo.DeleteWallet(context.Background(), "owner-1", "wallet-1")
	require.NoError(t, err)
}