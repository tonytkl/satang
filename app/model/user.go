package model

import "time"

// User represents a registered user in the system.
// DynamoDB keys:
//
//	PK = "USER#<ID>"
//	SK = "#METADATA#<ID>"
//	GSI_PK = "USER_EMAIL#<Email>"
//	GSI_SK = "USER#<ID>"
type User struct {
	PK        string    `dynamodbav:"PK"`
	SK        string    `dynamodbav:"SK"`
	UpdatedAt time.Time `dynamodbav:"UpdatedAt"`
	CreatedAt time.Time `dynamodbav:"CreatedAt"`
	Name      string    `dynamodbav:"Name"`
	Email     string    `dynamodbav:"Email"`
	ID        string    `dynamodbav:"ID"`
	GSISK     string    `dynamodbav:"GSI_SK"`
	GSIPK     string    `dynamodbav:"GSI_PK"`
}

func NewUser(id, name, email string) *User {
	now := time.Now().UTC()
	return &User{
		PK:        "USER#" + id,
		SK:        "#METADATA#" + id,
		UpdatedAt: now,
		CreatedAt: now,
		Name:      name,
		Email:     email,
		ID:        id,
		GSISK:     "USER#" + id,
		GSIPK:     "USER_EMAIL#" + email,
	}
}
