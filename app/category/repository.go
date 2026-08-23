package category

import (
	"context"

	"github.com/tonytkl/satang/clients"
	"github.com/tonytkl/satang/repository"
)

type CategoryRepository interface {
	CreateCategory(ctx context.Context, category *Category) error
	GetCategoryList(ctx context.Context, ownerID string, nextToken string, limit int32) ([]*Category, string, error)
	GetCategory(ctx context.Context, ownerID string, categoryID string) (*Category, error)
	EditCategory(ctx context.Context, ownerID string, categoryID string, changedFields map[string]any) error
	DeleteCategory(ctx context.Context, ownerID string, categoryID string) error
}

type categoryRepository struct {
	db             clients.DynamoDBClient
	tableName      string
	baseRepository repository.BaseRepository[*Category]
}

func NewCategoryRepository(db clients.DynamoDBClient, tableName string) CategoryRepository {
	return &categoryRepository{
		db:             db,
		tableName:      tableName,
		baseRepository: repository.NewBaseRepository(db, tableName, "CATEGORY", func() *Category { return &Category{} }),
	}
}

func (categoryRepository *categoryRepository) CreateCategory(ctx context.Context, category *Category) error {
	return categoryRepository.baseRepository.Save(ctx, category)
}

func (categoryRepository *categoryRepository) GetCategoryList(ctx context.Context, ownerID string, nextToken string, limit int32) ([]*Category, string, error) {
	return categoryRepository.baseRepository.List(ctx, ownerID, nextToken, limit)
}

func (categoryRepository *categoryRepository) GetCategory(ctx context.Context, ownerID string, categoryID string) (*Category, error) {
	return categoryRepository.baseRepository.Get(ctx, ownerID, categoryID)
}

func (categoryRepository *categoryRepository) EditCategory(ctx context.Context, ownerID string, categoryID string, changedFields map[string]any) error {
	return categoryRepository.baseRepository.Update(ctx, ownerID, categoryID, changedFields)
}

func (categoryRepository *categoryRepository) DeleteCategory(ctx context.Context, ownerID string, categoryID string) error {
	return categoryRepository.baseRepository.Delete(ctx, ownerID, categoryID)
}
