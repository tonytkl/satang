package category

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tonytkl/satang/clients"
)

type mockCategoryDynamoDB struct {
	putItemFn                  func(ctx context.Context, table string, item any) error
	updateItemFn               func(ctx context.Context, table string, key map[string]any, updateExpression string, expressionValues map[string]any, expressionNames map[string]string, conditionExpression string) error
	getItemFn                  func(ctx context.Context, table string, key map[string]any, out any) error
	deleteItemFn               func(ctx context.Context, table string, key map[string]any) error
	queryItemsFn               func(ctx context.Context, table string, keyConditionExpression string, expressionValues map[string]any, indexName string, filterExpression string, out any) error
	queryItemsWithPaginationFn func(ctx context.Context, table string, keyConditionExpression string, expressionValues map[string]any, indexName string, filterExpression string, limit int32, nextToken string, out any) (string, error)
	scanItemsFn                func(ctx context.Context, table string, filterExpression string, expressionValues map[string]any, out any) error
}

var _ clients.DynamoDBClient = (*mockCategoryDynamoDB)(nil)

func (m *mockCategoryDynamoDB) PutItem(ctx context.Context, table string, item any) error {
	if m.putItemFn != nil {
		return m.putItemFn(ctx, table, item)
	}
	return nil
}

func (m *mockCategoryDynamoDB) UpdateItem(ctx context.Context, table string, key map[string]any, updateExpression string, expressionValues map[string]any, expressionNames map[string]string, conditionExpression string) error {
	if m.updateItemFn != nil {
		return m.updateItemFn(ctx, table, key, updateExpression, expressionValues, expressionNames, conditionExpression)
	}
	return nil
}

func (m *mockCategoryDynamoDB) GetItem(ctx context.Context, table string, key map[string]any, out any) error {
	if m.getItemFn != nil {
		return m.getItemFn(ctx, table, key, out)
	}
	return nil
}

func (m *mockCategoryDynamoDB) DeleteItem(ctx context.Context, table string, key map[string]any) error {
	if m.deleteItemFn != nil {
		return m.deleteItemFn(ctx, table, key)
	}
	return nil
}

func (m *mockCategoryDynamoDB) QueryItems(ctx context.Context, table string, keyConditionExpression string, expressionValues map[string]any, indexName string, filterExpression string, out any) error {
	if m.queryItemsFn != nil {
		return m.queryItemsFn(ctx, table, keyConditionExpression, expressionValues, indexName, filterExpression, out)
	}
	return nil
}

func (m *mockCategoryDynamoDB) QueryItemsWithPagination(ctx context.Context, table string, keyConditionExpression string, expressionValues map[string]any, indexName string, filterExpression string, limit int32, nextToken string, out any) (string, error) {
	if m.queryItemsWithPaginationFn != nil {
		return m.queryItemsWithPaginationFn(ctx, table, keyConditionExpression, expressionValues, indexName, filterExpression, limit, nextToken, out)
	}
	return "", nil
}

func (m *mockCategoryDynamoDB) ScanItems(ctx context.Context, table string, filterExpression string, expressionValues map[string]any, out any) error {
	if m.scanItemsFn != nil {
		return m.scanItemsFn(ctx, table, filterExpression, expressionValues, out)
	}
	return nil
}

func TestCategoryRepositoryCreateCategory(t *testing.T) {
	db := &mockCategoryDynamoDB{
		putItemFn: func(_ context.Context, table string, item any) error {
			require.Equal(t, "categories", table)
			category, ok := item.(*Category)
			require.True(t, ok)
			assert.Equal(t, "USER#owner-1", category.PK)
			assert.Equal(t, "CATEGORY#category-1", category.SK)
			return nil
		},
	}

	repo := NewCategoryRepository(db, "categories")
	err := repo.CreateCategory(context.Background(), &Category{ID: "category-1", OwnerID: "owner-1", Name: "Food"})
	require.NoError(t, err)
}

func TestCategoryRepositoryListCategories(t *testing.T) {
	db := &mockCategoryDynamoDB{
		queryItemsWithPaginationFn: func(_ context.Context, table string, keyConditionExpression string, expressionValues map[string]any, _, _ string, limit int32, nextToken string, out any) (string, error) {
			require.Equal(t, "categories", table)
			assert.Equal(t, "PK = :ownerPK AND begins_with(SK, :modelPrefix)", keyConditionExpression)
			assert.Equal(t, "USER#owner-1", expressionValues[":ownerPK"])
			assert.Equal(t, "CATEGORY#", expressionValues[":modelPrefix"])
			assert.Equal(t, int32(5), limit)
			assert.Equal(t, "tok-1", nextToken)

			dst, ok := out.(*[]*Category)
			require.True(t, ok)
			*dst = []*Category{{ID: "category-1"}}
			return "tok-2", nil
		},
	}

	repo := NewCategoryRepository(db, "categories")
	got, nextToken, err := repo.ListCategories(context.Background(), "owner-1", "tok-1", 5)
	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, "category-1", got[0].ID)
	assert.Equal(t, "tok-2", nextToken)
}

func TestCategoryRepositoryGetCategory(t *testing.T) {
	db := &mockCategoryDynamoDB{
		getItemFn: func(_ context.Context, table string, key map[string]any, out any) error {
			require.Equal(t, "categories", table)
			assert.Equal(t, "USER#owner-1", key["PK"])
			assert.Equal(t, "CATEGORY#category-1", key["SK"])

			category, ok := out.(*Category)
			require.True(t, ok)
			category.ID = "category-1"
			category.OwnerID = "owner-1"
			category.Name = "Food"
			return nil
		},
	}

	repo := NewCategoryRepository(db, "categories")
	got, err := repo.GetCategory(context.Background(), "owner-1", "category-1")
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "category-1", got.ID)
}

func TestCategoryRepositoryEditCategory(t *testing.T) {
	db := &mockCategoryDynamoDB{
		updateItemFn: func(_ context.Context, table string, key map[string]any, updateExpression string, expressionValues map[string]any, _ map[string]string, conditionExpression string) error {
			require.Equal(t, "categories", table)
			assert.Equal(t, "USER#owner-1", key["PK"])
			assert.Equal(t, "CATEGORY#category-1", key["SK"])
			assert.Contains(t, updateExpression, "SET")
			assert.Equal(t, "category-1", expressionValues[":id"])
			assert.Equal(t, "attribute_exists(PK) AND attribute_exists(SK) AND ID = :id", conditionExpression)
			return nil
		},
	}

	repo := NewCategoryRepository(db, "categories")
	err := repo.EditCategory(context.Background(), "owner-1", "category-1", map[string]any{"Name": "Travel"})
	require.NoError(t, err)
}

func TestCategoryRepositoryDeleteCategory(t *testing.T) {
	db := &mockCategoryDynamoDB{
		deleteItemFn: func(_ context.Context, table string, key map[string]any) error {
			require.Equal(t, "categories", table)
			assert.Equal(t, "USER#owner-1", key["PK"])
			assert.Equal(t, "CATEGORY#category-1", key["SK"])
			return nil
		},
	}

	repo := NewCategoryRepository(db, "categories")
	err := repo.DeleteCategory(context.Background(), "owner-1", "category-1")
	require.NoError(t, err)
}
