package repository

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/tonytkl/satang/clients"
	"github.com/tonytkl/satang/utils"
)

type SatangModel interface {
	// Model functions: Only shared functions with all models
	GetID() string
	SetPK(string)
	SetSK(sk string)
	GetCreatedAt() time.Time
	SetCreatedAt(createdAt time.Time)
	GetUpdatedAt() time.Time
	SetUpdatedAt(updatedAt time.Time)
	GetOwnerID() string
}

type BaseRepository[T SatangModel] interface {
	Save(ctx context.Context, item T) error
	Get(ctx context.Context, ownerID string, itemID string) (T, error)
	List(ctx context.Context, ownerID string, nextToken string, limit int32) ([]T, string, error)
	Update(ctx context.Context, ownerID string, itemID string, changedFields map[string]any) error
	Delete(ctx context.Context, ownerID string, itemID string) error
}

type baseRepository[T SatangModel] struct {
	db        clients.DynamoDBClient
	tableName string
	skModel   string
	newItem   func() T
}

func NewBaseRepository[T SatangModel](db clients.DynamoDBClient, tableName string, skModel string, newItem func() T) BaseRepository[T] {
	return &baseRepository[T]{
		db:        db,
		tableName: tableName,
		skModel:   skModel,
		newItem:   newItem,
	}
}

func (repository *baseRepository[T]) Save(ctx context.Context, item T) error {
	item.SetPK(utils.GetPartitionKey("USER", item.GetOwnerID()))
	item.SetSK(utils.GetPartitionKey(repository.skModel, item.GetID()))

	if item.GetCreatedAt().IsZero() {
		item.SetCreatedAt(time.Now().UTC())
	}

	if item.GetUpdatedAt().IsZero() {
		item.SetUpdatedAt(time.Now().UTC())
	}

	putItemErr := repository.db.PutItem(ctx, repository.tableName, item)

	if putItemErr != nil {
		return putItemErr
	}

	return nil
}

func (repository *baseRepository[T]) Get(ctx context.Context, ownerID string, itemID string) (T, error) {
	var zero T

	if ownerID == "" {
		return zero, errors.New("owner ID is required")
	}

	if itemID == "" {
		return zero, errors.New("item ID is required")
	}

	key := map[string]any{
		"PK": utils.GetPartitionKey("USER", ownerID),
		"SK": utils.GetPartitionKey(repository.skModel, itemID),
	}

	item := repository.newItem()

	err := repository.db.GetItem(
		ctx,
		repository.tableName,
		key,
		item,
	)

	if err != nil {
		return zero, fmt.Errorf("Error on Get: %w", err)
	}

	return item, nil
}

func (repository *baseRepository[T]) List(ctx context.Context, ownerID string, nextToken string, limit int32) ([]T, string, error) {
	if ownerID == "" {
		return nil, "", errors.New("owner ID is required")
	}
	queryExpression := "PK = :ownerPK AND begins_with(SK, :modelPrefix)"
	experessionValues := map[string]any{
		":ownerPK":     utils.GetPartitionKey("USER", ownerID),
		":modelPrefix": repository.skModel + "#",
	}

	var items []T
	encodedNextToken, err := repository.db.QueryItemsWithPagination(
		ctx,
		repository.tableName,
		queryExpression,
		experessionValues,
		"",
		"",
		limit,
		nextToken,
		&items,
	)

	if err != nil {
		return nil, "", fmt.Errorf("Errors on List: %w", err)
	}

	return items, encodedNextToken, nil
}

func (repository *baseRepository[T]) Update(ctx context.Context, ownerID string, itemID string, changedFields map[string]any) error {
	if ownerID == "" {
		return errors.New("owner ID is required")
	}

	if itemID == "" {
		return errors.New("item ID is required")
	}

	if len(changedFields) == 0 {
		return errors.New("Update payload is required")
	}

	key := map[string]any{
		"PK": utils.GetPartitionKey("USER", ownerID),
		"SK": utils.GetPartitionKey(repository.skModel, itemID),
	}

	raw := make(map[string]any, len(changedFields)+1)
	for key, value := range changedFields {
		raw[key] = value
	}
	raw["UpdatedAt"] = time.Now().UTC()

	delete(raw, "PK")
	delete(raw, "SK")
	delete(raw, "ID")
	delete(raw, "OwnerID")
	delete(raw, "CreatedAt")

	if len(raw) == 0 {
		return errors.New("no mutable fields to update")
	}

	setParts := make([]string, 0, len(raw))
	exprValues := map[string]any{
		":id": itemID,
	}
	exprNames := map[string]string{}
	attrs := make([]string, 0, len(raw))
	for attr := range raw {
		attrs = append(attrs, attr)
	}
	sort.Strings(attrs)

	for index, attr := range attrs {
		nameField := fmt.Sprintf("#n%d", index)
		valueField := fmt.Sprintf(":v%d", index)
		setParts = append(setParts, nameField+" = "+valueField)
		exprNames[nameField] = attr
		exprValues[valueField] = raw[attr]
	}
	updateExpression := "SET " + strings.Join(setParts, ",")
	conditionExpression := "attribute_exists(PK) AND attribute_exists(SK) AND ID = :id"

	if err := repository.db.UpdateItem(ctx, repository.tableName, key, updateExpression, exprValues, exprNames, conditionExpression); err != nil {
		return fmt.Errorf("update item: %w", err)
	}

	return nil
}

func (repository *baseRepository[T]) Delete(ctx context.Context, ownerID string, itemID string) error {
	if ownerID == "" {
		return errors.New("owner ID is required")
	}

	if itemID == "" {
		return errors.New("item ID is required")
	}

	key := map[string]any{
		"PK": utils.GetPartitionKey("USER", ownerID),
		"SK": utils.GetPartitionKey(repository.skModel, itemID),
	}

	err := repository.db.DeleteItem(
		ctx,
		repository.tableName,
		key,
	)

	if err != nil {
		return fmt.Errorf("Error on Delete: %w", err)
	}

	return nil
}
