package model

import "time"

// TransactionType represents whether a transaction is a credit or debit.
type TransactionType string

const (
	TransactionTypeCredit TransactionType = "CREDIT"
	TransactionTypeDebit  TransactionType = "DEBIT"
)

// Transaction represents a financial transaction associated with a wallet.
// DynamoDB keys:
//
//	PK = "WALLET#<WalletID>"
//	SK = "TX#<Date>#<ID>"
//	GSI_PK = "TX_TYPE#<Type>"
//	GSI_SK = "TX#<Date>#<ID>"
type Transaction struct {
	PK          string          `dynamodbav:"PK"`
	SK          string          `dynamodbav:"SK"`
	GSIPK       string          `dynamodbav:"GSI_PK"`
	GSISK       string          `dynamodbav:"GSI_SK"`
	ID          string          `dynamodbav:"ID"`
	WalletID    string          `dynamodbav:"WalletID"`
	Type        TransactionType `dynamodbav:"Type"`
	Amount      float64         `dynamodbav:"Amount"`
	Currency    string          `dynamodbav:"Currency"`
	Category    string          `dynamodbav:"Category"`
	Description string          `dynamodbav:"Description"`
	Date        time.Time       `dynamodbav:"Date"`
	CreatedAt   time.Time       `dynamodbav:"CreatedAt"`
	UpdatedAt   time.Time       `dynamodbav:"UpdatedAt"`
}

func NewTransaction(id, walletID, category, description, currency string, txType TransactionType, amount float64, date time.Time) *Transaction {
	now := time.Now().UTC()
	dateStr := date.UTC().Format("2006-01-02")
	return &Transaction{
		PK:          "WALLET#" + walletID,
		SK:          "TX#" + dateStr + "#" + id,
		GSIPK:       "TX_TYPE#" + string(txType),
		GSISK:       "TX#" + dateStr + "#" + id,
		ID:          id,
		WalletID:    walletID,
		Type:        txType,
		Amount:      amount,
		Currency:    currency,
		Category:    category,
		Description: description,
		Date:        date.UTC(),
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}
