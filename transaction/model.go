package transaction

import (
	"time"
)

type TransactionType string
type Label string

const (
	TransactionIncome   TransactionType = "income"
	TransactionExpense  TransactionType = "expense"
	TransactionTransfer TransactionType = "transfer"
)

const (
	LabelNeed  Label = "need"
	LabelLife  Label = "life"
	LabelWaste Label = "waste"
)

type Wallet struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	UserID    int64     `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Category struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	UserID    int64     `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Tag struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	UserID    int64     `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type TransactionTag struct {
	TransactionID int64 `json:"transaction_id"`
	TagID         int64 `json:"tag_id"`
}

type Transaction struct {
	ID         int64           `json:"id"`
	Type       TransactionType `json:"type"`
	Amount     int64           `json:"amount"` // in smallest currency unit, e.g., cents
	WalletID   int64           `json:"wallet_id"`
	CategoryID int64           `json:"category_id"`
	Currency   string          `json:"currency"`
	Note       string          `json:"note"`
	Label      Label           `json:"label"`
	UserID     int64           `json:"user_id"`
	CreatedAt  time.Time       `json:"created_at"`
	UpdatedAt  time.Time       `json:"updated_at"`
}
