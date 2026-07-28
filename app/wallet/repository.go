package wallet

import (
	"context"

	"github.com/tonytkl/satang/clients"
	"github.com/tonytkl/satang/repository"
)

type WalletRepository interface {
	CreateWallet(ctx context.Context, wallet *Wallet) error
	GetWalletList(ctx context.Context, ownerID string, nextToken string, limit int32) ([]*Wallet, string, error)
	GetWallet(ctx context.Context, ownerID string, walletID string) (*Wallet, error)
}

type walletRepository struct {
	db             clients.DynamoDBClient
	tableName      string
	baseRepository repository.BaseRepository[*Wallet]
}

func NewWalletRepository(db clients.DynamoDBClient, tableName string) WalletRepository {
	return &walletRepository{
		db:             db,
		tableName:      tableName,
		baseRepository: repository.NewBaseRepository(db, tableName, "WALLET", func() *Wallet { return &Wallet{} }),
	}
}

func (walletRepository *walletRepository) CreateWallet(ctx context.Context, wallet *Wallet) error {
	return walletRepository.baseRepository.Save(ctx, wallet)
}

func (walletRepository *walletRepository) GetWalletList(ctx context.Context, ownerID string, nextToken string, limit int32) ([]*Wallet, string, error) {
	return walletRepository.baseRepository.List(ctx, ownerID, nextToken, limit)
}

func (walletRepository *walletRepository) GetWallet(ctx context.Context, ownerID string, walletID string) (*Wallet, error) {
	return walletRepository.baseRepository.Get(ctx, ownerID, walletID)
}
