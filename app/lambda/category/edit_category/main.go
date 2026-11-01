package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/tonytkl/satang/category"
	"github.com/tonytkl/satang/clients"
	"github.com/tonytkl/satang/utils"
)

type editCategoryRequest struct {
	Name         *string `json:"name"`
	CategoryType *string `json:"categoryType"`
}

type errorResponse struct {
	Message string `json:"message"`
}

type editCategoryLambda struct {
	service category.Service
}

func main() {
	ctx := context.Background()
	db, err := clients.NewDynamoDBClient(ctx)
	if err != nil {
		panic(fmt.Errorf("create dynamodb client: %w", err))
	}

	tableName := os.Getenv("TABLE_NAME")
	if strings.TrimSpace(tableName) == "" {
		panic("TABLE_NAME is required")
	}

	categoryRepository := category.NewRepository(db, tableName)
	categoryService := category.NewService(categoryRepository)
	handler := &editCategoryLambda{service: categoryService}

	lambda.Start(handler.Handle)
}

func (handler *editCategoryLambda) Handle(ctx context.Context, request events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	categoryID := strings.TrimSpace(request.PathParameters["category_id"])
	if categoryID == "" {
		return utils.JsonResponse(http.StatusBadRequest, errorResponse{Message: "Category ID is required"})
	}

	var payload editCategoryRequest
	if err := json.Unmarshal([]byte(request.Body), &payload); err != nil {
		return utils.JsonResponse(http.StatusBadRequest, errorResponse{Message: "Invalid JSON payload"})
	}

	// TODO: Use actual OwnerID from token
	ownerID := "1"

	mapPayload := make(map[string]any, 2)
	if payload.Name != nil {
		mapPayload["Name"] = *payload.Name
	}
	if payload.CategoryType != nil {
		mapPayload["Type"] = *payload.CategoryType
	}

	editdCategory, err := handler.service.EditCategory(
		ctx,
		ownerID,
		categoryID,
		mapPayload,
	)
	if err != nil {
		if errors.Is(err, category.ErrCategoryNotFound) {
			return utils.JsonResponse(http.StatusNotFound, errorResponse{Message: "Category not found"})
		}
		return utils.JsonResponse(http.StatusBadRequest, errorResponse{Message: err.Error()})
	}

	categoryResponse := category.BuildCategoryRead(editdCategory)

	return utils.JsonResponse(http.StatusOK, categoryResponse)
}
