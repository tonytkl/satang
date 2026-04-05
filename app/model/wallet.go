package model

import "time"

// Wallet represents a wallet holding a currency balance.
// The wallet lives in its own partition to support multiple members.
// DynamoDB keys:
//
//	PK = "WALLET#<ID>"
//	SK = "#METADATA#<ID>"
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
		PK:        "WALLET#" + id,
		SK:        "#METADATA#" + id,
		ID:        id,
		OwnerID:   ownerID,
		Name:      name,
		Currency:  currency,
		Balance:   0,
		CreatedAt: now,
		UpdatedAt: now,
	}
}
