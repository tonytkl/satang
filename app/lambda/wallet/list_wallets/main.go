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
	"github.com/tonytkl/satang/clients"
	"github.com/tonytkl/satang/transaction"
	"github.com/tonytkl/satang/utils"
	"github.com/tonytkl/satang/wallet"
)

type errorResponse struct {
	Message string `json:"message"`
}

type listWalletsLambda struct {
	service wallet.Service
}

type listWalletsResponse struct {
	Wallets   []wallet.WalletRead `json:"wallets"`
	NextToken string              `json:"nextToken"`
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

	handler := listWalletsLambda{service: walletService}

	lambda.Start(handler.Handle)
}

func (handler *listWalletsLambda) Handle(ctx context.Context, request events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	// TODO: Get user from authentication context
	ownerID := "1"

	// Pagination
	nextTokenQuery := request.QueryStringParameters["nextToken"]
	limitQuery := request.QueryStringParameters["limit"]

	var limit int32
	if limitQuery != "" {
		parsedLimit, err := strconv.ParseInt(limitQuery, 10, 32)
		if err != nil {
			return utils.JsonResponse(http.StatusBadRequest, errorResponse{Message: "limit must be a valid integer"})
		}
		limit = int32(parsedLimit)
	}

	wallets, nextToken, err := handler.service.ListWallets(ctx, ownerID, nextTokenQuery, limit)
	if err != nil {
		return utils.JsonResponse(http.StatusInternalServerError, errorResponse{Message: err.Error()})
	}

	responseWallets := wallet.BuildWalletRead(wallets)

	response := listWalletsResponse{
		Wallets:   responseWallets,
		NextToken: nextToken,
	}
	return utils.JsonResponse(http.StatusOK, response)
}
