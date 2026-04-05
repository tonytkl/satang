package model

import "time"

// WalletMemberRole defines the permission level of a wallet member.
type WalletMemberRole string

const (
	WalletMemberRoleOwner  WalletMemberRole = "OWNER"
	WalletMemberRoleMember WalletMemberRole = "MEMBER"
)

// WalletMember represents the many-to-many relationship between users and wallets
// using the adjacency list pattern. Each membership is stored as two items:
//
//  1. Forward link (user → wallet) — for listing all wallets a user belongs to:
//     PK = "USER#<UserID>"
//     SK = "WALLET#<WalletID>"
//
//  2. Reverse link (wallet → user) — for listing all members of a wallet:
//     PK = "WALLET#<WalletID>"
//     SK = "MEMBER#<UserID>"
//
// Both items must be written and deleted together (TransactWriteItems).
type WalletMember struct {
	PK       string           `dynamodbav:"PK"`
	SK       string           `dynamodbav:"SK"`
	UserID   string           `dynamodbav:"UserID"`
	WalletID string           `dynamodbav:"WalletID"`
	Role     WalletMemberRole `dynamodbav:"Role"`
	JoinedAt time.Time        `dynamodbav:"JoinedAt"`
}

// NewWalletMemberForwardLink returns the user-partition item (user → wallet).
// Query PK="USER#<UserID>", SK begins_with "WALLET#" to list a user's wallets.
func NewWalletMemberForwardLink(userID, walletID string, role WalletMemberRole) *WalletMember {
	return &WalletMember{
		PK:       "USER#" + userID,
		SK:       "WALLET#" + walletID,
		UserID:   userID,
		WalletID: walletID,
		Role:     role,
		JoinedAt: time.Now().UTC(),
	}
}

// NewWalletMemberReverseLink returns the wallet-partition item (wallet → user).
// Query PK="WALLET#<WalletID>", SK begins_with "MEMBER#" to list a wallet's members.
func NewWalletMemberReverseLink(userID, walletID string, role WalletMemberRole) *WalletMember {
	return &WalletMember{
		PK:       "WALLET#" + walletID,
		SK:       "MEMBER#" + userID,
		UserID:   userID,
		WalletID: walletID,
		Role:     role,
		JoinedAt: time.Now().UTC(),
	}
}
