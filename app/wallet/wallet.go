package wallet

import "time"

// Wallet represents a wallet holding a currency balance.
// The wallet lives in its own partition to support multiple members.
// DynamoDB keys:
//
//	PK = "USER#<OwnerID>"
//	SK = "WALLET#<ID>"
//
// Membership is modelled separately via WalletMember (adjacency list pattern).
type Wallet struct {
	PK        string    `dynamodbav:"PK"`
	SK        string    `dynamodbav:"SK"`
	ID        string    `dynamodbav:"ID"`
	OwnerID   string    `dynamodbav:"OwnerID"`
	Name      string    `dynamodbav:"Name"`
	Currency  string    `dynamodbav:"Currency"`
	Balance   float64   `dynamodbav:"Balance"`
	CreatedAt time.Time `dynamodbav:"CreatedAt"`
	UpdatedAt time.Time `dynamodbav:"UpdatedAt"`
}

func NewWallet(id, ownerID, name, currency string) *Wallet {
	now := time.Now().UTC()
	return &Wallet{
		PK:        "USER#" + ownerID,
		SK:        "WALLET#" + id,
		ID:        id,
		OwnerID:   ownerID,
		Name:      name,
		Currency:  currency,
		Balance:   0,
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
