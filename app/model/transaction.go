package model

import (
	"time"

	"github.com/tonytkl/satang/utils"
)

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
//		GSI_ByCategoryPK = "TX_CATEGORY#<CATEGORY_ID>"
//		GSI_ByCategorySK = "TX#<Date>#<ID>"
//		GSI_ByWalletPK = "TX_WALLET#<WALLET_ID>"
//		GSI_ByWalletSK = "TX#<Date>#<ID>"
//	 	GSI_ByTransactionID = "TX_ID#<ID>"
type Transaction struct {
	PK                  string          `dynamodbav:"PK"`
	SK                  string          `dynamodbav:"SK"`
	GSI_ByCategoryPK    string          `dynamodbav:"GSI_PK"`
	GSI_ByCategorySK    string          `dynamodbav:"GSI_SK"`
	GSI_ByWalletPK      string          `dynamodbav:"GSI2_PK,omitempty"`
	GSI_ByWalletSK      string          `dynamodbav:"GSI2_SK,omitempty"`
	GSI_ByTransactionID string          `dynamodbav:"GSI3_PK,omitempty"`
	GSI_ByTransactionSK string          `dynamodbav:"GSI3_SK,omitempty"`
	ID                  string          `dynamodbav:"ID"`
	WalletID            string          `dynamodbav:"WalletID"`
	WalletName          string          `dynamodbav:"WalletName,omitempty"`
	Type                TransactionType `dynamodbav:"Type"`
	Amount              float64         `dynamodbav:"Amount"`
	Currency            string          `dynamodbav:"Currency"`
	CategoryID          string          `dynamodbav:"CategoryID"`
	CategoryName        string          `dynamodbav:"CategoryName,omitempty"`
	Description         *string         `dynamodbav:"Description,omitempty"`
	Date                time.Time       `dynamodbav:"Date"`
	ImageURL            *string         `dynamodbav:"ImageURL,omitempty"`
	OwnerID             string          `dynamodbav:"OwnerID"`
	CreatedAt           time.Time       `dynamodbav:"CreatedAt"`
	UpdatedAt           time.Time       `dynamodbav:"UpdatedAt"`
}

func NewTransaction(walletID string, walletName string, categoryID string, categoryName string, description string, currency string, imageURL string, txType TransactionType, amount float64, date time.Time, ownerID string) *Transaction {
	now := time.Now().UTC()
	dateStr := date.UTC().Format("2006-01-02")
	id := utils.GetUUID()
	return &Transaction{
		PK:                  "USER#" + ownerID,
		SK:                  "TX#" + dateStr + "#" + id,
		GSI_ByCategoryPK:    "TX_CATEGORY#" + string(categoryID),
		GSI_ByCategorySK:    "TX#" + dateStr + "#" + id,
		GSI_ByWalletPK:      "TX_WALLET#" + string(walletID),
		GSI_ByWalletSK:      "TX#" + dateStr + "#" + id,
		GSI_ByTransactionID: "TX_ID#" + id,
		GSI_ByTransactionSK: "TX#" + dateStr + "#" + id,
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
