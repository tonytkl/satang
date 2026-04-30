package repositories

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/tonytkl/satang/clients"
	"github.com/tonytkl/satang/model"
	"github.com/tonytkl/satang/repositories"
)

type mockDynamoDB struct {
	putItemFn    func(ctx context.Context, table string, item any) error
	updateItemFn func(ctx context.Context, table string, key map[string]any, updateExpression string, expressionValues map[string]any, conditionExpression string) error
	getItemFn    func(ctx context.Context, table string, key map[string]any, out any) error
	deleteItemFn func(ctx context.Context, table string, key map[string]any) error
	queryItemsFn func(ctx context.Context, table string, keyConditionExpression string, expressionValues map[string]any, indexName string, filterExpression string, out any) error
	scanItemsFn  func(ctx context.Context, table string, filterExpression string, expressionValues map[string]any, out any) error
}

var _ clients.DynamoDBClient = (*mockDynamoDB)(nil)

func (m *mockDynamoDB) PutItem(ctx context.Context, table string, item any) error {
	if m.putItemFn != nil {
		return m.putItemFn(ctx, table, item)
	}
	return nil
}

func (m *mockDynamoDB) UpdateItem(ctx context.Context, table string, key map[string]any, updateExpression string, expressionValues map[string]any, conditionExpression string) error {
	if m.updateItemFn != nil {
		return m.updateItemFn(ctx, table, key, updateExpression, expressionValues, conditionExpression)
	}
	return nil
}

func (m *mockDynamoDB) GetItem(ctx context.Context, table string, key map[string]any, out any) error {
	if m.getItemFn != nil {
		return m.getItemFn(ctx, table, key, out)
	}
	return nil
}

func (m *mockDynamoDB) DeleteItem(ctx context.Context, table string, key map[string]any) error {
	if m.deleteItemFn != nil {
		return m.deleteItemFn(ctx, table, key)
	}
	return nil
}

func (m *mockDynamoDB) QueryItems(ctx context.Context, table string, keyConditionExpression string, expressionValues map[string]any, indexName string, filterExpression string, out any) error {
	if m.queryItemsFn != nil {
		return m.queryItemsFn(ctx, table, keyConditionExpression, expressionValues, indexName, filterExpression, out)
	}
	return nil
}

func (m *mockDynamoDB) ScanItems(ctx context.Context, table string, filterExpression string, expressionValues map[string]any, out any) error {
	if m.scanItemsFn != nil {
		return m.scanItemsFn(ctx, table, filterExpression, expressionValues, out)
	}
	return nil
}

func TestTransactionRepositoryCreateSuccess(t *testing.T) {
	mock := &mockDynamoDB{
		putItemFn: func(_ context.Context, table string, item any) error {
			if table != "transactions" {
				t.Fatalf("table = %s, want transactions", table)
			}

			tx, ok := item.(*model.Transaction)
			if !ok {
				t.Fatalf("item type = %T, want *model.Transaction", item)
			}

			if tx.ID == "" {
				t.Fatalf("ID should be generated")
			}
			if tx.PK != "USER#user-1" {
				t.Fatalf("PK = %s, want USER#user-1", tx.PK)
			}
			if tx.SK != "TX#2026-04-15" {
				t.Fatalf("SK = %s, want TX#2026-04-15", tx.SK)
			}
			if tx.GSI_ByCategoryPK != "TX_CATEGORY#cat-1" {
				t.Fatalf("GSI_ByCategoryPK = %s, want TX_CATEGORY#cat-1", tx.GSI_ByCategoryPK)
			}
			if tx.GSI_ByCategorySK != "TX#2026-04-15" {
				t.Fatalf("GSI_ByCategorySK = %s, want TX#2026-04-15", tx.GSI_ByCategorySK)
			}
			if tx.GSI_ByWalletPK != "TX_WALLET#wallet-1" {
				t.Fatalf("GSI_ByWalletPK = %s, want TX_WALLET#wallet-1", tx.GSI_ByWalletPK)
			}
			if tx.GSI_ByWalletSK != "TX#2026-04-15" {
				t.Fatalf("GSI_ByWalletSK = %s, want TX#2026-04-15", tx.GSI_ByWalletSK)
			}
			if !strings.HasPrefix(tx.GSI_ByTransactionID, "TX_ID#") {
				t.Fatalf("GSI_ByTransactionID = %s, should start with TX_ID#", tx.GSI_ByTransactionID)
			}
			if tx.CreatedAt.IsZero() || tx.UpdatedAt.IsZero() {
				t.Fatalf("created/updated timestamps should be set")
			}

			return nil
		},
	}

	repo := repositories.NewTransactionRepository(mock, "transactions")
	tx := &model.Transaction{
		OwnerID:    "user-1",
		WalletID:   "wallet-1",
		CategoryID: "cat-1",
		Amount:     100,
		Currency:   "THB",
		Date:       time.Date(2026, 4, 15, 10, 0, 0, 0, time.UTC),
	}

	err := repo.Create(context.Background(), tx)
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
}

func TestTransactionRepositoryCreateValidationError(t *testing.T) {
	repo := repositories.NewTransactionRepository(&mockDynamoDB{}, "transactions")

	err := repo.Create(context.Background(), &model.Transaction{
		OwnerID:    "user-1",
		WalletID:   "wallet-1",
		CategoryID: "cat-1",
		Currency:   "THB",
		Date:       time.Date(2026, 4, 15, 10, 0, 0, 0, time.UTC),
	})

	if err == nil || err.Error() != "Transaction amount is required" {
		t.Fatalf("err = %v, want Transaction amount is required", err)
	}
}

func TestTransactionRepositoryListByGSISuccess(t *testing.T) {
	mock := &mockDynamoDB{
		queryItemsFn: func(_ context.Context, table string, keyConditionExpression string, expressionValues map[string]any, indexName string, filterExpression string, out any) error {
			if table != "transactions" {
				t.Fatalf("table = %s, want transactions", table)
			}
			if indexName != "GSI1" {
				t.Fatalf("indexName = %s, want GSI1", indexName)
			}
			if keyConditionExpression != "GSI_PK = :indexPK AND GSI_SK BETWEEN :from AND :to" {
				t.Fatalf("keyConditionExpression = %s", keyConditionExpression)
			}
			if filterExpression != "PK = :ownerPK" {
				t.Fatalf("filterExpression = %s, want PK = :ownerPK", filterExpression)
			}
			if expressionValues[":indexPK"] != "TX_CATEGORY#cat-1" {
				t.Fatalf(":indexPK = %v, want TX_CATEGORY#cat-1", expressionValues[":indexPK"])
			}
			if expressionValues[":ownerPK"] != "USER#user-1" {
				t.Fatalf(":ownerPK = %v, want USER#user-1", expressionValues[":ownerPK"])
			}
			if expressionValues[":from"] != "TX#2026-04-01" {
				t.Fatalf(":from = %v, want TX#2026-04-01", expressionValues[":from"])
			}
			if expressionValues[":to"] != "TX#2026-04-30" {
				t.Fatalf(":to = %v, want TX#2026-04-30", expressionValues[":to"])
			}

			dst, ok := out.(*[]model.Transaction)
			if !ok {
				t.Fatalf("out type = %T, want *[]model.Transaction", out)
			}
			*dst = []model.Transaction{{ID: "tx-1"}}

			return nil
		},
	}

	repo := repositories.NewTransactionRepository(mock, "transactions")
	from := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 4, 30, 0, 0, 0, 0, time.UTC)

	got, err := repo.ListByGSI(context.Background(), "GSI1", "TX_CATEGORY", "cat-1", "user-1", &from, &to)
	if err != nil {
		t.Fatalf("ListByGSI returned error: %v", err)
	}
	if len(got) != 1 || got[0].ID != "tx-1" {
		t.Fatalf("ListByGSI returned unexpected result: %#v", got)
	}
}

func TestTransactionRepositoryListByGSIErrors(t *testing.T) {
	repo := repositories.NewTransactionRepository(&mockDynamoDB{}, "transactions")

	_, err := repo.ListByGSI(context.Background(), "", "TX_CATEGORY", "cat-1", "", nil, nil)
	if err == nil || err.Error() != "index name, index partition key prefix, and target ID are required" {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = repo.ListByGSI(context.Background(), "BAD_INDEX", "TX_CATEGORY", "cat-1", "", nil, nil)
	if err == nil || !strings.Contains(err.Error(), "unsupported index name") {
		t.Fatalf("unexpected error: %v", err)
	}

	from := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)
	_, err = repo.ListByGSI(context.Background(), "GSI1", "TX_CATEGORY", "cat-1", "", &from, &to)
	if err == nil || err.Error() != "from date must not be after to date" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTransactionRepositoryListByGSINotFound(t *testing.T) {
	mock := &mockDynamoDB{
		queryItemsFn: func(_ context.Context, _ string, _ string, _ map[string]any, _ string, _ string, out any) error {
			dst := out.(*[]model.Transaction)
			*dst = []model.Transaction{}
			return nil
		},
	}

	repo := repositories.NewTransactionRepository(mock, "transactions")

	_, err := repo.ListByGSI(context.Background(), "GSI3", "TX_ID", "tx-1", "", nil, nil)
	if !errors.Is(err, repositories.ErrTransactionNotFound) {
		t.Fatalf("err = %v, want ErrTransactionNotFound", err)
	}
}

func TestTransactionRepositoryListWithinDateRangeSuccess(t *testing.T) {
	mock := &mockDynamoDB{
		queryItemsFn: func(_ context.Context, _ string, keyConditionExpression string, expressionValues map[string]any, indexName string, filterExpression string, out any) error {
			if keyConditionExpression != "PK = :pk AND SK BETWEEN :from AND :to" {
				t.Fatalf("keyConditionExpression = %s", keyConditionExpression)
			}
			if indexName != "" || filterExpression != "" {
				t.Fatalf("indexName/filterExpression should be empty")
			}
			if expressionValues[":pk"] != "USER#user-1" {
				t.Fatalf(":pk = %v, want USER#user-1", expressionValues[":pk"])
			}
			if expressionValues[":from"] != "TX#2026-04-01" {
				t.Fatalf(":from = %v, want TX#2026-04-01", expressionValues[":from"])
			}
			if expressionValues[":to"] != "TX#2026-04-30" {
				t.Fatalf(":to = %v, want TX#2026-04-30", expressionValues[":to"])
			}

			dst := out.(*[]model.Transaction)
			*dst = []model.Transaction{{ID: "tx-1"}, {ID: "tx-2"}}
			return nil
		},
	}

	repo := repositories.NewTransactionRepository(mock, "transactions")
	from := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 4, 30, 0, 0, 0, 0, time.UTC)

	got, err := repo.ListWithinDateRange(context.Background(), "user-1", from, to)
	if err != nil {
		t.Fatalf("ListWithinDateRange returned error: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("len(got) = %d, want 2", len(got))
	}
}

func TestTransactionRepositoryGetByKeySuccess(t *testing.T) {
	mock := &mockDynamoDB{
		queryItemsFn: func(_ context.Context, _ string, keyConditionExpression string, expressionValues map[string]any, indexName string, filterExpression string, out any) error {
			if indexName != "GSI3" {
				t.Fatalf("indexName = %s, want GSI3", indexName)
			}
			if keyConditionExpression != "GSI3_PK = :indexPK" {
				t.Fatalf("keyConditionExpression = %s, want GSI3_PK = :indexPK", keyConditionExpression)
			}
			if filterExpression != "" {
				t.Fatalf("filterExpression = %s, want empty", filterExpression)
			}
			if expressionValues[":indexPK"] != "TX_ID#tx-1" {
				t.Fatalf(":indexPK = %v, want TX_ID#tx-1", expressionValues[":indexPK"])
			}

			dst := out.(*[]model.Transaction)
			*dst = []model.Transaction{{ID: "tx-1"}}
			return nil
		},
	}

	repo := repositories.NewTransactionRepository(mock, "transactions")

	got, err := repo.GetByKey(context.Background(), "tx-1")
	if err != nil {
		t.Fatalf("GetByKey returned error: %v", err)
	}
	if got == nil || got.ID != "tx-1" {
		t.Fatalf("GetByKey returned unexpected result: %#v", got)
	}
}

func TestTransactionRepositoryUpdateSuccess(t *testing.T) {
	desc := "new description"
	image := "https://example.com/image.png"

	mock := &mockDynamoDB{
		updateItemFn: func(_ context.Context, table string, key map[string]any, updateExpression string, expressionValues map[string]any, conditionExpression string) error {
			if table != "transactions" {
				t.Fatalf("table = %s, want transactions", table)
			}
			if key["PK"] != "USER#user-1" {
				t.Fatalf("key PK = %v, want USER#user-1", key["PK"])
			}
			if key["SK"] != "TX#2026-04-20" {
				t.Fatalf("key SK = %v, want TX#2026-04-20", key["SK"])
			}

			wantCond := "attribute_exists(PK) AND attribute_exists(SK) AND ID = :transactionID"
			if conditionExpression != wantCond {
				t.Fatalf("conditionExpression = %s", conditionExpression)
			}

			if !strings.Contains(updateExpression, "SET WalletID = :walletID") {
				t.Fatalf("unexpected updateExpression: %s", updateExpression)
			}

			if expressionValues[":walletID"] != "wallet-2" {
				t.Fatalf(":walletID = %v, want wallet-2", expressionValues[":walletID"])
			}
			if expressionValues[":gsiCategoryPK"] != "TX_CATEGORY#cat-2" {
				t.Fatalf(":gsiCategoryPK = %v, want TX_CATEGORY#cat-2", expressionValues[":gsiCategoryPK"])
			}
			if expressionValues[":gsiWalletPK"] != "TX_WALLET#wallet-2" {
				t.Fatalf(":gsiWalletPK = %v, want TX_WALLET#wallet-2", expressionValues[":gsiWalletPK"])
			}
			if expressionValues[":transactionID"] != "tx-1" {
				t.Fatalf(":transactionID = %v, want tx-1", expressionValues[":transactionID"])
			}

			updatedAt, ok := expressionValues[":updatedAt"].(time.Time)
			if !ok || updatedAt.IsZero() {
				t.Fatalf(":updatedAt should be a non-zero time.Time")
			}

			return nil
		},
	}

	repo := repositories.NewTransactionRepository(mock, "transactions")
	err := repo.Update(context.Background(), "user-1", "2026-04-20", "tx-1", &model.Transaction{
		WalletID:     "wallet-2",
		WalletName:   "Cash",
		Amount:       999,
		Currency:     "THB",
		CategoryID:   "cat-2",
		CategoryName: "Food",
		Description:  &desc,
		ImageURL:     &image,
	})
	if err != nil {
		t.Fatalf("Update returned error: %v", err)
	}
}

func TestTransactionRepositoryDeleteSuccess(t *testing.T) {
	mock := &mockDynamoDB{
		deleteItemFn: func(_ context.Context, table string, key map[string]any) error {
			if table != "transactions" {
				t.Fatalf("table = %s, want transactions", table)
			}
			if key["PK"] != "USER#user-1" {
				t.Fatalf("key PK = %v, want USER#user-1", key["PK"])
			}
			if key["SK"] != "TX#2026-04-20" {
				t.Fatalf("key SK = %v, want TX#2026-04-20", key["SK"])
			}
			return nil
		},
	}

	repo := repositories.NewTransactionRepository(mock, "transactions")
	err := repo.Delete(context.Background(), "user-1", "2026-04-20", "tx-1")
	if err != nil {
		t.Fatalf("Delete returned error: %v", err)
	}
}

func TestTransactionRepositoryUpdateAndDeleteErrorPaths(t *testing.T) {
	repo := repositories.NewTransactionRepository(&mockDynamoDB{}, "transactions")

	err := repo.Update(context.Background(), "", "2026-04-20", "tx-1", &model.Transaction{})
	if err == nil || err.Error() != "owner ID is required" {
		t.Fatalf("unexpected Update error: %v", err)
	}

	err = repo.Delete(context.Background(), "user-1", "bad-date", "tx-1")
	if err == nil || !strings.Contains(err.Error(), "invalid transaction date format") {
		t.Fatalf("unexpected Delete error: %v", err)
	}
}

func TestTransactionRepositoryDBErrorWrapping(t *testing.T) {
	dbErr := errors.New("dynamodb failed")

	mock := &mockDynamoDB{
		updateItemFn: func(_ context.Context, _ string, _ map[string]any, _ string, _ map[string]any, _ string) error {
			return dbErr
		},
		deleteItemFn: func(_ context.Context, _ string, _ map[string]any) error {
			return dbErr
		},
		queryItemsFn: func(_ context.Context, _ string, _ string, _ map[string]any, _ string, _ string, _ any) error {
			return dbErr
		},
	}

	repo := repositories.NewTransactionRepository(mock, "transactions")

	err := repo.Update(context.Background(), "user-1", "2026-04-20", "tx-1", &model.Transaction{
		WalletID:   "wallet-1",
		CategoryID: "cat-1",
	})
	if err == nil || !strings.Contains(err.Error(), "update transaction") || !errors.Is(err, dbErr) {
		t.Fatalf("unexpected Update error: %v", err)
	}

	_, err = repo.ListByGSI(context.Background(), "GSI3", "TX_ID", "tx-1", "", nil, nil)
	if err == nil || !strings.Contains(err.Error(), "query transaction by ID") || !errors.Is(err, dbErr) {
		t.Fatalf("unexpected ListByGSI error: %v", err)
	}

	err = repo.Delete(context.Background(), "user-1", "2026-04-20", "tx-1")
	if err == nil || !strings.Contains(err.Error(), "delete transaction") || !errors.Is(err, dbErr) {
		t.Fatalf("unexpected Delete error: %v", err)
	}

	_, err = repo.ListWithinDateRange(context.Background(), "", time.Now(), time.Now())
	if err == nil || err.Error() != "owner ID is required" {
		t.Fatalf("unexpected ListWithinDateRange validation error: %v", err)
	}

	_, err = repo.GetByKey(context.Background(), "")
	if err == nil || err.Error() != "ID is required" {
		t.Fatalf("unexpected GetByKey validation error: %v", err)
	}

	if fmt.Sprintf("%v", repositories.ErrTransactionNotFound) == "" {
		t.Fatalf("ErrTransactionNotFound should be defined")
	}
}
