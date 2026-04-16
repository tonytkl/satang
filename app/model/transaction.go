package model

import "time"

// TransactionType represents whether a transaction is an income, expense, or transfer.
type TransactionType string

const (
	TransactionTypeIncome   TransactionType = "INCOME"
	TransactionTypeExpense  TransactionType = "EXPENSE"
	TransactionTypeTransfer TransactionType = "TRANSFER"
)

// Transaction represents a financial transaction associated with a wallet.
// DynamoDB keys:
//
//	PK = "WALLET#<WalletID>"
//	SK = "TX#<Date>#<ID>"
//	GSI_PK = "TX_TYPE#<Type>"
//	GSI_SK = "TX#<Date>#<ID>"
type Transaction struct {
	PK           string          `dynamodbav:"PK"`
	SK           string          `dynamodbav:"SK"`
	GSIPK        string          `dynamodbav:"GSI_PK"`
	GSISK        string          `dynamodbav:"GSI_SK"`
	ID           string          `dynamodbav:"ID"`
	WalletID     string          `dynamodbav:"WalletID"`
	WalletName   string          `dynamodbav:"WalletName,omitempty"`
	Type         TransactionType `dynamodbav:"Type"`
	Amount       float64         `dynamodbav:"Amount"`
	Currency     string          `dynamodbav:"Currency"`
	CategoryID   string          `dynamodbav:"CategoryID"`
	CategoryName string          `dynamodbav:"CategoryName,omitempty"`
	Description  string          `dynamodbav:"Description"`
	Date         time.Time       `dynamodbav:"Date"`
	ImageURL     string          `dynamodbav:"ImageURL,omitempty"`
	CreatedAt    time.Time       `dynamodbav:"CreatedAt"`
	UpdatedAt    time.Time       `dynamodbav:"UpdatedAt"`
}

func NewTransaction(id, walletID, walletName, categoryID, categoryName, description, currency, imageURL string, txType TransactionType, amount float64, date time.Time) *Transaction {
	now := time.Now().UTC()
	dateStr := date.UTC().Format("2006-01-02")
	return &Transaction{
		PK:           "WALLET#" + walletID,
		SK:           "TX#" + dateStr + "#" + id,
		GSIPK:        "TX_TYPE#" + string(txType),
		GSISK:        "TX#" + dateStr + "#" + id,
		ID:           id,
		WalletID:     walletID,
		WalletName:   walletName,
		Type:         txType,
		Amount:       amount,
		Currency:     currency,
		CategoryID:   categoryID,
		CategoryName: categoryName,
		Description:  description,
		Date:         date.UTC(),
		ImageURL:     imageURL,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}
