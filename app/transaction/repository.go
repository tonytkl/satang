package transaction

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/tonytkl/satang/clients"
	"github.com/tonytkl/satang/repository"
	"github.com/tonytkl/satang/utils"
)

// ErrTransactionNotFound is returned when no transaction matches the query.
var ErrTransactionNotFound = errors.New("transaction not found")

// Repository defines persistence operations for transactions.
type Repository interface {
	CreateTransaction(ctx context.Context, transaction *Transaction) error
	GetTransaction(ctx context.Context, ownerID string, transactionID string) (*Transaction, error)
	EditTransaction(ctx context.Context, ownerID string, transactionID string, changedFields map[string]any) error
	DeleteTransaction(ctx context.Context, ownerID string, transactionID string) error
	ListTransactionsOfSubModel(ctx context.Context, subModelName string, targetID string, ownerID string, fromDate time.Time, toDate time.Time, nextToken string, limit int32) ([]Transaction, string, error)
}

type transactionRepository struct {
	db             clients.DynamoDBClient
	tableName      string
	baseRepository repository.BaseRepository[*Transaction]
}

// NewRepository creates a transaction repository backed by DynamoDB.
func NewRepository(db clients.DynamoDBClient, tableName string) Repository {
	return &transactionRepository{
		db:             db,
		tableName:      tableName,
		baseRepository: repository.NewBaseRepository(db, tableName, "TX", func() *Transaction { return &Transaction{} }),
	}
}

// Create stores a transaction and populates its derived keys and timestamps.
// Not using base repository because of GSIs
func (repository *transactionRepository) CreateTransaction(ctx context.Context, transaction *Transaction) error {
	sortingKey := utils.GetPartitionKeyWithDate("TX", transaction.Date, transaction.ID)

	transaction.PK = utils.GetPartitionKey("USER", transaction.OwnerID)
	transaction.SK = utils.GetPartitionKey("TX", transaction.ID)

	transaction.GSI_ByDatePK = utils.GetPartitionKey("USER", transaction.OwnerID)
	transaction.GSI_ByDateSK = sortingKey

	transaction.GSI_ByCategoryPK = utils.GetPartitionKeySubModel("USER", transaction.OwnerID, "TX_CATEGORY", transaction.CategoryID)
	transaction.GSI_ByCategorySK = sortingKey

	transaction.GSI_ByWalletPK = utils.GetPartitionKeySubModel("USER", transaction.OwnerID, "TX_WALLET", transaction.WalletID)
	transaction.GSI_ByWalletSK = sortingKey

	if transaction.CreatedAt.IsZero() {
		transaction.SetCreatedAt(time.Now().UTC())
	}
	if transaction.UpdatedAt.IsZero() {
		transaction.SetUpdatedAt(transaction.CreatedAt)
	}

	err := repository.db.PutItem(ctx, repository.tableName, transaction)
	if err != nil {
		return err
	}

	return nil
}

func (repository *transactionRepository) GetTransaction(ctx context.Context, ownerID string, transactionID string) (*Transaction, error) {
	return repository.baseRepository.Get(ctx, ownerID, transactionID)
}

func (repository *transactionRepository) EditTransaction(ctx context.Context, ownerID string, transactionID string, changedFields map[string]any) error {
	updatedFields := make(map[string]any, len(changedFields)+5)
	for key, value := range changedFields {
		updatedFields[key] = value
	}

	if walletID, ok := updatedFields["WalletID"].(string); ok {
		updatedFields["GSI_ByWalletPK"] = utils.GetPartitionKeySubModel("USER", ownerID, "TX_WALLET", walletID)
	}

	if categoryID, ok := updatedFields["CategoryID"].(string); ok {
		updatedFields["GSI_ByCategoryPK"] = utils.GetPartitionKeySubModel("USER", ownerID, "TX_CATEGORY", categoryID)
	}

	dateValue, hasDate := updatedFields["Date"]
	if hasDate {
		transactionDate, ok := dateValue.(time.Time)
		if !ok {
			return errors.New("Date must be time.Time")
		}

		sortingKey := utils.GetPartitionKeyWithDate("TX", transactionDate, transactionID)
		updatedFields["GSI_ByDateSK"] = sortingKey
		updatedFields["GSI_ByCategorySK"] = sortingKey
		updatedFields["GSI_ByWalletSK"] = sortingKey
	}

	return repository.baseRepository.Update(ctx, ownerID, transactionID, updatedFields)
}

func (repository *transactionRepository) DeleteTransaction(ctx context.Context, ownerID string, transactionID string) error {
	return repository.baseRepository.Delete(ctx, ownerID, transactionID)
}

// ListByGSI lists transactions using the provided GSI name and partition key prefix.
func (repository *transactionRepository) ListTransactionsOfSubModel(ctx context.Context, subModelName string, targetID string, ownerID string, fromDate time.Time, toDate time.Time, nextToken string, limit int32) ([]Transaction, string, error) {
	// Getting partition keys based on models
	indexName, indexPartitionKey, indexSortingKey, err := getIndexPartitionKeyAndSortingKey(subModelName)
	if err != nil {
		return nil, "", err
	}

	// Forming datetime config
	if fromDate.After(toDate) {
		return nil, "", errors.New("from date must not be after to date")
	}
	fromSK := utils.GetPartitionKeyWithDate("TX", fromDate, "")
	toSK := utils.GetPartitionKeyWithDate("TX", toDate, "")

	// Forming query expression
	queryExpression := indexPartitionKey + " = :indexPK AND " + indexSortingKey + " BETWEEN :from AND :to"
	var expressionValues map[string]any
	if strings.ToUpper(subModelName) == "DATE" {
		expressionValues = map[string]any{
			":indexPK": utils.GetPartitionKey("USER", ownerID),
			":from":    fromSK,
			":to":      toSK,
		}
	} else {
		expressionValues = map[string]any{
			":indexPK": utils.GetPartitionKeySubModel("USER", ownerID, "TX_"+strings.ToUpper(subModelName), targetID),
			":from":    fromSK,
			":to":      toSK,
		}
	}

	transactions := []Transaction{}

	encodedNextToken, err := repository.db.QueryItemsWithPagination(
		ctx,
		repository.tableName,
		queryExpression,
		expressionValues,
		indexName,
		"", // no filter
		limit,
		nextToken,
		&transactions,
	)

	if err != nil {
		return nil, "", fmt.Errorf("query transaction by ID: %w", err)
	}

	return transactions, encodedNextToken, nil
}

// getIndexPartitionKeyAndSortingKey resolves a submodel name to its DynamoDB GSI name and key attributes.
func getIndexPartitionKeyAndSortingKey(model string) (string, string, string, error) {
	switch model {
	case "date":
		return "GSI1", "GSI1_PK", "GSI1_SK", nil
	case "wallet":
		return "GSI2", "GSI2_PK", "GSI2_SK", nil
	case "category":
		return "GSI3", "GSI3_PK", "GSI3_SK", nil
	default:
		return "", "", "", fmt.Errorf("unsupported model name: %s", model)
	}
}
