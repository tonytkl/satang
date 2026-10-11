# Satang App

The `app` directory contains the Go module for Satang's application logic and data layer. It provides:
- Domain packages (e.g. `transaction`, `wallet`, `category`), each containing its own model, repository, service, and (where applicable) schema.
- A DynamoDB client wrapper built on AWS SDK v2.
- A generic base repository shared by domains.
- AWS Lambda handler implementations.

## Module

- Module path: `github.com/tonytkl/satang`
- Go version: `1.25.7`

## Directory Overview

The code is organized by domain rather than by technical layer.

```text
app/
├── <domain>/               # One package per domain (transaction, wallet, category, ...)
│   ├── <domain>.go         # Domain model and key generation
│   ├── repository.go       # Data access
│   ├── service.go          # Business logic
│   └── schema.go           # Request/response schemas (where applicable)
├── repository/             # Generic BaseRepository shared by domains
├── clients/                # AWS SDK client wrappers (DynamoDB)
├── lambda/<domain>/<op>/   # Lambda handler entry points (main.go) per operation
├── models/                 # Shared models
├── utils/                  # Helpers (API Gateway, dates, keys, constants)
└── go.mod
```

## Go API Conventions

- Return value slices (e.g. `[]Wallet`) for collections and values for single-resource fetches.
- Keep conversion logic at the repository boundary when persistence uses pointer-backed internals.
- Domain repositories build on the shared `BaseRepository[T]`.

## Local Development

### 1. Start DynamoDB Local

From the repository root, run:

```bash
make local-dynamodb-up
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

The project uses a single-table design pattern with DynamoDB, utilizing prefixed partition/sort keys and Global Secondary Indexes (GSIs) for flexible querying.

**Key Design:**
- `PK` (Partition Key) — Primary access pattern
- `SK` (Sort Key) — Primary sort/range pattern
- `GSI1_PK` / `GSI1_SK`, `GSI2_PK` / `GSI2_SK`, `GSI3_PK` / `GSI3_SK` — Secondary access patterns via Global Secondary Indexes

**Naming Conventions:**
Keys are prefixed with entity type for clarity. For example:
- Entity partition keys use format: `<ENTITY_TYPE>#<id>`
- Entity sort keys use format: `<ENTITY_TYPE>#<date-or-identifier>`
- Index keys enable queries by specific attributes (e.g., by category, wallet, or entity ID)

Refer to each domain package for its specific key scheme.

## Core Components

### Domain packages (`<domain>/`)
- Self-contained: model, repository, service, and schema for one domain.
- Adding a domain means adding a new package, plus handlers under `lambda/<domain>/`.

### Shared repository (`repository/`)
- `BaseRepository[T]` provides generic `Save`, `Get`, `List`, `Update`, `Delete`, and `PrepareSave` for any `SatangModel`.

### Clients (`clients/`)
- `DynamoDBClient` interface (`PutItem`, `GetItem`, `DeleteItem`, `QueryItems`, `ScanItems`) with `NewDynamoDBClient` and `NewDynamoDBClientWithConfig` constructors.
- Supports `.env` configuration and local endpoint override.

### AWS Lambda (`lambda/`)
- One `main.go` per operation, grouped by domain (`lambda/<domain>/<operation>/`).
- Handles API Gateway request/response marshaling and error handling.

### Utilities (`utils/`)
- `api_gateway_helper.go`, `date_helper.go`, `dynamodb_key_helper.go`, `constants.go`.
