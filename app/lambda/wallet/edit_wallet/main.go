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

type editWalletRequest struct {
	Name       *string `json:"name"`
	WalletType *string `json:"walletType"`
}

type errorResponse struct {
	Message string `json:"message"`
}

type editWalletLambda struct {
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
	handler := &editWalletLambda{service: walletService}

	lambda.Start(handler.Handle)
}

func (handler *editWalletLambda) Handle(ctx context.Context, request events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	walletID := strings.TrimSpace(request.PathParameters["wallet_id"])
	if walletID == "" {
		return utils.JsonResponse(http.StatusBadRequest, errorResponse{Message: "Wallet ID is required"})
	}

	var payload editWalletRequest
	if err := json.Unmarshal([]byte(request.Body), &payload); err != nil {
		return utils.JsonResponse(http.StatusBadRequest, errorResponse{Message: "Invalid JSON payload"})
	}

	// TODO: Use actual OwnerID from token
	ownerID := "1"

	mapPayload := make(map[string]any, 2)
	if payload.Name != nil {
		mapPayload["Name"] = *payload.Name
	}
	if payload.WalletType != nil {
		mapPayload["Type"] = *payload.WalletType
	}

	editdWallet, err := handler.service.EditWallet(
		ctx,
		ownerID,
		walletID,
		mapPayload,
	)
	if err != nil {
		if errors.Is(err, wallet.ErrWalletNotFound) {
			return utils.JsonResponse(http.StatusNotFound, errorResponse{Message: "Wallet not found"})
		}
		return utils.JsonResponse(http.StatusBadRequest, errorResponse{Message: err.Error()})
	}

	walletResponse := wallet.BuildWalletRead(editdWallet)

	return utils.JsonResponse(http.StatusOK, walletResponse)
}
