package main

import (
	"context"
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

type errorResponse struct {
	Message string `json:"message"`
}

type getCategoryLambda struct {
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
	handler := &getCategoryLambda{service: categoryService}

	lambda.Start(handler.Handle)
}

func (handler *getCategoryLambda) Handle(ctx context.Context, request events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	categoryID := strings.TrimSpace(request.PathParameters["category_id"])
	if categoryID == "" {
		return utils.JsonResponse(http.StatusBadRequest, errorResponse{Message: "Category ID is required"})
	}

	// TODO: Get user from authentication context
	ownerID := "1"

	c, err := handler.service.GetCategory(ctx, ownerID, categoryID)
	if err != nil {
		if errors.Is(err, category.ErrCategoryNotFound) {
			return utils.JsonResponse(http.StatusNotFound, errorResponse{Message: "Category not found"})
		}
		return utils.JsonResponse(http.StatusBadRequest, errorResponse{Message: err.Error()})
	}
	if c == (category.Category{}) {
		return utils.JsonResponse(http.StatusNotFound, errorResponse{Message: "Category not found"})
	}

	responseSchemas := category.BuildCategoryRead(c)
	return utils.JsonResponse(http.StatusOK, responseSchemas)
}
