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

type mockListCategoriesService struct {
	category.Service
	listCategoriesFunc func(ctx context.Context, ownerID string, nextToken string, limit int32) ([]category.Category, string, error)
}

func (m *mockListCategoriesService) ListCategories(ctx context.Context, ownerID string, nextToken string, limit int32) ([]category.Category, string, error) {
	if m.listCategoriesFunc != nil {
		return m.listCategoriesFunc(ctx, ownerID, nextToken, limit)
	}
	return nil, "", nil
}

var _ category.Service = (*mockListCategoriesService)(nil)

func TestListCategoriesLambdaHandleSuccess(t *testing.T) {
	expectedNextToken := "next-token-1"
	createdAt := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	updatedAt := time.Date(2026, 1, 3, 4, 5, 6, 0, time.UTC)
	handler := &listCategorysLambda{
		service: &mockListCategoriesService{
			listCategoriesFunc: func(ctx context.Context, ownerID string, nextToken string, limit int32) ([]category.Category, string, error) {
				assert.Equal(t, "1", ownerID)
				assert.Equal(t, expectedNextToken, nextToken)
				assert.Equal(t, int32(10), limit)

				return []category.Category{{
					ID:        "category-1",
					Name:      "Groceries",
					Type:      category.CategoryTypeExpense,
					OwnerID:   "1",
					CreatedAt: createdAt,
					UpdatedAt: updatedAt,
				}}, "next-token-2", nil
			},
		},
	}

	response, err := handler.Handle(context.Background(), events.APIGatewayV2HTTPRequest{
		QueryStringParameters: map[string]string{
			"nextToken": expectedNextToken,
			"limit":     "10",
		},
	})

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, response.StatusCode)

	var payload listCategorysResponse
	require.NoError(t, json.Unmarshal([]byte(response.Body), &payload))
	require.Len(t, payload.Categorys, 1)
	assert.Equal(t, category.CategoryRead{
		ID:           "category-1",
		Name:         "Groceries",
		CategoryType: category.CategoryTypeExpense,
		OwnerID:      "1",
		CreatedAt:    createdAt,
		UpdatedAt:    updatedAt,
	}, payload.Categorys[0])
	assert.Equal(t, "next-token-2", payload.NextToken)
}

func TestListCategoriesLambdaHandleInvalidLimit(t *testing.T) {
	handler := &listCategorysLambda{service: &mockListCategoriesService{}}

	response, err := handler.Handle(context.Background(), events.APIGatewayV2HTTPRequest{
		QueryStringParameters: map[string]string{"limit": "abc"},
	})

	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, response.StatusCode)

	var payload errorResponse
	require.NoError(t, json.Unmarshal([]byte(response.Body), &payload))
	assert.Equal(t, "limit must be a valid integer", payload.Message)
}

func TestListCategoriesLambdaHandleNegativeLimit(t *testing.T) {
	serviceCalled := false
	handler := &listCategorysLambda{
		service: &mockListCategoriesService{
			listCategoriesFunc: func(context.Context, string, string, int32) ([]category.Category, string, error) {
				serviceCalled = true
				return nil, "", nil
			},
		},
	}

	response, err := handler.Handle(context.Background(), events.APIGatewayV2HTTPRequest{
		QueryStringParameters: map[string]string{"limit": "-1"},
	})

	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, response.StatusCode)
	assert.False(t, serviceCalled)

	var payload errorResponse
	require.NoError(t, json.Unmarshal([]byte(response.Body), &payload))
	assert.Equal(t, "limit must be greater than or equal to 0", payload.Message)
}

func TestListCategoriesLambdaHandleServiceError(t *testing.T) {
	handler := &listCategorysLambda{
		service: &mockListCategoriesService{
			listCategoriesFunc: func(context.Context, string, string, int32) ([]category.Category, string, error) {
				return nil, "", errors.New("database unavailable")
			},
		},
	}

	response, err := handler.Handle(context.Background(), events.APIGatewayV2HTTPRequest{})

	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, response.StatusCode)

	var payload errorResponse
	require.NoError(t, json.Unmarshal([]byte(response.Body), &payload))
	assert.Equal(t, "database unavailable", payload.Message)
}
