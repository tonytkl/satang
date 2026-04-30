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

// ErrTransactionNotFound is returned when no transaction matches the query.
var ErrTransactionNotFound = errors.New("transaction not found")

// TransactionRepository defines persistence operations for transactions.
type TransactionRepository interface {
	Create(ctx context.Context, transaction *model.Transaction) error
	ListByGSI(ctx context.Context, indexName string, indexPartitionKeyPrefix string, targetID string, ownerID string, fromDate *time.Time, toDate *time.Time) ([]model.Transaction, error)
	ListWithinDateRange(ctx context.Context, ownerID string, fromDate time.Time, toDate time.Time) ([]model.Transaction, error)
	GetByKey(ctx context.Context, id string) (*model.Transaction, error)
	Update(ctx context.Context, ownerID string, transactionDate string, transactionID string, transaction *model.Transaction) error
	Delete(ctx context.Context, ownerID string, transactionDate string, transactionID string) error
}

type transactionRepository struct {
	db        clients.DynamoDBClient
	tableName string
}

// NewTransactionRepository creates a transaction repository backed by DynamoDB.
func NewTransactionRepository(db clients.DynamoDBClient, tableName string) TransactionRepository {
	return &transactionRepository{
		db:        db,
		tableName: tableName,
	}
}

// Create stores a transaction and populates its derived keys and timestamps.
func (repository *transactionRepository) Create(ctx context.Context, transaction *model.Transaction) error {
	if err := validateTransaction(transaction); err != nil {
		return err
	}
	sortingKey := utils.GetSortingKey("TX", transaction.Date, transaction.ID)
	transaction.ID = utils.GetUUID()
	transaction.PK = utils.GetPartitionKey("USER", transaction.OwnerID)
	transaction.SK = sortingKey

	transaction.GSI_ByCategoryPK = utils.GetPartitionKey("TX_CATEGORY", transaction.CategoryID)
	transaction.GSI_ByCategorySK = sortingKey

	transaction.GSI_ByWalletPK = utils.GetPartitionKey("TX_WALLET", transaction.WalletID)
	transaction.GSI_ByWalletSK = sortingKey

	transaction.GSI_ByTransactionID = "TX_ID#" + transaction.ID

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

// ListByGSI lists transactions using the provided GSI name and partition key prefix.
func (repository *transactionRepository) ListByGSI(ctx context.Context, indexName string, indexPartitionKeyPrefix string, targetID string, ownerID string, fromDate *time.Time, toDate *time.Time) ([]model.Transaction, error) {
	if targetID == "" || indexName == "" || indexPartitionKeyPrefix == "" {
		return nil, errors.New("index name, index partition key prefix, and target ID are required")
	}

	indexPartitionKeyField, err := getIndexPartitionKeyField(indexName)
	if err != nil {
		return nil, err
	}

	queryExpression := indexPartitionKeyField + " = :indexPK"
	expressionValues := map[string]any{
		":indexPK": utils.GetPartitionKey(indexPartitionKeyPrefix, targetID),
	}
	filterExpression := ""
	if ownerID != "" {
		expressionValues[":ownerPK"] = utils.GetPartitionKey("USER", ownerID)
		filterExpression = "PK = :ownerPK"
	}

	if fromDate != nil && toDate != nil {
		if fromDate.After(*toDate) {
			return nil, errors.New("from date must not be after to date")
		}

		indexSortKeyField, err := getIndexSortKeyField(indexName)
		if err != nil {
			return nil, err
		}

		queryExpression += " AND " + indexSortKeyField + " BETWEEN :from AND :to"
		expressionValues[":from"] = utils.GetSortingKey("TX", *fromDate, "")
		expressionValues[":to"] = utils.GetSortingKey("TX", *toDate, "")
	}

	transactions := []model.Transaction{}

	err = repository.db.QueryItems(
		ctx,
		repository.tableName,
		queryExpression,
		expressionValues,
		indexName,
		filterExpression,
		&transactions,
	)

	if err != nil {
		return nil, fmt.Errorf("query transaction by ID: %w", err)
	}

	if len(transactions) == 0 {
		return nil, ErrTransactionNotFound
	}

	return transactions, nil
}

// ListWithinDateRange lists transactions for a user within a date range.
func (repository *transactionRepository) ListWithinDateRange(ctx context.Context, ownerID string, fromDate time.Time, toDate time.Time) ([]model.Transaction, error) {
	if ownerID == "" {
		return nil, errors.New("owner ID is required")
	}
	if fromDate.IsZero() || toDate.IsZero() {
		return nil, errors.New("from date and to date are required")
	}
	if fromDate.After(toDate) {
		return nil, errors.New("from date must not be after to date")
	}

	fromSK := utils.GetSortingKey("TX", fromDate, "")
	toSK := utils.GetSortingKey("TX", toDate, "")

	transactions := []model.Transaction{}
	err := repository.db.QueryItems(
		ctx,
		repository.tableName,
		"PK = :pk AND SK BETWEEN :from AND :to",
		map[string]any{
			":pk":   utils.GetPartitionKey("USER", ownerID),
			":from": fromSK,
			":to":   toSK,
		},
		"",
		"",
		&transactions,
	)
	if err != nil {
		return nil, fmt.Errorf("list transactions within date range: %w", err)
	}

	return transactions, nil
}

func (repository *transactionRepository) GetByKey(ctx context.Context, id string) (*model.Transaction, error) {
	if id == "" {
		return nil, errors.New("ID is required")
	}

	transactions, err := repository.ListByGSI(ctx, "GSI3", "TX_ID", id, "", nil, nil)

	if err != nil {
		return nil, err
	}

	return &transactions[0], nil
}

// Update modifies mutable attributes of an existing transaction.
func (repository *transactionRepository) Update(ctx context.Context, ownerID string, transactionDate string, transactionID string, transaction *model.Transaction) error {
	if ownerID == "" {
		return errors.New("owner ID is required")
	}
	if transactionDate == "" {
		return errors.New("transaction date is required")
	}
	if transactionID == "" {
		return errors.New("transaction ID is required")
	}
	if transaction == nil {
		return errors.New("transaction is required")
	}

	date, err := time.Parse("2006-01-02", transactionDate)
	if err != nil {
		return fmt.Errorf("invalid transaction date format: %w", err)
	}

	sortingKey := utils.GetSortingKey("TX", date, transactionID)
	updatedAt := time.Now().UTC()

	key := map[string]any{
		"PK": utils.GetPartitionKey("USER", ownerID),
		"SK": sortingKey,
	}

	updateExpression := "SET WalletID = :walletID, WalletName = :walletName, Amount = :amount, Currency = :currency, CategoryID = :categoryID, CategoryName = :categoryName, Description = :description, ImageURL = :imageURL, GSI_PK = :gsiCategoryPK, GSI2_PK = :gsiWalletPK, UpdatedAt = :updatedAt"

	expressionValues := map[string]any{
		":walletID":      transaction.WalletID,
		":walletName":    transaction.WalletName,
		":amount":        transaction.Amount,
		":currency":      transaction.Currency,
		":categoryID":    transaction.CategoryID,
		":categoryName":  transaction.CategoryName,
		":description":   transaction.Description,
		":imageURL":      transaction.ImageURL,
		":gsiCategoryPK": utils.GetPartitionKey("TX_CATEGORY", transaction.CategoryID),
		":gsiWalletPK":   utils.GetPartitionKey("TX_WALLET", transaction.WalletID),
		":updatedAt":     updatedAt,
		":transactionID": transactionID,
	}

	conditionExpression := "attribute_exists(PK) AND attribute_exists(SK) AND ID = :transactionID"

	if err := repository.db.UpdateItem(ctx, repository.tableName, key, updateExpression, expressionValues, conditionExpression); err != nil {
		return fmt.Errorf("update transaction: %w", err)
	}

	return nil
}

// Delete removes an existing transaction.
func (repository *transactionRepository) Delete(ctx context.Context, ownerID string, transactionDate string, transactionID string) error {
	if ownerID == "" {
		return errors.New("owner ID is required")
	}
	if transactionDate == "" {
		return errors.New("transaction date is required")
	}
	if transactionID == "" {
		return errors.New("transaction ID is required")
	}

	date, err := time.Parse("2006-01-02", transactionDate)
	if err != nil {
		return fmt.Errorf("invalid transaction date format: %w", err)
	}
	sortingKey := utils.GetSortingKey("TX", date, transactionID)

	key := map[string]any{
		"PK": "USER#" + ownerID,
		"SK": sortingKey,
	}

	if err := repository.db.DeleteItem(ctx, repository.tableName, key); err != nil {
		return fmt.Errorf("delete transaction: %w", err)
	}
	return nil
}

// validateTransaction ensures the required transaction fields are present.
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

// getIndexPartitionKeyField resolves a GSI name to its partition key attribute.
func getIndexPartitionKeyField(indexName string) (string, error) {
	switch indexName {
	case "GSI1":
		return "GSI_PK", nil
	case "GSI2":
		return "GSI2_PK", nil
	case "GSI3":
		return "GSI3_PK", nil
	default:
		return "", fmt.Errorf("unsupported index name: %s", indexName)
	}
}

// getIndexSortKeyField resolves a GSI name to its sort key attribute.
func getIndexSortKeyField(indexName string) (string, error) {
	switch indexName {
	case "GSI1":
		return "GSI_SK", nil
	case "GSI2":
		return "GSI2_SK", nil
	case "GSI3":
		return "GSI3_SK", nil
	default:
		return "", fmt.Errorf("unsupported index name: %s", indexName)
	}
}
