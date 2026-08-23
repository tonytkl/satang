package wallet

import "time"

type WalletType string

const (
	WalletTypeDebit      WalletType = "DEBIT"
	WalletTypeCredit     WalletType = "CREDIT"
	WalletTypeInvestment WalletType = "INVESTMENT"
)

// Wallet represents a wallet holding a currency balance.
// DynamoDB keys:
//
//	PK = "USER#<OwnerID>"
//	SK = "WALLET#<ID>"
type Wallet struct {
	PK        string     `dynamodbav:"PK"`
	SK        string     `dynamodbav:"SK"`
	ID        string     `dynamodbav:"ID"`
	OwnerID   string     `dynamodbav:"OwnerID"`
	Name      string     `dynamodbav:"Name"`
	Currency  string     `dynamodbav:"Currency"`
	Balance   float64    `dynamodbav:"Balance"`
	Type      WalletType `dynamodbav:"Type"`
	IsActive  bool       `dynamodbav:"IsActive"`
	CreatedAt time.Time  `dynamodbav:"CreatedAt"`
	UpdatedAt time.Time  `dynamodbav:"UpdatedAt"`
}

func NewWallet(id string, ownerID string, name string, currency string, walletType WalletType) *Wallet {
	now := time.Now().UTC()
	return &Wallet{
		PK:        "USER#" + ownerID,
		SK:        "WALLET#" + id,
		ID:        id,
		OwnerID:   ownerID,
		Name:      name,
		Currency:  currency,
		Balance:   0,
		Type:      walletType,
		IsActive:  true,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func (wallet *Wallet) GetID() string {
	return wallet.ID
}

func (wallet *Wallet) SetPK(pk string) {
	wallet.PK = pk
}

func (wallet *Wallet) SetSK(sk string) {
	wallet.SK = sk
}

func (wallet *Wallet) GetCreatedAt() time.Time {
	return wallet.CreatedAt
}

func (wallet *Wallet) SetCreatedAt(createdAt time.Time) {
	wallet.CreatedAt = createdAt
}

func (wallet *Wallet) GetUpdatedAt() time.Time {
	return wallet.UpdatedAt
}

func (wallet *Wallet) SetUpdatedAt(updatedAt time.Time) {
	wallet.UpdatedAt = updatedAt
}

func (wallet *Wallet) GetOwnerID() string {
	return wallet.OwnerID
}
