package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/tonytkl/satang/clients"
	"github.com/tonytkl/satang/model"
	"github.com/tonytkl/satang/repositories"
	"github.com/tonytkl/satang/services"
	"github.com/tonytkl/satang/utils"
)

type errorResponse struct {
	Message string `json:"message"`
}

type getListTransactionsLambda struct {
	service services.TransactionService
}

type getListTransactionResponse struct {
	Transactions []model.Transaction `json:"transactions"`
	NextToken    string              `json:"nextToken"`
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

	handler := getListTransactionsLambda{service: transactionService}

	lambda.Start(handler.Handle)
}

func (handler *getListTransactionsLambda) Handle(ctx context.Context, request events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	// TODO: Get user from authentication context
	ownerID := "1"

	// Date filter
	fromDateQuery := request.QueryStringParameters["fromDate"]
	toDateQuery := request.QueryStringParameters["toDate"]

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

	// Set from date
	now := time.Now()
	var fromDate time.Time
	if fromDateQuery == "" {
		fromDate = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	} else {
		var err error
		fromDate, err = utils.ParseDate(fromDateQuery)
		if err != nil {
			return utils.JsonResponse(http.StatusBadRequest, errorResponse{Message: err.Error()})
		}
	}

	var toDate time.Time
	if toDateQuery == "" {
		toDate = now
	} else {
		var err error
		toDate, err = utils.ParseDate(toDateQuery)
		if err != nil {
			return utils.JsonResponse(http.StatusBadRequest, errorResponse{Message: err.Error()})
		}
	}
	transactions, nextToken, err := handler.service.GetTransactionsBetweenPeriod(ctx, ownerID, fromDate, toDate, limit, nextTokenQuery)
	if err != nil {
		return utils.JsonResponse(http.StatusInternalServerError, errorResponse{Message: err.Error()})
	}

	response := getListTransactionResponse{
		Transactions: transactions,
		NextToken:    nextToken,
	}
	return utils.JsonResponse(http.StatusOK, response)
}
