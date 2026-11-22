package main

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tonytkl/satang/category"
	"github.com/tonytkl/satang/clients"
	"github.com/tonytkl/satang/transaction"
	"github.com/tonytkl/satang/utils"
	"github.com/tonytkl/satang/wallet"
)

type atomicWalletService struct {
	wallet.Service
	err error
}

func (service atomicWalletService) GetWallet(context.Context, string, string) (wallet.Wallet, error) {
	if service.err != nil {
		return wallet.Wallet{}, service.err
	}
	return wallet.Wallet{Name: "Main"}, nil
}

type atomicCategoryService struct {
	category.Service
	err error
}

func (service atomicCategoryService) GetCategory(context.Context, string, string) (category.Category, error) {
	if service.err != nil {
		return category.Category{}, service.err
	}
	return category.Category{Name: "Food"}, nil
}

type atomicTransactionService struct {
	transaction.Service
}

func (atomicTransactionService) PrepareCreateTransaction(walletID, walletName, categoryID, categoryName, description, currency, imageURL, txType string, amount float64, date time.Time, ownerID string) (clients.WriteOp, error) {
	return clients.WriteOp{
		Kind:  clients.WritePut,
		Table: "items",
		Item:  map[string]string{"walletID": walletID, "categoryID": categoryID},
	}, nil
}

type atomicWriter struct {
	ops []clients.WriteOp
	err error
}

func (writer *atomicWriter) TransactWrite(_ context.Context, ops ...clients.WriteOp) error {
	writer.ops = ops
	return writer.err
}

func TestHandle_AtomicallyChecksWalletAndCategory(t *testing.T) {
	writer := &atomicWriter{}
	handler := &createTransactionLambda{
		txService:       atomicTransactionService{},
		walletService:   atomicWalletService{},
		categoryService: atomicCategoryService{},
		writer:          writer,
		tableName:       "items",
	}

	response, err := handler.Handle(context.Background(), events.APIGatewayV2HTTPRequest{
		Body: `{"walletId":"wallet-1","categoryId":"category-1","currency":"THB","type":"expense","amount":10,"date":"2025-05-17"}`,
	})

	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, response.StatusCode)
	require.Len(t, writer.ops, 3)

	assert.Equal(t, clients.WriteCheck, writer.ops[0].Kind)
	assert.Equal(t, "items", writer.ops[0].Table)
	assert.Equal(t, map[string]any{
		"PK": utils.GetPartitionKey("USER", "1"),
		"SK": utils.GetPartitionKey("WALLET", "wallet-1"),
	}, writer.ops[0].Key)
	assert.Equal(t, "attribute_exists(PK)", writer.ops[0].Condition)

	assert.Equal(t, clients.WriteCheck, writer.ops[1].Kind)
	assert.Equal(t, map[string]any{
		"PK": utils.GetPartitionKey("USER", "1"),
		"SK": utils.GetPartitionKey("CATEGORY", "category-1"),
	}, writer.ops[1].Key)
	assert.Equal(t, "attribute_exists(PK)", writer.ops[1].Condition)

	assert.Equal(t, clients.WritePut, writer.ops[2].Kind)
	assert.Equal(t, "items", writer.ops[2].Table)
}

func TestHandle_PropagatesAtomicWriteFailure(t *testing.T) {
	writer := &atomicWriter{err: errors.New("transaction canceled")}
	handler := &createTransactionLambda{
		txService:       atomicTransactionService{},
		walletService:   atomicWalletService{},
		categoryService: atomicCategoryService{},
		writer:          writer,
		tableName:       "items",
	}

	response, err := handler.Handle(context.Background(), events.APIGatewayV2HTTPRequest{
		Body: `{"walletId":"wallet-1","categoryId":"category-1","currency":"THB","type":"expense","amount":10,"date":"2025-05-17"}`,
	})

	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, response.StatusCode)
	assert.Len(t, writer.ops, 3)
}

func TestHandle_ReportsReferenceRemovedBeforeAtomicWrite(t *testing.T) {
	writer := &atomicWriter{err: &types.TransactionCanceledException{
		CancellationReasons: []types.CancellationReason{
			{Code: aws.String("None")},
			{Code: aws.String("ConditionalCheckFailed")},
			{Code: aws.String("None")},
		},
	}}
	handler := &createTransactionLambda{
		txService:       atomicTransactionService{},
		walletService:   atomicWalletService{},
		categoryService: atomicCategoryService{},
		writer:          writer,
		tableName:       "items",
	}

	response, err := handler.Handle(context.Background(), events.APIGatewayV2HTTPRequest{
		Body: `{"walletId":"wallet-1","categoryId":"category-1","currency":"THB","type":"expense","amount":10,"date":"2025-05-17"}`,
	})

	require.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, response.StatusCode)
	assert.Contains(t, response.Body, "No category match given query")
}
