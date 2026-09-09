package transaction

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/tonytkl/satang/utils"
)

type TransactionService interface {
	CreateTransaction(ctx context.Context, walletID string, walletName string, categoryID string, categoryName string, description string, currency string, imageURL string, txType string, amount float64, date time.Time, ownerID string) error
	GetTransaction(ctx context.Context, transactionID string, ownerID string) (*Transaction, error)
	ListTransactions(ctx context.Context, ownerID string, fromDate time.Time, toDate time.Time, limit int32, nextToken string) ([]Transaction, string, error)
	ListTransactionsOfCategory(ctx context.Context, ownerID string, fromDate time.Time, toDate time.Time, limit int32, nextToken string, categoryID string) ([]Transaction, string, error)
	ListTransactionsOfWallet(ctx context.Context, ownerID string, fromDate time.Time, toDate time.Time, limit int32, nextToken string, wallet string) ([]Transaction, string, error)
	EditTransaction(ctx context.Context, ownerID string, transactionID string, changedFields map[string]any) error
	DeleteTransaction(ctx context.Context, ownerID string, transactionID string) error
}

type transactionService struct {
	repository TransactionRepository
}

func NewTransactionService(repository TransactionRepository) TransactionService {
	return &transactionService{
		repository: repository,
	}
}

func (service *transactionService) CreateTransaction(ctx context.Context, walletID string, walletName string, categoryID string, categoryName string, description string, currency string, imageURL string, txType string, amount float64, date time.Time, ownerID string) error {
	transactionType, err := getTransactionType(txType)
	if err != nil {
		return err
	}
	transaction := NewTransaction(
		walletID,
		walletName,
		categoryID,
		categoryName,
		description,
		currency,
		imageURL,
		transactionType,
		amount,
		date,
		ownerID,
	)
	if err := validateTransaction(transaction); err != nil {
		return err
	}
	if err := service.repository.CreateTransaction(ctx, transaction); err != nil {
		return err
	}
	return nil
}

func (service *transactionService) GetTransaction(ctx context.Context, transactionID string, ownerID string) (*Transaction, error) {
	if transactionID == "" {
		return nil, errors.New("Transaction ID is required")
	}

	if ownerID == "" {
		return nil, errors.New("Owner ID is required")
	}

	return service.repository.GetTransaction(ctx, ownerID, transactionID)
}

func (service *transactionService) ListTransactions(ctx context.Context, ownerID string, fromDate time.Time, toDate time.Time, limit int32, nextToken string) ([]Transaction, string, error) {
	if ownerID == "" {
		return nil, "", errors.New("owner ID is required")
	}
	if toDate.IsZero() {
		// Default to today (UTC)
		toDate = time.Now().UTC()
	} else {
		toDate = toDate.UTC()
	}
	if fromDate.IsZero() {
		// Default to 7 days backward from the end of the range (UTC)
		fromDate = toDate.AddDate(0, 0, -7)
	} else {
		fromDate = fromDate.UTC()
	}
	if fromDate.After(toDate) {
		return nil, "", errors.New("from date must not be after to date")
	}
	if limit < 0 {
		return nil, "", errors.New("limit must be greater than or equal to 0")
	}

	// Default pagination size if not set
	if limit == 0 {
		limit = utils.DEFAULT_PAGINATION_SIZE
	}

	return service.repository.ListTransactionsOfSubModel(ctx, "date", "", ownerID, fromDate, toDate, nextToken, limit)
}

func (service *transactionService) ListTransactionsOfCategory(ctx context.Context, ownerID string, fromDate time.Time, toDate time.Time, limit int32, nextToken string, categoryID string) ([]Transaction, string, error) {
	if ownerID == "" {
		return nil, "", errors.New("owner ID is required")
	}
	if strings.TrimSpace(categoryID) == "" {
		return nil, "", errors.New("category ID is required")
	}
	if toDate.IsZero() {
		toDate = time.Now().UTC()
	} else {
		toDate = toDate.UTC()
	}
	if fromDate.IsZero() {
		fromDate = toDate.AddDate(0, 0, -7)
	} else {
		fromDate = fromDate.UTC()
	}
	if fromDate.After(toDate) {
		return nil, "", errors.New("from date must not be after to date")
	}
	if limit < 0 {
		return nil, "", errors.New("limit must be greater than or equal to 0")
	}

	// Default pagination size if not set
	if limit == 0 {
		limit = utils.DEFAULT_PAGINATION_SIZE
	}

	return service.repository.ListTransactionsOfSubModel(ctx, "category", categoryID, ownerID, fromDate, toDate, nextToken, limit)
}

func (service *transactionService) ListTransactionsOfWallet(ctx context.Context, ownerID string, fromDate time.Time, toDate time.Time, limit int32, nextToken string, walletID string) ([]Transaction, string, error) {
	if ownerID == "" {
		return nil, "", errors.New("owner ID is required")
	}
	if strings.TrimSpace(walletID) == "" {
		return nil, "", errors.New("wallet ID is required")
	}
	if toDate.IsZero() {
		toDate = time.Now().UTC()
	} else {
		toDate = toDate.UTC()
	}
	if fromDate.IsZero() {
		fromDate = toDate.AddDate(0, 0, -7)
	} else {
		fromDate = fromDate.UTC()
	}
	if fromDate.After(toDate) {
		return nil, "", errors.New("from date must not be after to date")
	}
	if limit < 0 {
		return nil, "", errors.New("limit must be greater than or equal to 0")
	}

	// Default pagination size if not set
	if limit == 0 {
		limit = utils.DEFAULT_PAGINATION_SIZE
	}

	return service.repository.ListTransactionsOfSubModel(ctx, "wallet", walletID, ownerID, fromDate, toDate, nextToken, limit)
}

func (service *transactionService) EditTransaction(ctx context.Context, ownerID string, transactionID string, changedFields map[string]any) error {
	if changedFields == nil {
		return errors.New("Update payload is required")
	}

	normalizedFields := make(map[string]any, len(changedFields))
	for key, value := range changedFields {
		normalizedFields[key] = value
	}

	if _, ok := normalizedFields["OwnerID"]; ok {
		return errors.New("Owner ID is not updateable")
	}

	if typeValue, ok := normalizedFields["Type"]; ok {
		strTransactionType, ok := typeValue.(string)
		if !ok {
			return errors.New("Type must be a string")
		}

		categoryType, err := getTransactionType(strTransactionType)
		if err != nil {
			return err
		}
		normalizedFields["Type"] = categoryType
	}

	return service.repository.EditTransaction(ctx, ownerID, transactionID, normalizedFields)
}

func (service *transactionService) DeleteTransaction(ctx context.Context, ownerID string, transactionID string) error {
	return service.repository.DeleteTransaction(ctx, ownerID, transactionID)
}

func getTransactionType(txType string) (TransactionType, error) {
	if strings.ToLower(txType) == "income" {
		return TransactionTypeIncome, nil
	}
	if strings.ToLower(txType) == "expense" {
		return TransactionTypeExpense, nil
	}
	if strings.ToLower(txType) == "transfer" {
		return TransactionTypeTransfer, nil
	}
	return "", errors.New("Valid transactiontion type is required")
}

// validateTransaction ensures the required transaction fields are present.
func validateTransaction(transaction *Transaction) error {
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
