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

type createCategoryRequest struct {
	Name         string `json:"name"`
	CategoryType string `json:"type"`
}

type errorResponse struct {
	Message string `json:"message"`
}

type createCategoryLambda struct {
	categoryService category.Service
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
	handler := &createCategoryLambda{categoryService: categoryService}

	lambda.Start(handler.Handle)
}

func (handler *createCategoryLambda) Handle(ctx context.Context, request events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	var payload createCategoryRequest
	if err := json.Unmarshal([]byte(request.Body), &payload); err != nil {
		return utils.JsonResponse(http.StatusBadRequest, errorResponse{Message: "Invalid JSON payload"})
	}

	if err := validatePayload(payload); err != nil {
		return utils.JsonResponse(http.StatusBadRequest, errorResponse{Message: err.Error()})
	}

	// TODO: Use actual OwnerID from token
	ownerID := "1"

	createdCategory, err := handler.categoryService.CreateCategory(
		ctx,
		ownerID,
		payload.Name,
		payload.CategoryType,
	)
	if err != nil {
		return utils.JsonResponse(http.StatusBadRequest, errorResponse{Message: err.Error()})
	}

	categoryResponse := category.BuildCategoryRead(createdCategory)

	return utils.JsonResponse(http.StatusCreated, categoryResponse)
}

func validatePayload(payload createCategoryRequest) error {
	if strings.TrimSpace(payload.CategoryType) == "" {
		return errors.New("type is required")
	}
	return nil
}
