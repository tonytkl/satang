package model

import "time"

type CategoryType string

const (
	CategoryTypeIncome   CategoryType = "INCOME"
	CategoryTypeExpense  CategoryType = "EXPENSE"
	CategoryTypeTransfer CategoryType = "TRANSFER"
)

// Category represents a transaction category defined by a user.
// Categories are used to classify transactions and can be customized by each user.
// DynamoDB keys:
//
// PK = "CATEGORY#<OwnerID>"
// SK = "CATEGORY#<ID>"
// GSI_PK = "CATEGORY#<OwnerID>#TYPE#<Type>"
// GSI_SK = "CATEGORY#<ID>"

type Category struct {
	PK        string       `dynamodbav:"PK"`
	SK        string       `dynamodbav:"SK"`
	GSIPK     string       `dynamodbav:"GSI_PK"`
	GSISK     string       `dynamodbav:"GSI_SK"`
	ID        string       `dynamodbav:"ID"`
	OwnerID   string       `dynamodbav:"OwnerID"`
	Type      CategoryType `dynamodbav:"Type"`
	Name      string       `dynamodbav:"Name"`
	CreatedAt time.Time    `dynamodbav:"CreatedAt"`
	UpdatedAt time.Time    `dynamodbav:"UpdatedAt"`
}

func categoryGSIKey(ownerID string, categoryType CategoryType) string {
	return "CATEGORY#" + ownerID + "#TYPE#" + string(categoryType)
}

func NewCategory(id, ownerID, name string, categoryType CategoryType) *Category {
	now := time.Now().UTC()
	return &Category{
		PK:        "CATEGORY#" + ownerID,
		SK:        "CATEGORY#" + id,
		GSIPK:     categoryGSIKey(ownerID, categoryType),
		GSISK:     "CATEGORY#" + id,
		ID:        id,
		OwnerID:   ownerID,
		Type:      categoryType,
		Name:      name,
		CreatedAt: now,
		UpdatedAt: now,
	}
}
