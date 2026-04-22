# Satang App

The `app` directory contains the Go module for Satang's data layer. It provides:
- Domain models for users, wallets, categories, and transactions.
- A DynamoDB client wrapper built on AWS SDK v2.
- Repository implementations for transaction data access.
- Unit tests for client behavior.

## Module

- Module path: `github.com/tonytkl/satang`
- Go version: `1.25.7`

## Directory Overview

```text
app/
├── clients/                # AWS client wrappers (DynamoDB)
├── model/                  # Domain models and key conventions
├── repositories/           # Data access layer
├── tests/
│   ├── test_clients/       # DynamoDB client tests
│   └── test_utils/         # Test helpers/assertions
├── utils/                  # Key/UUID helpers
├── docker/
│   └── dynamodb/           # Local DynamoDB persistence volume
├── docker-compose.yml      # Local DynamoDB service
└── go.mod
```

## Local Development

### 1. Start DynamoDB Local

```bash
docker compose up -d
```

This starts `amazon/dynamodb-local` on port `8000`.

### 2. Configure Local Mode

Set environment variables before running app code that initializes the DynamoDB client:

```bash
export ENVIRONMENT=local
export AWS_DYNAMODB_ENDPOINT=http://localhost:8000
```

If `AWS_DYNAMODB_ENDPOINT` is omitted while in local mode, the client defaults to `http://localhost:8000`.

### 3. Run Tests

```bash
go test ./...
```

## Data Modeling

The project uses a single-table style key design with prefixed partition/sort keys and GSIs.

Primary keys and GSIs used by transactions:
- `PK`
- `SK`
- `GSI_PK` / `GSI_SK` (define different access)

Examples:
- User partition: `USER#<owner-id>`
- Transaction sort key: `TX#<yyyy-mm-dd>#<transaction-id>`
- Category index key: `TX_CATEGORY#<category-id>`
- Wallet index key: `TX_WALLET#<wallet-id>`
- Transaction-id index key: `TX_ID#<transaction-id>`

## Core Components

- `clients/dynamodb.go`
	- Defines `DynamoDBClient` interface (`PutItem`, `GetItem`, `DeleteItem`, `QueryItems`, `ScanItems`).
	- Provides `NewDynamoDBClient` and `NewDynamoDBClientWithConfig`.
	- Supports `.env` loading and local endpoint override.

- `repositories/transaction_repository.go`
	- Main repository used for transaction create/get/list flows.
	- Supports querying by GSI and date ranges.

- `model/*.go`
	- Contains entity structs and key conventions.

## Notes

- The repository currently includes `_transactionRepository` and `transactionRepository` variants; the non-underscore `TransactionRepository` is the primary exported API.
- `utils.GetSortingKey` currently formats keys as `TX#<date>` (without ID suffix), while some model constructors include ID in transaction SKs. Keep this in mind when evolving key strategies.
