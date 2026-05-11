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
	"github.com/tonytkl/satang/repositories"
	"github.com/tonytkl/satang/services"
	"github.com/tonytkl/satang/utils"
)

type createTransactionRequest struct {
	WalletID     string  `json:"walletId"`
	WalletName   string  `json:"walletName"`
	CategoryID   string  `json:"categoryId"`
	CategoryName string  `json:"categoryName"`
	Description  string  `json:"description"`
	Currency     string  `json:"currency"`
	ImageURL     string  `json:"imageUrl"`
	Type         string  `json:"type"`
	Amount       float64 `json:"amount"`
	Date         string  `json:"date"`
	OwnerID      string  `json:"ownerId"`
}

type errorResponse struct {
	Message string `json:"message"`
}

type createTransactionLambda struct {
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
	handler := &createTransactionLambda{service: transactionService}

	lambda.Start(handler.Handle)
}

func (handler *createTransactionLambda) Handle(ctx context.Context, request events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	var payload createTransactionRequest
	if err := json.Unmarshal([]byte(request.Body), &payload); err != nil {
		return utils.JsonResponse(http.StatusBadRequest, errorResponse{Message: "Invalid JSON payload"})
	}

	if err := validatePayload(payload); err != nil {
		return utils.JsonResponse(http.StatusBadRequest, errorResponse{Message: err.Error()})
	}

	date, err := utils.ParseDate(payload.Date)
	if err != nil {
		return utils.JsonResponse(http.StatusBadRequest, errorResponse{Message: "Date must be RFC3339 or YYYY-MM-DD"})
	}

	err = handler.service.CreateTransaction(
		ctx,
		payload.WalletID,
		payload.WalletName,
		payload.CategoryID,
		payload.CategoryName,
		payload.Description,
		payload.Currency,
		payload.ImageURL,
		payload.Type,
		payload.Amount,
		date,
		payload.OwnerID,
	)
	if err != nil {
		return utils.JsonResponse(http.StatusBadRequest, errorResponse{Message: err.Error()})
	}

	return events.APIGatewayV2HTTPResponse{
		StatusCode: http.StatusCreated,
	}, nil
}

func validatePayload(payload createTransactionRequest) error {
	if strings.TrimSpace(payload.WalletID) == "" {
		return errors.New("walletId is required")
	}
	if strings.TrimSpace(payload.WalletName) == "" {
		return errors.New("walletName is required")
	}
	if strings.TrimSpace(payload.CategoryID) == "" {
		return errors.New("categoryId is required")
	}
	if strings.TrimSpace(payload.CategoryName) == "" {
		return errors.New("categoryName is required")
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
	if strings.TrimSpace(payload.OwnerID) == "" {
		return errors.New("ownerId is required")
	}
	return nil
}
