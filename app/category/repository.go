package category

import (
	"context"
	"errors"

	"github.com/tonytkl/satang/clients"
	"github.com/tonytkl/satang/repository"
)

var ErrCategoryNotFound = errors.New("Category not found")

type Repository interface {
	CreateCategory(ctx context.Context, category Category) error
	ListCategories(ctx context.Context, ownerID string, nextToken string, limit int32) ([]Category, string, error)
	GetCategory(ctx context.Context, ownerID string, categoryID string) (Category, error)
	EditCategory(ctx context.Context, ownerID string, categoryID string, changedFields map[string]any) error
	DeleteCategory(ctx context.Context, ownerID string, categoryID string) error
}

type categoryRepository struct {
	db             clients.DynamoDBClient
	tableName      string
	baseRepository repository.BaseRepository[*Category]
}

func NewRepository(db clients.DynamoDBClient, tableName string) Repository {
	return &categoryRepository{
		db:             db,
		tableName:      tableName,
		baseRepository: repository.NewBaseRepository(db, tableName, "CATEGORY", func() *Category { return &Category{} }),
	}
}

func (categoryRepository *categoryRepository) CreateCategory(ctx context.Context, category Category) error {
	return categoryRepository.baseRepository.Save(ctx, &category)
}

func (categoryRepository *categoryRepository) ListCategories(ctx context.Context, ownerID string, nextToken string, limit int32) ([]Category, string, error) {
	items, encodedNextToken, err := categoryRepository.baseRepository.List(ctx, ownerID, nextToken, limit)
	if err != nil {
		return nil, "", err
	}

	categories := make([]Category, 0, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}
		categories = append(categories, *item)
	}

	return categories, encodedNextToken, nil
}

func (categoryRepository *categoryRepository) GetCategory(ctx context.Context, ownerID string, categoryID string) (Category, error) {
	category, err := categoryRepository.baseRepository.Get(ctx, ownerID, categoryID)
	if err != nil {
		if errors.Is(err, clients.ErrItemNotFound) {
			return Category{}, ErrCategoryNotFound
		}
		return Category{}, err
	}
	if category == nil {
		return Category{}, ErrCategoryNotFound
	}
	return *category, nil
}

func (categoryRepository *categoryRepository) EditCategory(ctx context.Context, ownerID string, categoryID string, changedFields map[string]any) error {
	_, err := categoryRepository.baseRepository.Update(ctx, ownerID, categoryID, changedFields)
	return err
}

func (categoryRepository *categoryRepository) DeleteCategory(ctx context.Context, ownerID string, categoryID string) error {
	return categoryRepository.baseRepository.Delete(ctx, ownerID, categoryID)
}
