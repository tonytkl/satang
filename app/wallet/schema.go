package wallet

import "time"

type WalletRead struct {
	ID         string     `json:"id"`
	Name       string     `json:"name"`
	WalletType WalletType `json:"type"`
	Currency   string     `json:"currency"`
	OwnerID    string     `json:"ownerID"`
	CreatedAt  time.Time  `json:"createdAt"`
	UpdatedAt  time.Time  `json:"updatedAt"`
}

func BuildWalletRead(wallets []Wallet) []WalletRead {
	walletResponse := make([]WalletRead, 0, len(wallets))
	for _, wallet := range wallets {
		schemaWallet := WalletRead{
			ID:         wallet.ID,
			Name:       wallet.Name,
			WalletType: wallet.Type,
			Currency:   wallet.Currency,
			OwnerID:    wallet.OwnerID,
			CreatedAt:  wallet.CreatedAt,
			UpdatedAt:  wallet.UpdatedAt,
		}
		walletResponse = append(walletResponse, schemaWallet)
	}
	return walletResponse
}
