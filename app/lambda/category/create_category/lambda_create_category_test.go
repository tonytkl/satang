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

type mockCategoryService struct {
	category.Service
	createCategoryFunc func(ctx context.Context, ownerID string, name string, categoryType string) (category.Category, error)
}

func (m *mockCategoryService) CreateCategory(ctx context.Context, ownerID string, name string, categoryType string) (category.Category, error) {
	if m.createCategoryFunc != nil {
		return m.createCategoryFunc(ctx, ownerID, name, categoryType)
	}
	return category.Category{}, nil
}

var _ category.Service = (*mockCategoryService)(nil)

func TestHandle_ValidPayload(t *testing.T) {
	createdAt := time.Date(2026, time.January, 2, 3, 4, 5, 0, time.UTC)
	updatedAt := createdAt.Add(time.Minute)
	service := &mockCategoryService{
		createCategoryFunc: func(_ context.Context, ownerID string, name string, categoryType string) (category.Category, error) {
			assert.Equal(t, "1", ownerID)
			assert.Equal(t, "Groceries", name)
			assert.Equal(t, "expense", categoryType)
			return category.Category{
				ID:        "category-1",
				OwnerID:   ownerID,
				Name:      name,
				Type:      category.CategoryTypeExpense,
				CreatedAt: createdAt,
				UpdatedAt: updatedAt,
			}, nil
		},
	}
	handler := &createCategoryLambda{categoryService: service}

	response, err := handler.Handle(context.Background(), events.APIGatewayV2HTTPRequest{
		Body: `{"name":"Groceries","type":"expense"}`,
	})

	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, response.StatusCode)

	var result category.CategoryRead
	require.NoError(t, json.Unmarshal([]byte(response.Body), &result))
	assert.Equal(t, category.CategoryRead{
		ID:           "category-1",
		Name:         "Groceries",
		CategoryType: category.CategoryTypeExpense,
		OwnerID:      "1",
		CreatedAt:    createdAt,
		UpdatedAt:    updatedAt,
	}, result)
}

func TestHandle_InvalidJSONPayload(t *testing.T) {
	handler := &createCategoryLambda{categoryService: &mockCategoryService{}}

	response, err := handler.Handle(context.Background(), events.APIGatewayV2HTTPRequest{
		Body: "invalid json",
	})

	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, response.StatusCode)
	assertErrorMessage(t, response, "Invalid JSON payload")
}

func TestHandle_MissingCategoryType(t *testing.T) {
	serviceCalled := false
	service := &mockCategoryService{
		createCategoryFunc: func(context.Context, string, string, string) (category.Category, error) {
			serviceCalled = true
			return category.Category{}, nil
		},
	}
	handler := &createCategoryLambda{categoryService: service}

	response, err := handler.Handle(context.Background(), events.APIGatewayV2HTTPRequest{
		Body: `{"name":"Groceries","type":"  "}`,
	})

	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, response.StatusCode)
	assertErrorMessage(t, response, "type is required")
	assert.False(t, serviceCalled)
}

func TestHandle_ServiceErrorReturnsInternalServerError(t *testing.T) {
	service := &mockCategoryService{
		createCategoryFunc: func(context.Context, string, string, string) (category.Category, error) {
			return category.Category{}, assert.AnError
		},
	}
	handler := &createCategoryLambda{categoryService: service}

	response, err := handler.Handle(context.Background(), events.APIGatewayV2HTTPRequest{
		Body: `{"name":"Groceries","type":"expense"}`,
	})

	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, response.StatusCode)
	assertErrorMessage(t, response, assert.AnError.Error())
}

func TestHandle_ServiceDeadline(t *testing.T) {
	handler := &createCategoryLambda{categoryService: &mockCategoryService{
		createCategoryFunc: func(context.Context, string, string, string) (category.Category, error) {
			return category.Category{}, context.DeadlineExceeded
		},
	}}
	response, err := handler.Handle(context.Background(), events.APIGatewayV2HTTPRequest{Body: `{"name":"Groceries","type":"expense"}`})
	require.NoError(t, err)
	assert.Equal(t, http.StatusGatewayTimeout, response.StatusCode)
	assertErrorMessage(t, response, context.DeadlineExceeded.Error())
}

func TestHandle_InvalidCategoryType(t *testing.T) {
	handler := &createCategoryLambda{categoryService: category.NewService(nil)}
	response, err := handler.Handle(context.Background(), events.APIGatewayV2HTTPRequest{Body: `{"type":"invalid"}`})
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, response.StatusCode)
	assertErrorMessage(t, response, "Invalid category type")
}

func TestHandle_DynamoDBErrors(t *testing.T) {
	tests := []struct {
		name   string
		status int
		code   string
		want   int
	}{
		{"throttled", 400, "ProvisionedThroughputExceededException", 429},
		{"request timeout", 408, "UnknownError", 408},
		{"rate limit", 429, "UnknownError", 429},
		{"bad gateway", 502, "UnknownError", 502},
		{"unavailable", 503, "UnknownError", 503},
		{"gateway timeout", 504, "UnknownError", 504},
		{"unexpected client error", 400, "UnknownError", 500},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/x-amz-json-1.0")
				w.WriteHeader(tt.status)
				_ = json.NewEncoder(w).Encode(map[string]string{"__type": tt.code, "message": "upstream failure"})
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
			handler := &createCategoryLambda{categoryService: category.NewService(category.NewRepository(db, "categories"))}
			response, err := handler.Handle(context.Background(), events.APIGatewayV2HTTPRequest{Body: `{"name":"Groceries","type":"expense"}`})
			require.NoError(t, err)
			assert.Equal(t, tt.want, response.StatusCode)
			assert.Contains(t, response.Body, "upstream failure")
		})
	}
}

func assertErrorMessage(t *testing.T, response events.APIGatewayV2HTTPResponse, want string) {
	t.Helper()

	var errResp errorResponse
	require.NoError(t, json.Unmarshal([]byte(response.Body), &errResp))
	assert.Equal(t, want, errResp.Message)
}
