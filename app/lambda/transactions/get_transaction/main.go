package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/tonytkl/satang/clients"
	"github.com/tonytkl/satang/repositories"
	"github.com/tonytkl/satang/services"
	"github.com/tonytkl/satang/utils"
)

type errorResponse struct {
	Message string `json:"message"`
}

type getTransactionLambda struct {
	service services.TransactionService
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

	repository := repositories.NewTransactionRepository(db, tableName)
	transactionService := services.NewTransactionService(repository)
	handler := &getTransactionLambda{service: transactionService}

	lambda.Start(handler.Handle)
}

func (handler *getTransactionLambda) Handle(ctx context.Context, request events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	transactionID := request.PathParameters["transaction_id"]
	transaction, err := handler.service.GetTransaction(ctx, transactionID)

	if err == repositories.ErrTransactionNotFound {
		return utils.JsonResponse(http.StatusNotFound, errorResponse{Message: err.Error()})
	}

	if err != nil {
		return utils.JsonResponse(http.StatusBadRequest, errorResponse{Message: err.Error()})
	}

	return utils.JsonResponse(http.StatusOK, transaction)
}
