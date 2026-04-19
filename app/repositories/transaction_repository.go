package repositories

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/tonytkl/satang/clients"
	"github.com/tonytkl/satang/model"
	"github.com/tonytkl/satang/utils"
)

var ErrTransactionNotFound = errors.New("transaction not found")

type TransactionRepository interface {
	Create(ctx context.Context, transaction *model.Transaction) error
	GetByKey(ctx context.Context, transactionID string) (*model.Transaction, error)
	ListByWallet()
	ListByCategory()
	ListWithinDateRange()
	Update()
	Delete()
}

type transactionRepository struct {
	db        clients.DynamoDBClient
	tableName string
}

func NewTransactionRepository(db clients.DynamoDBClient, tableName string) TransactionRepository {
	return &transactionRepository{
		db:        db,
		tableName: tableName,
	}
}

func (repository *transactionRepository) Create(ctx context.Context, transaction *model.Transaction) error {
	if err := validateTransaction(transaction); err != nil {
		return err
	}

	transaction.ID = utils.GetUUID()
	transaction.PK = utils.GetPartitionKey("WALLET", transaction.WalletID)
	transaction.SK = utils.GetSortingKey("TX", transaction.Date, transaction.ID)

	if transaction.CreatedAt.IsZero() {
		transaction.CreatedAt = time.Now().UTC()
	}
	if transaction.UpdatedAt.IsZero() {
		transaction.UpdatedAt = transaction.CreatedAt
	}

	err := repository.db.PutItem(ctx, repository.tableName, transaction)
	if err != nil {
		return err
	}

	return nil
}

func (repository *transactionRepository) GetByKey(ctx context.Context, transactionID string) (*model.Transaction, error) {
	if transactionID == "" {
		return nil, errors.New("Transaction ID is required")
	}

	transactions := []model.Transaction{}
	expression := map[string]any{
		":pk": utils.GetPartitionKey("TX_ID", transactionID),
	}
	err := repository.db.QueryItems(
		ctx,
		repository.tableName,
		"GSI2_PK = :pk",
		expression,
		"GSI2",
		transactions)

	if err != nil {
		return nil, fmt.Errorf("query transaction by ID: %w", err)
	}

	if len(transactions) == 0 {
		return nil, ErrTransactionNotFound
	}

	return &transactions[0], nil
}

func (repository *transactionRepository) ListByWallet() {

}

func (repository *transactionRepository) ListByCategory() {

}

func (repository *transactionRepository) ListWithinDateRange() {

}

func (repository *transactionRepository) Update() {

}

func (repository *transactionRepository) Delete() {

}

func validateTransaction(transaction *model.Transaction) error {
	if transaction == nil {
		return errors.New("Transaction is required")
	}

	if transaction.Amount == 0 {
		return errors.New("Transaction amount is required")
	}

	if transaction.Currency == "" {
		return errors.New("Transaction currency is required")
	}

	if transaction.WalletID == "" {
		return errors.New("Wallet is required")
	}

	if transaction.CategoryID == "" {
		return errors.New("Category is required")
	}

	if transaction.Date.IsZero() {
		return errors.New("Transaction date is required")
	}

	return nil
}
