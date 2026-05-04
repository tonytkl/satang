package services

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/tonytkl/satang/model"
	"github.com/tonytkl/satang/repositories"
)

type TransactionService interface {
	CreateTransaction(ctx context.Context, walletID string, walletName string, categoryID string, categoryName string, description string, currency string, imageURL string, txType string, amount float64, date time.Time, ownerID string) error
}

type transactionService struct {
	repository repositories.TransactionRepository
}

func NewTransactionService(repository repositories.TransactionRepository) TransactionService {
	return &transactionService{
		repository: repository,
	}
}

func (service *transactionService) CreateTransaction(ctx context.Context, walletID string, walletName string, categoryID string, categoryName string, description string, currency string, imageURL string, txType string, amount float64, date time.Time, ownerID string) error {
	transactionType, err := getTransactionType(txType)
	if err != nil {
		return err
	}
	transaction := model.NewTransaction(
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
	if err := service.repository.Create(ctx, transaction); err != nil {
		return err
	}
	return nil
}

func getTransactionType(txType string) (model.TransactionType, error) {
	if strings.ToLower(txType) == "income" {
		return model.TransactionTypeIncome, nil
	}
	if strings.ToLower(txType) == "expense" {
		return model.TransactionTypeExpense, nil
	}
	if strings.ToLower(txType) == "transfer" {
		return model.TransactionTypeTransfer, nil
	}
	return "", errors.New("Valid transactiontion type is required")
}
