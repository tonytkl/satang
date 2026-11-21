package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tonytkl/satang/category"
	"github.com/tonytkl/satang/clients"
)

type mockEditCategoryService struct {
	category.Service
	editCategoryFunc func(ctx context.Context, ownerID string, categoryID string, changedFields map[string]any) (category.Category, error)
}

func (service *mockEditCategoryService) EditCategory(ctx context.Context, ownerID string, categoryID string, changedFields map[string]any) (category.Category, error) {
	if service.editCategoryFunc != nil {
		return service.editCategoryFunc(ctx, ownerID, categoryID, changedFields)
	}
	return category.Category{}, nil
}

var _ category.Service = (*mockEditCategoryService)(nil)

func TestHandle_ValidPayload(t *testing.T) {
	createdAt := time.Date(2026, time.January, 2, 3, 4, 5, 0, time.UTC)
	updatedAt := createdAt.Add(time.Minute)
	handler := &editCategoryLambda{service: &mockEditCategoryService{
		editCategoryFunc: func(_ context.Context, ownerID string, categoryID string, changedFields map[string]any) (category.Category, error) {
			assert.Equal(t, "1", ownerID)
			assert.Equal(t, "category-1", categoryID)
			assert.Equal(t, map[string]any{"Name": "Travel", "Type": "income"}, changedFields)
			return category.Category{
				ID:        categoryID,
				OwnerID:   ownerID,
				Name:      "Travel",
				Type:      category.CategoryTypeIncome,
				CreatedAt: createdAt,
				UpdatedAt: updatedAt,
			}, nil
		},
	}}

	response, err := handler.Handle(context.Background(), events.APIGatewayV2HTTPRequest{
		PathParameters: map[string]string{"category_id": " category-1 "},
		Body:           `{"name":"Travel","type":"income"}`,
	})

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, response.StatusCode)
	assert.Equal(t, "application/json", response.Headers["Content-Type"])

	var result category.CategoryRead
	require.NoError(t, json.Unmarshal([]byte(response.Body), &result))
	assert.Equal(t, category.CategoryRead{
		ID:           "category-1",
		Name:         "Travel",
		CategoryType: category.CategoryTypeIncome,
		OwnerID:      "1",
		CreatedAt:    createdAt,
		UpdatedAt:    updatedAt,
	}, result)
}

func TestHandle_MissingCategoryID(t *testing.T) {
	serviceCalled := false
	handler := &editCategoryLambda{service: &mockEditCategoryService{
		editCategoryFunc: func(context.Context, string, string, map[string]any) (category.Category, error) {
			serviceCalled = true
			return category.Category{}, nil
		},
	}}

	response, err := handler.Handle(context.Background(), events.APIGatewayV2HTTPRequest{
		PathParameters: map[string]string{"category_id": "   "},
		Body:           `{}`,
	})

	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, response.StatusCode)
	assert.False(t, serviceCalled)
	assertEditCategoryErrorMessage(t, response, "Category ID is required")
}

func TestHandle_InvalidJSONPayload(t *testing.T) {
	handler := &editCategoryLambda{service: &mockEditCategoryService{}}
	response, err := handler.Handle(context.Background(), events.APIGatewayV2HTTPRequest{
		PathParameters: map[string]string{"category_id": "category-1"},
		Body:           `{invalid`,
	})

	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, response.StatusCode)
	assertEditCategoryErrorMessage(t, response, "Invalid JSON payload")
}

func TestHandle_PartialPayload(t *testing.T) {
	handler := &editCategoryLambda{service: &mockEditCategoryService{
		editCategoryFunc: func(_ context.Context, _ string, _ string, changedFields map[string]any) (category.Category, error) {
			assert.Equal(t, map[string]any{"Name": "Travel"}, changedFields)
			return category.Category{ID: "category-1", Name: "Travel"}, nil
		},
	}}

	response, err := handler.Handle(context.Background(), events.APIGatewayV2HTTPRequest{
		PathParameters: map[string]string{"category_id": "category-1"},
		Body:           `{"name":"Travel"}`,
	})

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, response.StatusCode)
}

func TestHandle_ServiceError(t *testing.T) {
	handler := &editCategoryLambda{service: &mockEditCategoryService{
		editCategoryFunc: func(context.Context, string, string, map[string]any) (category.Category, error) {
			return category.Category{}, assert.AnError
		},
	}}
	response, err := handler.Handle(context.Background(), events.APIGatewayV2HTTPRequest{
		PathParameters: map[string]string{"category_id": "category-1"},
		Body:           `{"name":"Travel"}`,
	})

	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, response.StatusCode)
	assertEditCategoryErrorMessage(t, response, assert.AnError.Error())
}

func TestHandle_CategoryNotFound(t *testing.T) {
	handler := &editCategoryLambda{service: &mockEditCategoryService{
		editCategoryFunc: func(context.Context, string, string, map[string]any) (category.Category, error) {
			return category.Category{}, category.ErrCategoryNotFound
		},
	}}
	response, err := handler.Handle(context.Background(), events.APIGatewayV2HTTPRequest{
		PathParameters: map[string]string{"category_id": "category-404"},
		Body:           `{"name":"Travel"}`,
	})

	require.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, response.StatusCode)
	assertEditCategoryErrorMessage(t, response, "Category not found")
}

func TestHandle_ServiceDeadline(t *testing.T) {
	handler := &editCategoryLambda{service: &mockEditCategoryService{
		editCategoryFunc: func(context.Context, string, string, map[string]any) (category.Category, error) {
			return category.Category{}, context.DeadlineExceeded
		},
	}}
	response, err := handler.Handle(context.Background(), events.APIGatewayV2HTTPRequest{
		PathParameters: map[string]string{"category_id": "category-1"}, Body: `{"name":"Travel"}`,
	})
	require.NoError(t, err)
	assert.Equal(t, http.StatusGatewayTimeout, response.StatusCode)
}

func TestHandle_InvalidCategoryType(t *testing.T) {
	handler := &editCategoryLambda{service: category.NewService(nil)}
	response, err := handler.Handle(context.Background(), events.APIGatewayV2HTTPRequest{
		PathParameters: map[string]string{"category_id": "id"}, Body: `{"type":"invalid"}`,
	})
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, response.StatusCode)
}

func TestHandle_EmptyUpdate(t *testing.T) {
	handler := &editCategoryLambda{service: category.NewService(category.NewRepository(nil, "categories"))}
	response, err := handler.Handle(context.Background(), events.APIGatewayV2HTTPRequest{
		PathParameters: map[string]string{"category_id": "id"}, Body: `{}`,
	})
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, response.StatusCode)
	assertEditCategoryErrorMessage(t, response, "Update payload is required")
}

func TestHandle_DynamoDBUpdateErrors(t *testing.T) {
	for _, tt := range []struct {
		name string
		code string
		want int
	}{
		{"throttled", "ProvisionedThroughputExceededException", http.StatusTooManyRequests},
		{"missing category", "ConditionalCheckFailedException", http.StatusNotFound},
	} {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/x-amz-json-1.0")
				w.WriteHeader(http.StatusBadRequest)
				_ = json.NewEncoder(w).Encode(map[string]string{"__type": tt.code, "message": "update failed"})
			}))
			defer server.Close()
			t.Setenv("ENVIRONMENT", "local")
			t.Setenv("AWS_DYNAMODB_ENDPOINT", server.URL)
			db := clients.NewDynamoDBClientWithConfig(aws.Config{
				Region:      "us-east-1",
				HTTPClient:  server.Client(),
				Credentials: aws.NewCredentialsCache(credentials.NewStaticCredentialsProvider("test", "test", "test")),
				Retryer:     func() aws.Retryer { return aws.NopRetryer{} },
			})
			handler := &editCategoryLambda{service: category.NewService(category.NewRepository(db, "categories"))}
			response, err := handler.Handle(context.Background(), events.APIGatewayV2HTTPRequest{
				PathParameters: map[string]string{"category_id": "id"}, Body: `{"name":"Travel"}`,
			})
			require.NoError(t, err)
			assert.Equal(t, tt.want, response.StatusCode)
		})
	}
}

func assertEditCategoryErrorMessage(t *testing.T, response events.APIGatewayV2HTTPResponse, expected string) {
	t.Helper()
	var result errorResponse
	require.NoError(t, json.Unmarshal([]byte(response.Body), &result))
	assert.Equal(t, expected, result.Message)
}
