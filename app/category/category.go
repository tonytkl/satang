package category

import "time"

type CategoryType string

const (
	CategoryTypeIncome   CategoryType = "INCOME"
	CategoryTypeExpense  CategoryType = "EXPENSE"
	CategoryTypeTransfer CategoryType = "TRANSFER"
)

// Category represents a category that transaction occurred in.
// DynamoDB keys:
//
//	PK = "USER#<OwnerID>"
//	SK = "CATEGORY#<ID>"
type Category struct {
	PK        string       `dynamodbav:"PK"`
	SK        string       `dynamodbav:"SK"`
	ID        string       `dynamodbav:"ID"`
	OwnerID   string       `dynamodbav:"OwnerID"`
	Name      string       `dynamodbav:"Name"`
	Type      CategoryType `dynamodbav:"Type"`
	IsActive  bool         `dynamodbav:"IsActive"`
	CreatedAt time.Time    `dynamodbav:"CreatedAt"`
	UpdatedAt time.Time    `dynamodbav:"UpdatedAt"`
}

func NewCategory(id string, ownerID string, name string, categoryType CategoryType) *Category {
	now := time.Now().UTC()
	return &Category{
		PK:        "USER#" + ownerID,
		SK:        "CATEGORY#" + id,
		ID:        id,
		OwnerID:   ownerID,
		Name:      name,
		Type:      categoryType,
		IsActive:  true,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func (category *Category) GetID() string {
	return category.ID
}

func (category *Category) SetPK(pk string) {
	category.PK = pk
}

func (category *Category) SetSK(sk string) {
	category.SK = sk
}

func (category *Category) GetCreatedAt() time.Time {
	return category.CreatedAt
}

func (category *Category) SetCreatedAt(createdAt time.Time) {
	category.CreatedAt = createdAt
}

func (category *Category) GetUpdatedAt() time.Time {
	return category.UpdatedAt
}

func (category *Category) SetUpdatedAt(updatedAt time.Time) {
	category.UpdatedAt = updatedAt
}

func (category *Category) GetOwnerID() string {
	return category.OwnerID
}
