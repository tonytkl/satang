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
	"github.com/tonytkl/satang/clients"
	"github.com/tonytkl/satang/transaction"
	"github.com/tonytkl/satang/utils"
)

type errorResponse struct {
	Message string `json:"message"`
}

type getTransactionLambda struct {
	service transaction.TransactionService
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

	repository := transaction.NewTransactionRepository(db, tableName)
	transactionService := transaction.NewTransactionService(repository)
	handler := &getTransactionLambda{service: transactionService}

	lambda.Start(handler.Handle)
}

func (handler *getTransactionLambda) Handle(ctx context.Context, request events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	transactionID := strings.TrimSpace(request.PathParameters["transaction_id"])
	if transactionID == "" {
		return utils.JsonResponse(http.StatusBadRequest, errorResponse{Message: "Transaction ID is required"})
	}

	// TODO: Get user from authentication context
	ownerID := "1"

	tx, err := handler.service.GetTransaction(ctx, transactionID, ownerID)

	if err != nil {
		if errors.Is(err, transaction.ErrTransactionNotFound) {
			return utils.JsonResponse(http.StatusNotFound, errorResponse{Message: "Transaction not found"})
		}
		return utils.JsonResponse(http.StatusBadRequest, errorResponse{Message: err.Error()})
	}

	if tx == nil {
		return utils.JsonResponse(http.StatusNotFound, errorResponse{Message: "Transaction not found"})
	}

	responseSchemas, err := transaction.BuildTransactionSchemas([]transaction.Transaction{*tx})
	if err != nil {
		return utils.JsonResponse(http.StatusInternalServerError, errorResponse{Message: err.Error()})
	}

	return utils.JsonResponse(http.StatusOK, responseSchemas[0])
}
