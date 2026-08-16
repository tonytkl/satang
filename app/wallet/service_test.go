package wallet

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tonytkl/satang/transaction"
)

type mockWalletRepository struct {
	createWalletFn  func(ctx context.Context, wallet *Wallet) error
	getWalletListFn func(ctx context.Context, ownerID string, nextToken string, limit int32) ([]*Wallet, string, error)
	getWalletFn     func(ctx context.Context, ownerID string, walletID string) (*Wallet, error)
	editWalletFn    func(ctx context.Context, ownerID string, walletID string, changedFields map[string]any) error
	deleteWalletFn  func(ctx context.Context, ownerID string, walletID string) error
}

var _ WalletRepository = (*mockWalletRepository)(nil)

func (m *mockWalletRepository) CreateWallet(ctx context.Context, wallet *Wallet) error {
	if m.createWalletFn != nil {
		return m.createWalletFn(ctx, wallet)
	}
	return nil
}

func (m *mockWalletRepository) GetWalletList(ctx context.Context, ownerID string, nextToken string, limit int32) ([]*Wallet, string, error) {
	if m.getWalletListFn != nil {
		return m.getWalletListFn(ctx, ownerID, nextToken, limit)
	}
	return nil, "", nil
}

func (m *mockWalletRepository) GetWallet(ctx context.Context, ownerID string, walletID string) (*Wallet, error) {
	if m.getWalletFn != nil {
		return m.getWalletFn(ctx, ownerID, walletID)
	}
	return nil, nil
}

func (m *mockWalletRepository) EditWallet(ctx context.Context, ownerID string, walletID string, changedFields map[string]any) error {
	if m.editWalletFn != nil {
		return m.editWalletFn(ctx, ownerID, walletID, changedFields)
	}
	return nil
}

func (m *mockWalletRepository) DeleteWallet(ctx context.Context, ownerID string, walletID string) error {
	if m.deleteWalletFn != nil {
		return m.deleteWalletFn(ctx, ownerID, walletID)
	}
	return nil
}

type mockTransactionService struct {
	createTransactionFn func(ctx context.Context, walletID string, walletName string, categoryID string, categoryName string, description string, currency string, imageURL string, txType string, amount float64, date time.Time, ownerID string) error
}

var _ transaction.TransactionService = (*mockTransactionService)(nil)

func (m *mockTransactionService) CreateTransaction(ctx context.Context, walletID string, walletName string, categoryID string, categoryName string, description string, currency string, imageURL string, txType string, amount float64, date time.Time, ownerID string) error {
	if m.createTransactionFn != nil {
		return m.createTransactionFn(ctx, walletID, walletName, categoryID, categoryName, description, currency, imageURL, txType, amount, date, ownerID)
	}
	return nil
}

func (m *mockTransactionService) GetTransaction(ctx context.Context, transactionID string, ownerID string) (*transaction.Transaction, error) {
	return nil, nil
}

func (m *mockTransactionService) GetTransactionsBetweenPeriod(ctx context.Context, ownerID string, fromDate time.Time, toDate time.Time, limit int32, nextToken string) ([]transaction.Transaction, string, error) {
	return nil, "", nil
}

func TestCreateWalletUsesDefaultCurrencyAndInitialBalance(t *testing.T) {
	repo := &mockWalletRepository{
		createWalletFn: func(ctx context.Context, wallet *Wallet) error {
			require.NotEmpty(t, wallet.ID)
			assert.Equal(t, "user-1", wallet.OwnerID)
			assert.Equal(t, "Primary Wallet", wallet.Name)
			assert.Equal(t, "THB", wallet.Currency)
			assert.Equal(t, WalletTypeDebit, wallet.Type)
			assert.Equal(t, 250.0, wallet.Balance)
			return nil
		},
	}

	transactionSvc := &mockTransactionService{
		createTransactionFn: func(ctx context.Context, walletID, walletName, categoryID, categoryName, description, currency, imageURL, txType string, amount float64, date time.Time, ownerID string) error {
			assert.Equal(t, "Primary Wallet", walletName)
			assert.Equal(t, "cat01", categoryID)
			assert.Equal(t, "Initial balance", categoryName)
			assert.Equal(t, "", description)
			assert.Equal(t, "THB", currency)
			assert.Equal(t, "", imageURL)
			assert.Equal(t, string(transaction.TransactionTypeIncome), txType)
			assert.Equal(t, 250.0, amount)
			assert.Equal(t, "user-1", ownerID)
			return nil
		},
	}

	service := NewService(repo, transactionSvc)
	err := service.CreateWallet(context.Background(), "user-1", "Primary Wallet", "", 250.0, "debit")
	require.NoError(t, err)
}

func TestCreateWalletInvalidTypeReturnsError(t *testing.T) {
	service := NewService(&mockWalletRepository{}, &mockTransactionService{})

	err := service.CreateWallet(context.Background(), "user-1", "Primary Wallet", "USD", 50.0, "invalid")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Invalid transaction type")
}

func TestEditWalletRejectsProtectedFields(t *testing.T) {
	tests := []struct {
		name          string
		changedFields map[string]any
		expectedError string
	}{
		{
			name: "owner id is not updateable",
			changedFields: map[string]any{
				"OwnerID": "user-2",
			},
			expectedError: "Owner ID is not updateable",
		},
		{
			name: "currency is not updateable",
			changedFields: map[string]any{
				"Currency": "USD",
			},
			expectedError: "Currency is not updateable",
		},
		{
			name: "balance is not updateable",
			changedFields: map[string]any{
				"Balance": 100.0,
			},
			expectedError: "Balance is not updateable",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repositoryCalled := false
			repo := &mockWalletRepository{
				editWalletFn: func(ctx context.Context, ownerID string, walletID string, changedFields map[string]any) error {
					repositoryCalled = true
					return nil
				},
			}

			svc := &service{repository: repo}
			err := svc.EditWallet(context.Background(), "user-1", "wallet-1", tc.changedFields)

			require.Error(t, err)
			assert.Equal(t, tc.expectedError, err.Error())
			assert.False(t, repositoryCalled)
		})
	}
}

func TestEditWalletDelegatesToRepository(t *testing.T) {
	changedFields := map[string]any{
		"Name": "Updated Wallet",
	}

	repo := &mockWalletRepository{
		editWalletFn: func(ctx context.Context, ownerID string, walletID string, fields map[string]any) error {
			assert.Equal(t, "user-1", ownerID)
			assert.Equal(t, "wallet-1", walletID)
			assert.Equal(t, changedFields, fields)
			return nil
		},
	}

	svc := &service{repository: repo}
	err := svc.EditWallet(context.Background(), "user-1", "wallet-1", changedFields)
	require.NoError(t, err)
}

func TestEditWalletReturnsRepositoryError(t *testing.T) {
	repoErr := errors.New("repository failure")

	repo := &mockWalletRepository{
		editWalletFn: func(ctx context.Context, ownerID string, walletID string, fields map[string]any) error {
			return repoErr
		},
	}

	svc := &service{repository: repo}
	err := svc.EditWallet(context.Background(), "user-1", "wallet-1", map[string]any{"Name": "Updated"})

	require.Error(t, err)
	assert.Equal(t, repoErr, err)
}

func TestSetActiveWalletDelegatesToEditWallet(t *testing.T) {
	repo := &mockWalletRepository{
		editWalletFn: func(ctx context.Context, ownerID string, walletID string, fields map[string]any) error {
			assert.Equal(t, "user-1", ownerID)
			assert.Equal(t, "wallet-1", walletID)
			assert.Equal(t, map[string]any{"IsActive": true}, fields)
			return nil
		},
	}

	svc := NewService(repo, &mockTransactionService{})
	err := svc.SetActiveWallet(context.Background(), "user-1", "wallet-1", true)
	require.NoError(t, err)
}
