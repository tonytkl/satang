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
//		PK = "USER#<UserID>"
//		SK = "TX#<Date>#<ID>"
//		GSI_ByCategoryPK = "TX_CATEGORY#<CATEGORY>"
//		GSI_ByTransactionTypeSK = "TX#<Date>#<ID>"
//	 	GSI_ByTransactionID = "TX_ID#<ID>"
type Transaction struct {
	PK                  string          `dynamodbav:"PK"`
	SK                  string          `dynamodbav:"SK"`
	GSI_ByCategoryPK    string          `dynamodbav:"GSI_PK"`
	GSI_ByCategorySK    string          `dynamodbav:"GSI_SK"`
	GSI_ByTransactionID string          `dynamodbav:"GSI2_PK"`
	ID                  string          `dynamodbav:"ID"`
	WalletID            string          `dynamodbav:"WalletID"`
	WalletName          string          `dynamodbav:"WalletName,omitempty"`
	Type                TransactionType `dynamodbav:"Type"`
	Amount              float64         `dynamodbav:"Amount"`
	Currency            string          `dynamodbav:"Currency"`
	CategoryID          string          `dynamodbav:"CategoryID"`
	CategoryName        string          `dynamodbav:"CategoryName,omitempty"`
	Description         *string         `dynamodbav:"Description"`
	Date                time.Time       `dynamodbav:"Date"`
	ImageURL            *string         `dynamodbav:"ImageURL,omitempty"`
	OwnerID             string          `dynamodbav:"OwnerID"`
	CreatedAt           time.Time       `dynamodbav:"CreatedAt"`
	UpdatedAt           time.Time       `dynamodbav:"UpdatedAt"`
}

func NewTransaction(id, walletID, walletName, categoryID, categoryName, description, currency, imageURL string, txType TransactionType, amount float64, date time.Time, ownerID string) *Transaction {
	now := time.Now().UTC()
	dateStr := date.UTC().Format("2006-01-02")
	return &Transaction{
		PK:                  "USER#" + walletID,
		SK:                  "TX#" + dateStr + "#" + id,
		GSI_ByCategoryPK:    "TX_CATEGORY#" + string(categoryID),
		GSI_ByCategorySK:    "TX#" + dateStr + "#" + id,
		GSI_ByTransactionID: "TX_ID#" + id,
		ID:                  id,
		WalletID:            walletID,
		WalletName:          walletName,
		Type:                txType,
		Amount:              amount,
		Currency:            currency,
		CategoryID:          categoryID,
		CategoryName:        categoryName,
		Description:         &description,
		Date:                date.UTC(),
		ImageURL:            &imageURL,
		OwnerID:             ownerID,
		CreatedAt:           now,
		UpdatedAt:           now,
	}
}
