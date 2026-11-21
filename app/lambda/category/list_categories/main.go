package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"
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

type listCategorysLambda struct {
	service category.Service
}

type listCategorysResponse struct {
	Categorys  []category.CategoryRead `json:"categories"`
	NextCursor string                  `json:"nextCursor"`
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

	handler := listCategorysLambda{service: categoryService}

	lambda.Start(handler.Handle)
}

func (handler *listCategorysLambda) Handle(ctx context.Context, request events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	// TODO: Get user from authentication context
	ownerID := "1"

	// Pagination
	cursorQuery := request.QueryStringParameters["cursor"]
	limitQuery := request.QueryStringParameters["limit"]

	var limit int32
	if limitQuery != "" {
		parsedLimit, err := strconv.ParseInt(limitQuery, 10, 32)
		if err != nil {
			return utils.JsonResponse(http.StatusBadRequest, errorResponse{Message: "limit must be a valid integer"})
		}

		if parsedLimit < 0 {
			return utils.JsonResponse(http.StatusBadRequest, errorResponse{Message: "limit must be greater than or equal to 0"})
		}

		limit = int32(parsedLimit)
	}

	categorys, cursor, err := handler.service.ListCategories(ctx, ownerID, cursorQuery, limit)
	if err != nil {
		return utils.JsonResponse(utils.ErrorStatusCode(err), errorResponse{Message: utils.ErrorMessage(err)})
	}

	responseCategorys := category.BuildCategoryReadList(categorys)

	response := listCategorysResponse{
		Categorys:  responseCategorys,
		NextCursor: cursor,
	}
	return utils.JsonResponse(http.StatusOK, response)
}
