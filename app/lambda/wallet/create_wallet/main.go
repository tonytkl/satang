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
	"github.com/tonytkl/satang/clients"
	"github.com/tonytkl/satang/transaction"
	"github.com/tonytkl/satang/utils"
	"github.com/tonytkl/satang/wallet"
)

type createWalletRequest struct {
	Name       string  `json:"Name"`
	Currency   string  `json:"currency"`
	Balance    float64 `json:"balance"`
	WalletType string  `json:"walleyType"`
}

type errorResponse struct {
	Message string `json:"message"`
}

type createWalletLambda struct {
	service wallet.Service
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

	walletRepository := wallet.NewRepository(db, tableName)
	transactionRepository := transaction.NewRepository(db, tableName)
	transactionService := transaction.NewService(transactionRepository)
	walletService := wallet.NewService(walletRepository, transactionService)
	handler := &createWalletLambda{service: walletService}

	lambda.Start(handler.Handle)
}

func (handler *createWalletLambda) Handle(ctx context.Context, request events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	var payload createWalletRequest
	if err := json.Unmarshal([]byte(request.Body), &payload); err != nil {
		return utils.JsonResponse(http.StatusBadRequest, errorResponse{Message: "Invalid JSON payload"})
	}

	if err := validatePayload(payload); err != nil {
		return utils.JsonResponse(http.StatusBadRequest, errorResponse{Message: err.Error()})
	}

	// TODO: Use actual OwnerID from token
	ownerID := "1"

	err := handler.service.CreateWallet(
		ctx,
		ownerID,
		payload.Name,
		payload.Currency,
		payload.Balance,
		payload.WalletType,
	)
	if err != nil {
		return utils.JsonResponse(http.StatusBadRequest, errorResponse{Message: err.Error()})
	}

	return events.APIGatewayV2HTTPResponse{
		StatusCode: http.StatusCreated,
	}, nil
}

func validatePayload(payload createWalletRequest) error {
	if strings.TrimSpace(payload.WalletType) == "" {
		return errors.New("type is required")
	}
	return nil
}
