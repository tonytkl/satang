package category

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockCategoryRepository struct {
	createCategoryFn func(ctx context.Context, category *Category) error
	getCategoryListFn func(ctx context.Context, ownerID string, nextToken string, limit int32) ([]*Category, string, error)
	getCategoryFn    func(ctx context.Context, ownerID string, categoryID string) (*Category, error)
	editCategoryFn   func(ctx context.Context, ownerID string, categoryID string, changedFields map[string]any) error
	deleteCategoryFn func(ctx context.Context, ownerID string, categoryID string) error
}

var _ CategoryRepository = (*mockCategoryRepository)(nil)

func (m *mockCategoryRepository) CreateCategory(ctx context.Context, category *Category) error {
	if m.createCategoryFn != nil {
		return m.createCategoryFn(ctx, category)
	}
	return nil
}

func (m *mockCategoryRepository) GetCategoryList(ctx context.Context, ownerID string, nextToken string, limit int32) ([]*Category, string, error) {
	if m.getCategoryListFn != nil {
		return m.getCategoryListFn(ctx, ownerID, nextToken, limit)
	}
	return nil, "", nil
}

func (m *mockCategoryRepository) GetCategory(ctx context.Context, ownerID string, categoryID string) (*Category, error) {
	if m.getCategoryFn != nil {
		return m.getCategoryFn(ctx, ownerID, categoryID)
	}
	return nil, nil
}

func (m *mockCategoryRepository) EditCategory(ctx context.Context, ownerID string, categoryID string, changedFields map[string]any) error {
	if m.editCategoryFn != nil {
		return m.editCategoryFn(ctx, ownerID, categoryID, changedFields)
	}
	return nil
}

func (m *mockCategoryRepository) DeleteCategory(ctx context.Context, ownerID string, categoryID string) error {
	if m.deleteCategoryFn != nil {
		return m.deleteCategoryFn(ctx, ownerID, categoryID)
	}
	return nil
}

func TestCreateCategoryCreatesCategoryWithValidType(t *testing.T) {
	repoCalled := false
	repo := &mockCategoryRepository{
		createCategoryFn: func(ctx context.Context, category *Category) error {
			repoCalled = true
			require.NotEmpty(t, category.ID)
			assert.Equal(t, "user-1", category.OwnerID)
			assert.Equal(t, "Groceries", category.Name)
			assert.Equal(t, CategoryTypeExpense, category.Type)
			assert.True(t, category.IsActive)
			assert.Equal(t, "USER#user-1", category.PK)
			assert.Equal(t, "CATEGORY#"+category.ID, category.SK)
			return nil
		},
	}

	svc := NewService(repo)
	err := svc.CreateCategory(context.Background(), "user-1", "Groceries", "expense")

	require.NoError(t, err)
	assert.True(t, repoCalled)
}

func TestCreateCategoryRejectsInvalidType(t *testing.T) {
	repoCalled := false
	repo := &mockCategoryRepository{
		createCategoryFn: func(ctx context.Context, category *Category) error {
			repoCalled = true
			return nil
		},
	}

	svc := NewService(repo)
	err := svc.CreateCategory(context.Background(), "user-1", "Groceries", "invalid")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "Invalid category type")
	assert.False(t, repoCalled)
}

func TestEditCategoryRejectsProtectedFields(t *testing.T) {
	repoCalled := false
	repo := &mockCategoryRepository{
		editCategoryFn: func(ctx context.Context, ownerID string, categoryID string, changedFields map[string]any) error {
			repoCalled = true
			return nil
		},
	}

	svc := &service{repository: repo}
	err := svc.EditCategory(context.Background(), "user-1", "category-1", map[string]any{"OwnerID": "user-2"})

	require.Error(t, err)
	assert.Equal(t, "Owner ID is not updateable", err.Error())
	assert.False(t, repoCalled)
}

func TestGetCategoryListUsesDefaultLimitWhenZero(t *testing.T) {
	var gotLimit int32
	repo := &mockCategoryRepository{
		getCategoryListFn: func(ctx context.Context, ownerID string, nextToken string, limit int32) ([]*Category, string, error) {
			gotLimit = limit
			return []*Category{}, "", nil
		},
	}

	svc := NewService(repo)
	_, _, err := svc.GetCategoryList(context.Background(), "user-1", "", 0)

	require.NoError(t, err)
	assert.Equal(t, int32(30), gotLimit)
}

func TestEditCategoryDelegatesToRepository(t *testing.T) {
	repo := &mockCategoryRepository{
		editCategoryFn: func(ctx context.Context, ownerID string, categoryID string, changedFields map[string]any) error {
			assert.Equal(t, "user-1", ownerID)
			assert.Equal(t, "category-1", categoryID)
			assert.Equal(t, map[string]any{"Name": "Rental"}, changedFields)
			return nil
		},
	}

	svc := &service{repository: repo}
	err := svc.EditCategory(context.Background(), "user-1", "category-1", map[string]any{"Name": "Rental"})
	require.NoError(t, err)
}

func TestSetActiveCategoryDelegatesToEditCategory(t *testing.T) {
	repoCalled := false
	repo := &mockCategoryRepository{
		editCategoryFn: func(ctx context.Context, ownerID string, categoryID string, changedFields map[string]any) error {
			repoCalled = true
			assert.Equal(t, "user-1", ownerID)
			assert.Equal(t, "category-1", categoryID)
			assert.Equal(t, map[string]any{"IsActive": true}, changedFields)
			return nil
		},
	}

	svc := NewService(repo)
	err := svc.SetActiveCategory(context.Background(), "user-1", "category-1", true)

	require.NoError(t, err)
	assert.True(t, repoCalled)
}

func TestEditCategoryReturnsRepositoryError(t *testing.T) {
	repoErr := errors.New("repository failure")
	repo := &mockCategoryRepository{
		editCategoryFn: func(ctx context.Context, ownerID string, categoryID string, changedFields map[string]any) error {
			return repoErr
		},
	}

	svc := &service{repository: repo}
	err := svc.EditCategory(context.Background(), "user-1", "category-1", map[string]any{"Name": "Updated"})

	require.Error(t, err)
	assert.Equal(t, repoErr, err)
}
