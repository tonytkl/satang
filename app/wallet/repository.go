package wallet

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/tonytkl/satang/clients"
	"github.com/tonytkl/satang/utils"
)

type WalletRepository interface {
	Create(ctx context.Context, wallet *Wallet) error
	GetWalletList(ctx context.Context, ownerID string) ([]Wallet, error)
	GetWallet(ctx context.Context, ownerID string, walletID string) (*Wallet, error)
}

type walletRepository struct {
	db        clients.DynamoDBClient
	tableName string
}

func NewWalletRepository(db clients.DynamoDBClient, tableName string) WalletRepository {
	return &walletRepository{
		db:        db,
		tableName: tableName,
	}
}

func (repository *walletRepository) Create(ctx context.Context, wallet *Wallet) error {
	wallet.PK = utils.GetPartitionKey("USER", wallet.OwnerID)
	wallet.SK = utils.GetPartitionKey("WALLET", wallet.ID)

	if wallet.CreatedAt.IsZero() {
		wallet.CreatedAt = time.Now().UTC()
	}

	if wallet.UpdatedAt.IsZero() {
		wallet.UpdatedAt = time.Now().UTC()
	}

	err := repository.db.PutItem(ctx, repository.tableName, wallet)

	if err != nil {
		return err
	}

	return nil
}

func (repository *walletRepository) GetWalletList(ctx context.Context, ownerID string) ([]Wallet, error) {
	if ownerID == "" {
		return nil, errors.New("owner ID is required")
	}

	queryExpression := "PK = :ownerPK AND begins_with(SK, :walletSKPrefix)"
	experessionValues := map[string]any{
		":ownerPK":        utils.GetPartitionKey("USER", ownerID),
		":walletSKPrefix": "WALLET#",
	}

	wallets := []Wallet{}

	err := repository.db.QueryItems(
		ctx,
		repository.tableName,
		queryExpression,
		experessionValues,
		"",
		"",
		&wallets,
	)

	if err != nil {
		return nil, fmt.Errorf("Errors on GetWalletList: %w", err)
	}

	return wallets, nil
}

func (repository *walletRepository) GetWallet(ctx context.Context, ownerID string, walletID string) (*Wallet, error) {
	if ownerID == "" {
		return nil, errors.New("owner ID is required")
	}

	if walletID == "" {
		return nil, errors.New("wallet ID is required")
	}

	key := map[string]any{
		":ownerPK":  utils.GetPartitionKey("USER", ownerID),
		":walletSK": utils.GetPartitionKey("WALLET", walletID),
	}

	wallet := Wallet{}

	err := repository.db.GetItem(
		ctx,
		repository.tableName,
		key,
		&wallet,
	)

	if err != nil {
		return nil, fmt.Errorf("Errors on GetWallet: %w", err)
	}

	return &wallet, nil
}
