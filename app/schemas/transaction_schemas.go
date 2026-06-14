package schemas

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/tonytkl/satang/model"
)

type TransactionSchemas struct {
	ID           string                `json:"id"`
	WalletID     string                `json:"walletId"`
	WalletName   string                `json:"walletName,omitempty"`
	Type         model.TransactionType `json:"type"`
	Amount       float64               `json:"amount"`
	Currency     string                `json:"currency"`
	CategoryID   string                `json:"categoryID"`
	CategoryName string                `json:"categoryName,omitempty"`
	Description  *string               `json:"description,omitempty"`
	Date         time.Time             `json:"date"`
	ImageURL     *string               `json:"imageURL,omitempty"`
	OwnerID      string                `json:"ownerID"`
	CreatedAt    time.Time             `json:"createdAt"`
	UpdatedAt    time.Time             `json:"updatedAt"`
}

// BuildTransactionSchemas maps model transactions to response schemas and validates required fields.
func BuildTransactionSchemas(transactions []model.Transaction) ([]TransactionSchemas, error) {
	responseTransactions := make([]TransactionSchemas, 0, len(transactions))
	for _, tx := range transactions {
		schemaTransaction := TransactionSchemas{
			ID:           tx.ID,
			WalletID:     tx.WalletID,
			WalletName:   tx.WalletName,
			Type:         tx.Type,
			Amount:       tx.Amount,
			Currency:     tx.Currency,
			CategoryID:   tx.CategoryID,
			CategoryName: tx.CategoryName,
			Description:  tx.Description,
			Date:         tx.Date,
			ImageURL:     tx.ImageURL,
			OwnerID:      tx.OwnerID,
			CreatedAt:    tx.CreatedAt,
			UpdatedAt:    tx.UpdatedAt,
		}

		if err := validateTransactionSchema(schemaTransaction); err != nil {
			return nil, fmt.Errorf("transaction response schema validation failed for transaction ID %q: %w", tx.ID, err)
		}

		responseTransactions = append(responseTransactions, schemaTransaction)
	}

	return responseTransactions, nil
}

func validateTransactionSchema(transaction TransactionSchemas) error {
	if strings.TrimSpace(transaction.ID) == "" {
		return errors.New("id is required")
	}
	if strings.TrimSpace(transaction.WalletID) == "" {
		return errors.New("walletId is required")
	}
	if strings.TrimSpace(string(transaction.Type)) == "" {
		return errors.New("type is required")
	}
	if strings.TrimSpace(transaction.Currency) == "" {
		return errors.New("currency is required")
	}
	if strings.TrimSpace(transaction.CategoryID) == "" {
		return errors.New("categoryID is required")
	}
	if transaction.Date.IsZero() {
		return errors.New("date is required")
	}
	if strings.TrimSpace(transaction.OwnerID) == "" {
		return errors.New("ownerID is required")
	}

	return nil
}
