# Satang App

The `app` directory contains the Go module for Satang's application logic and data layer. It provides:
- Domain models for core entities (users, wallets, categories, transactions).
- A DynamoDB client wrapper built on AWS SDK v2.
- Repository layer for data access patterns.
- Service layer for business logic.
- AWS Lambda handler implementations.

## Module

- Module path: `github.com/tonytkl/satang`
- Go version: `1.25.7`

## Directory Overview

```text
app/
├── clients/                # AWS SDK client wrappers (DynamoDB)
├── model/                  # Domain models and entity definitions
├── repositories/           # Data access layer and query patterns
├── services/               # Business logic layer
├── lambda/                 # AWS Lambda handler implementations
├── utils/                  # Helper utilities (key generation, date handling, etc.)
└── go.mod
```

## Local Development

### 1. Start DynamoDB Local

From the repository root, run:

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

The project uses a single-table design pattern with DynamoDB, utilizing prefixed partition/sort keys and Global Secondary Indexes (GSIs) for flexible querying.

**Key Design:**
- `PK` (Partition Key) — Primary access pattern
- `SK` (Sort Key) — Primary sort/range pattern
- `GSI_PK` / `GSI_SK` — Secondary access patterns via Global Secondary Indexes

**Naming Conventions:**
Keys are prefixed with entity type for clarity. For example:
- Entity partition keys use format: `<ENTITY_TYPE>#<id>`
- Entity sort keys use format: `<ENTITY_TYPE>#<date-or-identifier>`
- Index keys enable queries by specific attributes (e.g., by category, wallet, or entity ID)

Refer to individual model definitions in `model/` and repository implementations in `repositories/` for specific key schemes.

## Core Components

### Clients (`clients/`)
- **`dynamodb.go`** — DynamoDB client wrapper
  - `DynamoDBClient` interface with standard operations: `PutItem`, `GetItem`, `DeleteItem`, `QueryItems`, `ScanItems`
  - `NewDynamoDBClient` and `NewDynamoDBClientWithConfig` constructors
  - Supports `.env` configuration and local endpoint override

### Data Layer (`repositories/`, `model/`)
- **`repositories/`** — Data access implementations
  - Entity repository patterns for query/write operations
  - Supports GSI queries, date range filtering, and batch operations
- **`model/`** — Domain entity definitions
  - Core entity structs (User, Wallet, Category, etc.)
  - Entity-specific key generation and validation

### Business Layer (`services/`)
- **`services/`** — Business logic and orchestration
  - Service implementations for domain operations
  - Composition of repositories and utilities

### AWS Integration (`lambda/`)
- **`lambda/`** — AWS Lambda handler entry points
  - HTTP endpoint handlers with API Gateway integration
  - Request/response marshaling and error handling

### Utilities (`utils/`)
- **`api_gateway_helper.go`** — API Gateway request/response utilities
- **`date_helper.go`** — Date and time utilities
- **`dynamodb_key_helper.go`** — Key generation and formatting helpers
