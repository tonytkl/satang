package wallet

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockWalletRepository struct {
	createWalletFn func(ctx context.Context, wallet Wallet) error
	listWalletsFn  func(ctx context.Context, ownerID string, nextToken string, limit int32) ([]Wallet, string, error)
	getWalletFn    func(ctx context.Context, ownerID string, walletID string) (Wallet, error)
	editWalletFn   func(ctx context.Context, ownerID string, walletID string, changedFields map[string]any) (Wallet, error)
	deleteWalletFn func(ctx context.Context, ownerID string, walletID string) error
}

var _ Repository = (*mockWalletRepository)(nil)

func (m *mockWalletRepository) CreateWallet(ctx context.Context, wallet Wallet) error {
	if m.createWalletFn != nil {
		return m.createWalletFn(ctx, wallet)
	}
	return nil
}

func (m *mockWalletRepository) ListWallets(ctx context.Context, ownerID string, nextToken string, limit int32) ([]Wallet, string, error) {
	if m.listWalletsFn != nil {
		return m.listWalletsFn(ctx, ownerID, nextToken, limit)
	}
	return nil, "", nil
}

func (m *mockWalletRepository) GetWallet(ctx context.Context, ownerID string, walletID string) (Wallet, error) {
	if m.getWalletFn != nil {
		return m.getWalletFn(ctx, ownerID, walletID)
	}
	return Wallet{}, nil
}

func (m *mockWalletRepository) EditWallet(ctx context.Context, ownerID string, walletID string, changedFields map[string]any) (Wallet, error) {
	if m.editWalletFn != nil {
		return m.editWalletFn(ctx, ownerID, walletID, changedFields)
	}
	return Wallet{}, nil
}

func (m *mockWalletRepository) DeleteWallet(ctx context.Context, ownerID string, walletID string) error {
	if m.deleteWalletFn != nil {
		return m.deleteWalletFn(ctx, ownerID, walletID)
	}
	return nil
}

func TestCreateWalletUsesDefaultCurrencyAndInitialBalance(t *testing.T) {
	repo := &mockWalletRepository{
		createWalletFn: func(ctx context.Context, wallet Wallet) error {
			require.NotEmpty(t, wallet.ID)
			assert.Equal(t, "user-1", wallet.OwnerID)
			assert.Equal(t, "Primary Wallet", wallet.Name)
			assert.Equal(t, "THB", wallet.Currency)
			assert.Equal(t, WalletTypeDebit, wallet.Type)
			assert.Equal(t, 250.0, wallet.Balance)
			return nil
		},
	}

	service := NewService(repo)
	got, err := service.CreateWallet(context.Background(), "user-1", "Primary Wallet", "", 250.0, "debit")
	require.NoError(t, err)
	assert.Equal(t, "Primary Wallet", got.Name)
	assert.Equal(t, "THB", got.Currency)
	assert.Equal(t, WalletTypeDebit, got.Type)
	assert.Equal(t, 250.0, got.Balance)
}

func TestCreateWalletInvalidTypeReturnsError(t *testing.T) {
	service := NewService(&mockWalletRepository{})

	_, err := service.CreateWallet(context.Background(), "user-1", "Primary Wallet", "USD", 50.0, "invalid")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Invalid wallet type")
}

func TestCreateWalletRejectsEmptyOwnerID(t *testing.T) {
	repoCalled := false
	repo := &mockWalletRepository{
		createWalletFn: func(ctx context.Context, wallet Wallet) error {
			repoCalled = true
			return nil
		},
	}

	service := NewService(repo)
	_, err := service.CreateWallet(context.Background(), "", "Primary Wallet", "USD", 50.0, "debit")

	require.Error(t, err)
	assert.Equal(t, "owner ID is required", err.Error())
	assert.False(t, repoCalled)
}

func TestListWalletsRejectsNegativeLimit(t *testing.T) {
	repoCalled := false
	repo := &mockWalletRepository{
		listWalletsFn: func(ctx context.Context, ownerID string, nextToken string, limit int32) ([]Wallet, string, error) {
			repoCalled = true
			return []Wallet{}, "", nil
		},
	}

	service := NewService(repo)
	_, _, err := service.ListWallets(context.Background(), "user-1", "", -1)

	require.Error(t, err)
	assert.Equal(t, "limit must be greater than or equal to 0", err.Error())
	assert.False(t, repoCalled)
}

func TestGetWalletRejectsEmptyInputsAndPropagatesRepositoryErrors(t *testing.T) {
	t.Run("empty owner id", func(t *testing.T) {
		repoCalled := false
		repo := &mockWalletRepository{
			getWalletFn: func(ctx context.Context, ownerID string, walletID string) (Wallet, error) {
				repoCalled = true
				return Wallet{}, nil
			},
		}

		service := NewService(repo)
		_, err := service.GetWallet(context.Background(), "", "wallet-1")

		require.Error(t, err)
		assert.Equal(t, "owner ID is required", err.Error())
		assert.False(t, repoCalled)
	})

	t.Run("empty wallet id", func(t *testing.T) {
		repoCalled := false
		repo := &mockWalletRepository{
			getWalletFn: func(ctx context.Context, ownerID string, walletID string) (Wallet, error) {
				repoCalled = true
				return Wallet{}, nil
			},
		}

		service := NewService(repo)
		_, err := service.GetWallet(context.Background(), "user-1", "")

		require.Error(t, err)
		assert.Equal(t, "wallet ID is required", err.Error())
		assert.False(t, repoCalled)
	})

	t.Run("repository error is propagated", func(t *testing.T) {
		repo := &mockWalletRepository{
			getWalletFn: func(ctx context.Context, ownerID string, walletID string) (Wallet, error) {
				assert.Equal(t, "user-1", ownerID)
				assert.Equal(t, "wallet-1", walletID)
				return Wallet{}, ErrWalletNotFound
			},
		}

		service := NewService(repo)
		got, err := service.GetWallet(context.Background(), "user-1", "wallet-1")

		require.ErrorIs(t, err, ErrWalletNotFound)
		assert.Equal(t, Wallet{}, got)
	})
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
				editWalletFn: func(ctx context.Context, ownerID string, walletID string, changedFields map[string]any) (Wallet, error) {
					repositoryCalled = true
					return Wallet{}, nil
				},
			}

			svc := &service{repository: repo}
			_, err := svc.EditWallet(context.Background(), "user-1", "wallet-1", tc.changedFields)

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
		editWalletFn: func(ctx context.Context, ownerID string, walletID string, fields map[string]any) (Wallet, error) {
			assert.Equal(t, "user-1", ownerID)
			assert.Equal(t, "wallet-1", walletID)
			assert.Equal(t, changedFields, fields)
			return Wallet{ID: walletID, OwnerID: ownerID, Name: "Updated Wallet"}, nil
		},
	}

	svc := &service{repository: repo}
	got, err := svc.EditWallet(context.Background(), "user-1", "wallet-1", changedFields)
	require.NoError(t, err)
	assert.Equal(t, "Updated Wallet", got.Name)
}

func TestEditWalletConvertsTypeBeforeRepository(t *testing.T) {
	changedFields := map[string]any{
		"Type": "credit",
	}

	repo := &mockWalletRepository{
		editWalletFn: func(ctx context.Context, ownerID string, walletID string, fields map[string]any) (Wallet, error) {
			assert.Equal(t, "user-1", ownerID)
			assert.Equal(t, "wallet-1", walletID)

			typeValue, ok := fields["Type"]
			require.True(t, ok)
			assert.Equal(t, WalletTypeCredit, typeValue)
			return Wallet{ID: walletID, OwnerID: ownerID, Type: WalletTypeCredit}, nil
		},
	}

	svc := &service{repository: repo}
	got, err := svc.EditWallet(context.Background(), "user-1", "wallet-1", changedFields)
	require.NoError(t, err)
	assert.Equal(t, WalletTypeCredit, got.Type)
}

func TestEditWalletRejectsNonStringType(t *testing.T) {
	repositoryCalled := false
	repo := &mockWalletRepository{
		editWalletFn: func(ctx context.Context, ownerID string, walletID string, changedFields map[string]any) (Wallet, error) {
			repositoryCalled = true
			return Wallet{}, nil
		},
	}

	svc := &service{repository: repo}
	_, err := svc.EditWallet(context.Background(), "user-1", "wallet-1", map[string]any{"Type": 123})

	require.Error(t, err)
	assert.Equal(t, "Type must be a string", err.Error())
	assert.False(t, repositoryCalled)
}

func TestEditWalletReturnsRepositoryError(t *testing.T) {
	repoErr := errors.New("repository failure")

	repo := &mockWalletRepository{
		editWalletFn: func(ctx context.Context, ownerID string, walletID string, fields map[string]any) (Wallet, error) {
			return Wallet{}, repoErr
		},
	}

	svc := &service{repository: repo}
	_, err := svc.EditWallet(context.Background(), "user-1", "wallet-1", map[string]any{"Name": "Updated"})

	require.Error(t, err)
	assert.Equal(t, repoErr, err)
}

func TestEditWalletReturnsWalletNotFound(t *testing.T) {
	repo := &mockWalletRepository{
		editWalletFn: func(context.Context, string, string, map[string]any) (Wallet, error) {
			return Wallet{}, ErrWalletNotFound
		},
	}

	svc := &service{repository: repo}
	_, err := svc.EditWallet(context.Background(), "user-1", "missing-wallet", map[string]any{"Name": "Updated"})

	require.ErrorIs(t, err, ErrWalletNotFound)
}

func TestSetActiveWalletDelegatesToEditWallet(t *testing.T) {
	repo := &mockWalletRepository{
		editWalletFn: func(ctx context.Context, ownerID string, walletID string, fields map[string]any) (Wallet, error) {
			assert.Equal(t, "user-1", ownerID)
			assert.Equal(t, "wallet-1", walletID)
			assert.Equal(t, map[string]any{"IsActive": true}, fields)
			return Wallet{}, nil
		},
	}

	svc := NewService(repo)
	err := svc.SetActiveWallet(context.Background(), "user-1", "wallet-1", true)
	require.NoError(t, err)
}
