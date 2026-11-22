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
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/tonytkl/satang/category"
	"github.com/tonytkl/satang/clients"
	"github.com/tonytkl/satang/transaction"
	"github.com/tonytkl/satang/utils"
	"github.com/tonytkl/satang/wallet"
)

type createTransactionRequest struct {
	WalletID    string  `json:"walletId"`
	CategoryID  string  `json:"categoryId"`
	Description string  `json:"description"`
	Currency    string  `json:"currency"`
	ImageURL    string  `json:"imageUrl"`
	Type        string  `json:"type"`
	Amount      float64 `json:"amount"`
	Date        string  `json:"date"`
}

type errorResponse struct {
	Message string `json:"message"`
}

type createTransactionLambda struct {
	txService       transaction.Service
	walletService   wallet.Service
	categoryService category.Service
	writer          clients.TransactionalWriter
	tableName       string
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

	// Transaction dependency
	txRepository := transaction.NewRepository(db, tableName)
	txService := transaction.NewService(txRepository)

	// Wallet dependency
	walletRepository := wallet.NewRepository(db, tableName)
	walletService := wallet.NewService(walletRepository)

	// Category dependency
	categoryRepository := category.NewRepository(db, tableName)
	categoryService := category.NewService(categoryRepository)

	handler := &createTransactionLambda{
		txService:       txService,
		walletService:   walletService,
		categoryService: categoryService,
		writer:          db,
		tableName:       tableName,
	}

	lambda.Start(handler.Handle)
}

func (handler *createTransactionLambda) Handle(ctx context.Context, request events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	var payload createTransactionRequest
	if err := json.Unmarshal([]byte(request.Body), &payload); err != nil {
		return utils.JsonResponse(http.StatusBadRequest, errorResponse{Message: "Invalid JSON payload"})
	}

	// Validate payload
	if err := validatePayload(payload); err != nil {
		return utils.JsonResponse(http.StatusBadRequest, errorResponse{Message: err.Error()})
	}

	// Validate datetime
	date, err := utils.ParseDate(payload.Date)
	if err != nil {
		return utils.JsonResponse(http.StatusBadRequest, errorResponse{Message: "Date must be RFC3339 or YYYY-MM-DD"})
	}

	// TODO: Use actual ownerID from token
	var ownerID = "1"

	// Validate wallet
	walletObj, err := handler.walletService.GetWallet(ctx, ownerID, payload.WalletID)
	if err != nil {
		if errors.Is(err, wallet.ErrWalletNotFound) {
			return utils.JsonResponse(http.StatusNotFound, errorResponse{Message: "No wallet match given query"})
		}
		return utils.JsonResponse(http.StatusInternalServerError, errorResponse{Message: "Failed to query wallet"})
	}

	// Validate category
	categoryObj, err := handler.categoryService.GetCategory(ctx, ownerID, payload.CategoryID)
	if err != nil {
		if errors.Is(err, category.ErrCategoryNotFound) {
			return utils.JsonResponse(http.StatusNotFound, errorResponse{Message: "No category match given query"})
		}
		return utils.JsonResponse(http.StatusInternalServerError, errorResponse{Message: "Failed to query category"})
	}

	createdTransaction, transactionOp, err := handler.txService.PrepareCreateTransaction(
		payload.WalletID,
		walletObj.Name,
		payload.CategoryID,
		categoryObj.Name,
		payload.Description,
		payload.Currency,
		payload.ImageURL,
		payload.Type,
		payload.Amount,
		date,
		ownerID,
	)
	if err != nil {
		return utils.JsonResponse(utils.ErrorStatusCode(err), errorResponse{Message: err.Error()})
	}

	response, err := transaction.BuildTransactionRead([]transaction.Transaction{createdTransaction})
	if err != nil {
		return utils.JsonResponse(http.StatusInternalServerError, errorResponse{Message: err.Error()})
	}

	walletCheck := clients.WriteOp{
		Kind:  clients.WriteCheck,
		Table: handler.tableName,
		Key: map[string]any{
			"PK": utils.GetPartitionKey("USER", ownerID),
			"SK": utils.GetPartitionKey("WALLET", payload.WalletID),
		},
		Condition: "attribute_exists(PK)",
	}
	categoryCheck := clients.WriteOp{
		Kind:  clients.WriteCheck,
		Table: handler.tableName,
		Key: map[string]any{
			"PK": utils.GetPartitionKey("USER", ownerID),
			"SK": utils.GetPartitionKey("CATEGORY", payload.CategoryID),
		},
		Condition: "attribute_exists(PK)",
	}
	if err := handler.writer.TransactWrite(ctx, walletCheck, categoryCheck, transactionOp); err != nil {
		var canceled *types.TransactionCanceledException
		if errors.As(err, &canceled) {
			referenceChecks := []string{
				"No wallet match given query",
				"No category match given query",
			}
			for i, message := range referenceChecks {
				if i < len(canceled.CancellationReasons) &&
					canceled.CancellationReasons[i].Code != nil &&
					*canceled.CancellationReasons[i].Code == "ConditionalCheckFailed" {
					return utils.JsonResponse(http.StatusNotFound, errorResponse{Message: message})
				}
			}
		}
		return utils.JsonResponse(utils.ErrorStatusCode(err), errorResponse{Message: err.Error()})
	}

	return utils.JsonResponse(http.StatusCreated, response[0])
}

func validatePayload(payload createTransactionRequest) error {
	if strings.TrimSpace(payload.WalletID) == "" {
		return errors.New("walletId is required")
	}
	if strings.TrimSpace(payload.CategoryID) == "" {
		return errors.New("categoryId is required")
	}
	if strings.TrimSpace(payload.Currency) == "" {
		return errors.New("currency is required")
	}
	if strings.TrimSpace(payload.Type) == "" {
		return errors.New("type is required")
	}
	if payload.Amount == 0 {
		return errors.New("amount is required")
	}
	if strings.TrimSpace(payload.Date) == "" {
		return errors.New("date is required")
	}
	return nil
}
