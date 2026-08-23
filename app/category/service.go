package category

import (
	"context"
	"errors"
	"strings"

	"github.com/tonytkl/satang/utils"
)

type Service interface {
	CreateCategory(ctx context.Context, ownerID string, name string, strCategoryType string) error
	GetCategoryList(ctx context.Context, ownerID string, nextToken string, limit int32) ([]*Category, string, error)
	GetCategory(ctx context.Context, ownerID string, categoryID string) (*Category, error)
	EditCategory(ctx context.Context, ownerID string, categoryID string, changedFields map[string]any) error
	SetActiveCategory(ctx context.Context, ownerID string, categoryID string, isActive bool) error
}

type service struct {
	repository CategoryRepository
}

func NewService(repository CategoryRepository) Service {
	return &service{
		repository: repository,
	}
}

func (service *service) CreateCategory(ctx context.Context, ownerID string, name string, strCategoryType string) error {
	if ownerID == "" {
		return errors.New("owner ID is required")
	}

	categoryID := utils.GetUUID()
	categoryType, err := getCategoryType(strCategoryType)
	if err != nil {
		return err
	}
	category := NewCategory(
		categoryID,
		ownerID,
		name,
		categoryType,
	)

	if err := service.repository.CreateCategory(ctx, category); err != nil {
		return err
	}

	return nil
}

func (service *service) GetCategoryList(ctx context.Context, ownerID string, nextToken string, limit int32) ([]*Category, string, error) {
	if limit < 0 {
		return nil, "", errors.New("limit must be greater than or equal to 0")
	}
	if limit == 0 {
		limit = utils.DEFAULT_PAGINATION_SIZE
	}
	return service.repository.GetCategoryList(
		ctx,
		ownerID,
		nextToken,
		limit,
	)
}

func (service *service) GetCategory(ctx context.Context, ownerID string, categoryID string) (*Category, error) {
	return service.repository.GetCategory(
		ctx,
		ownerID,
		categoryID,
	)
}

func (service *service) EditCategory(ctx context.Context, ownerID string, categoryID string, changedFields map[string]any) error {
	if _, ok := changedFields["OwnerID"]; ok {
		return errors.New("Owner ID is not updateable")
	}

	if typeValue, ok := changedFields["Type"]; ok {
		strCategoryType, ok := typeValue.(string)
		if !ok {
			return errors.New("Type must be a string")
		}

		categoryType, err := getCategoryType(strCategoryType)
		if err != nil {
			return err
		}
		changedFields["Type"] = categoryType
	}

	return service.repository.EditCategory(ctx, ownerID, categoryID, changedFields)
}

func (service *service) SetActiveCategory(ctx context.Context, ownerID string, categoryID string, isActive bool) error {
	changedFields := make(map[string]any, 1)
	changedFields["IsActive"] = isActive
	return service.EditCategory(ctx, ownerID, categoryID, changedFields)
}

// TODO: Implement delete category. Need to define how to do with existing transaction
// func (service *service) DeleteCategory(ctx context.Context, ownerID string, categoryID string) error {

// }

func getCategoryType(strCategoryType string) (CategoryType, error) {
	switch strings.ToLower(strings.TrimSpace(strCategoryType)) {
	case "expense":
		return CategoryTypeExpense, nil
	case "income":
		return CategoryTypeIncome, nil
	case "transfer":
		return CategoryTypeTransfer, nil
	default:
		return "", errors.New("Invalid category type")
	}
}
