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
	"github.com/tonytkl/satang/wallet"
)

type errorResponse struct {
	Message string `json:"message"`
}

type getWalletLambda struct {
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
	handler := &getWalletLambda{service: walletService}

	lambda.Start(handler.Handle)
}

func (handler *getWalletLambda) Handle(ctx context.Context, request events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	walletID := strings.TrimSpace(request.PathParameters["wallet_id"])
	if walletID == "" {
		return utils.JsonResponse(http.StatusBadRequest, errorResponse{Message: "Wallet ID is required"})
	}

	// TODO: Get user from authentication context
	ownerID := "1"

	w, err := handler.service.GetWallet(ctx, ownerID, walletID)
	if err != nil {
		if errors.Is(err, wallet.ErrWalletNotFound) {
			return utils.JsonResponse(http.StatusNotFound, errorResponse{Message: "Wallet not found"})
		}
		return utils.JsonResponse(http.StatusBadRequest, errorResponse{Message: err.Error()})
	}
	if w == (wallet.Wallet{}) {
		return utils.JsonResponse(http.StatusNotFound, errorResponse{Message: "Wallet not found"})
	}

	responseSchemas := wallet.BuildWalletRead([]wallet.Wallet{w})
	return utils.JsonResponse(http.StatusOK, responseSchemas[0])
}
