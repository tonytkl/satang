package wallet

import (
	"context"

	"github.com/tonytkl/satang/clients"
	"github.com/tonytkl/satang/repository"
)

type Repository interface {
	CreateWallet(ctx context.Context, wallet Wallet) error
	ListWallets(ctx context.Context, ownerID string, nextToken string, limit int32) ([]Wallet, string, error)
	GetWallet(ctx context.Context, ownerID string, walletID string) (Wallet, error)
	EditWallet(ctx context.Context, ownerID string, walletID string, changedFields map[string]any) error
	DeleteWallet(ctx context.Context, ownerID string, walletID string) error
}

type walletRepository struct {
	db             clients.DynamoDBClient
	tableName      string
	baseRepository repository.BaseRepository[*Wallet]
}

func NewRepository(db clients.DynamoDBClient, tableName string) Repository {
	return &walletRepository{
		db:             db,
		tableName:      tableName,
		baseRepository: repository.NewBaseRepository(db, tableName, "WALLET", func() *Wallet { return &Wallet{} }),
	}
}

func (walletRepository *walletRepository) CreateWallet(ctx context.Context, wallet Wallet) error {
	return walletRepository.baseRepository.Save(ctx, &wallet)
}

func (walletRepository *walletRepository) ListWallets(ctx context.Context, ownerID string, nextToken string, limit int32) ([]Wallet, string, error) {
	items, encodedNextToken, err := walletRepository.baseRepository.List(ctx, ownerID, nextToken, limit)
	if err != nil {
		return nil, "", err
	}

	wallets := make([]Wallet, 0, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}
		wallets = append(wallets, *item)
	}

	return wallets, encodedNextToken, nil
}

func (walletRepository *walletRepository) GetWallet(ctx context.Context, ownerID string, walletID string) (Wallet, error) {
	wallet, err := walletRepository.baseRepository.Get(ctx, ownerID, walletID)
	if err != nil {
		return Wallet{}, err
	}
	return *wallet, nil
}

func (walletRepository *walletRepository) EditWallet(ctx context.Context, ownerID string, walletID string, changedFields map[string]any) error {
	return walletRepository.baseRepository.Update(ctx, ownerID, walletID, changedFields)
}

func (walletRepository *walletRepository) DeleteWallet(ctx context.Context, ownerID string, walletID string) error {
	return walletRepository.baseRepository.Delete(ctx, ownerID, walletID)
}
