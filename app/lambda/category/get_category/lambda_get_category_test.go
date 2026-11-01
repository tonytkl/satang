package main

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tonytkl/satang/category"
)

type mockGetCategoryService struct {
	category.Service
	getCategoryFunc func(ctx context.Context, ownerID string, categoryID string) (category.Category, error)
}

func (m *mockGetCategoryService) GetCategory(ctx context.Context, ownerID string, categoryID string) (category.Category, error) {
	if m.getCategoryFunc != nil {
		return m.getCategoryFunc(ctx, ownerID, categoryID)
	}
	return category.Category{}, nil
}

var _ category.Service = (*mockGetCategoryService)(nil)

func TestGetCategoryLambdaHandleSuccess(t *testing.T) {
	createdAt := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	updatedAt := time.Date(2026, 1, 3, 4, 5, 6, 0, time.UTC)
	expected := category.CategoryRead{
		ID:           "category-123",
		Name:         "Groceries",
		CategoryType: category.CategoryTypeExpense,
		OwnerID:      "1",
		CreatedAt:    createdAt,
		UpdatedAt:    updatedAt,
	}

	handler := &getCategoryLambda{
		service: &mockGetCategoryService{
			getCategoryFunc: func(ctx context.Context, ownerID string, categoryID string) (category.Category, error) {
				assert.Equal(t, "1", ownerID)
				assert.Equal(t, "category-123", categoryID)
				return category.Category{
					ID:        "category-123",
					OwnerID:   "1",
					Name:      "Groceries",
					Type:      category.CategoryTypeExpense,
					CreatedAt: createdAt,
					UpdatedAt: updatedAt,
				}, nil
			},
		},
	}

	response, err := handler.Handle(context.Background(), events.APIGatewayV2HTTPRequest{
		PathParameters: map[string]string{"category_id": " category-123 "},
	})
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, response.StatusCode)

	var actual category.CategoryRead
	require.NoError(t, json.Unmarshal([]byte(response.Body), &actual))
	assert.Equal(t, expected, actual)
}

func TestGetCategoryLambdaHandleMissingCategoryID(t *testing.T) {
	serviceCalled := false
	handler := &getCategoryLambda{
		service: &mockGetCategoryService{
			getCategoryFunc: func(context.Context, string, string) (category.Category, error) {
				serviceCalled = true
				return category.Category{}, nil
			},
		},
	}

	response, err := handler.Handle(context.Background(), events.APIGatewayV2HTTPRequest{})
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, response.StatusCode)
	assert.False(t, serviceCalled)

	var payload errorResponse
	require.NoError(t, json.Unmarshal([]byte(response.Body), &payload))
	assert.Equal(t, "Category ID is required", payload.Message)
}

func TestGetCategoryLambdaHandleCategoryNotFound(t *testing.T) {
	handler := &getCategoryLambda{
		service: &mockGetCategoryService{
			getCategoryFunc: func(context.Context, string, string) (category.Category, error) {
				return category.Category{}, category.ErrCategoryNotFound
			},
		},
	}

	response, err := handler.Handle(context.Background(), events.APIGatewayV2HTTPRequest{
		PathParameters: map[string]string{"category_id": "category-404"},
	})
	require.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, response.StatusCode)

	var payload errorResponse
	require.NoError(t, json.Unmarshal([]byte(response.Body), &payload))
	assert.Equal(t, "Category not found", payload.Message)
}

func TestGetCategoryLambdaHandleServiceError(t *testing.T) {
	handler := &getCategoryLambda{
		service: &mockGetCategoryService{
			getCategoryFunc: func(context.Context, string, string) (category.Category, error) {
				return category.Category{}, errors.New("database unavailable")
			},
		},
	}

	response, err := handler.Handle(context.Background(), events.APIGatewayV2HTTPRequest{
		PathParameters: map[string]string{"category_id": "category-error"},
	})
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, response.StatusCode)

	var payload errorResponse
	require.NoError(t, json.Unmarshal([]byte(response.Body), &payload))
	assert.Equal(t, "database unavailable", payload.Message)
}

func TestGetCategoryLambdaHandleEmptyCategoryResult(t *testing.T) {
	handler := &getCategoryLambda{
		service: &mockGetCategoryService{
			getCategoryFunc: func(context.Context, string, string) (category.Category, error) {
				return category.Category{}, nil
			},
		},
	}

	response, err := handler.Handle(context.Background(), events.APIGatewayV2HTTPRequest{
		PathParameters: map[string]string{"category_id": "category-empty"},
	})
	require.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, response.StatusCode)

	var payload errorResponse
	require.NoError(t, json.Unmarshal([]byte(response.Body), &payload))
	assert.Equal(t, "Category not found", payload.Message)
}
