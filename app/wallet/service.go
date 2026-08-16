package wallet

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/tonytkl/satang/transaction"
	"github.com/tonytkl/satang/utils"
)

type Service interface {
	CreateWallet(ctx context.Context, ownerID string, name string, currency string, balance float64, strWalletType string) error
	GetWalletList(ctx context.Context, ownerID string, nextToken string, limit int32) ([]*Wallet, string, error)
	GetWallet(ctx context.Context, ownerID string, walletID string) (*Wallet, error)
	EditWallet(ctx context.Context, ownerID string, walletID string, changedFields map[string]any) error
	SetActiveWallet(ctx context.Context, ownerID string, walletID string, isActive bool) error
}

type service struct {
	repository         WalletRepository
	transactionService transaction.TransactionService
}

func NewService(repository WalletRepository, transactionService transaction.TransactionService) Service {
	return &service{
		repository:         repository,
		transactionService: transactionService,
	}
}

func (service *service) CreateWallet(ctx context.Context, ownerID string, name string, currency string, balance float64, strWalletType string) error {
	// TODO: Implement currency query from user model
	const defaultCurrency = "THB"
	if currency == "" {
		currency = defaultCurrency
	}
	walletID := utils.GetUUID()
	walletType, err := getTransactionType(strWalletType)
	if err != nil {
		return err
	}
	wallet := NewWallet(
		walletID,
		ownerID,
		name,
		currency,
		walletType,
	)
	wallet.Balance = balance

	if balance != 0 {
		if err := service.transactionService.CreateTransaction(
			ctx,
			wallet.ID,
			wallet.Name,
			// TODO: Query actual category ID
			"cat01",
			"Initial balance",
			"",
			wallet.Currency,
			"",
			string(transaction.TransactionTypeIncome),
			wallet.Balance,
			time.Now().UTC(),
			wallet.OwnerID,
		); err != nil {
			return err
		}
	}

	return service.repository.CreateWallet(ctx, wallet)
}

func (service *service) GetWalletList(ctx context.Context, ownerID string, nextToken string, limit int32) ([]*Wallet, string, error) {
	if limit == 0 {
		limit = utils.DEFAULT_PAGINATION_SIZE
	}
	return service.repository.GetWalletList(
		ctx,
		ownerID,
		nextToken,
		limit,
	)
}

func (service *service) GetWallet(ctx context.Context, ownerID string, walletID string) (*Wallet, error) {
	return service.repository.GetWallet(
		ctx,
		ownerID,
		walletID,
	)
}

func (service *service) EditWallet(ctx context.Context, ownerID string, walletID string, changedFields map[string]any) error {
	if changedFields["OwnerID"] != nil {
		return errors.New("Owner ID is not updateable")
	}

	if changedFields["Currency"] != nil {
		return errors.New("Currency is not updateable")
	}

	if changedFields["Balance"] != nil {
		return errors.New("Balance is not updateable")
	}
	return service.repository.EditWallet(ctx, ownerID, walletID, changedFields)
}

func (service *service) SetActiveWallet(ctx context.Context, ownerID string, walletID string, isActive bool) error {
	changedFields := make(map[string]any, 1)
	changedFields["IsActive"] = isActive
	return service.EditWallet(ctx, ownerID, walletID, changedFields)
}

// TODO: Implement delete wallet. Need to define how to do with existing transaction
// func (service *service) DeleteWallet(ctx context.Context, ownerID string, walletID string) error {

// }

func getTransactionType(strWalletType string) (WalletType, error) {
	if strings.ToLower(strWalletType) == "debit" {
		return WalletTypeDebit, nil
	}

	if strings.ToLower(strWalletType) == "credit" {
		return WalletTypeCredit, nil
	}

	if strings.ToLower(strWalletType) == "investment" {
		return WalletTypeInvestment, nil
	}

	return "", errors.New("Invalid transaction type")
}
