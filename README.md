# Satang

Satang is a Go-based backend workspace for personal finance data built on DynamoDB.

The repository is split into:
- `app/`: Go module with data models, DynamoDB client wrapper, repository logic, and tests.
- `infrastructure/`: Terraform configuration to provision DynamoDB resources in AWS.
- `aws/`: Reserved directory for additional AWS-related assets.

## Repository Structure

```text
.
├── app/
│   ├── clients/
│   ├── model/
│   ├── repositories/
│   ├── services/
│   ├── lambda/
│   └── go.mod
├── scripts/
├── docker-compose.local.yml
├── template.sam.yaml
├── infrastructure/
│   └── terraform/
├── aws/
└── README.md
```

## Prerequisites

- Go `1.25.7+`
- Docker (for local DynamoDB)
- AWS CLI v2 (for local DynamoDB table initialization)
- AWS SAM CLI (for local API Gateway + Lambda emulation)
- AWS credentials/profile configured (for real AWS access)
- Terraform `1.5+` (for infrastructure deployment)

## Quick Start

1. Prepare local dependencies for SAM (DynamoDB Local + table init + SAM build):

```bash
make local-up
```

2. Start local API Gateway + Lambda (keep this running in another terminal):

```bash
make sam-local-api
```

3. Run Go tests:

```bash
cd app
go test ./...
```

4. (Optional) Provision DynamoDB in AWS with Terraform:

```bash
cd infrastructure/terraform
terraform init
terraform plan
terraform apply
```

## Environment Variables (App)

The DynamoDB client supports local mode with these variables:

- `ENVIRONMENT=local`
- `AWS_DYNAMODB_ENDPOINT=http://localhost:8000` (optional, defaults to localhost in local mode)

Without local mode, the app uses AWS SDK default credential/provider chain.

## Additional Documentation

- App usage and internals: `app/README.md`
- Terraform and deployment details: `infrastructure/README.md`

## Local Development with SAM (API Gateway -> Lambda -> DynamoDB Local)

This flow gives you a local HTTP endpoint that emulates API Gateway and triggers your Lambda handlers, while persisting data in DynamoDB Local.

### Local development start commands

Use two terminals from repository root:

Terminal 1:

```bash
make local-up
```

Terminal 2:

```bash
make sam-local-api
```

After startup, use:

```text
http://127.0.0.1:3000
```

### 1. Start local dependencies

From repository root:

```bash
make local-up
```

What this does:
- Starts DynamoDB Local (`localhost:8000`) via Docker
- Persists DynamoDB data in Docker volume `satang_dynamodb-data`
- Creates `satang-dynamodb` table (with `GSI1`, `GSI2`, `GSI3`) if missing
- Builds SAM artifacts into `.aws-sam/build/`

To reset local DynamoDB data, remove the volume:

```bash
docker volume rm satang_dynamodb-data
```

### 2. Run local API Gateway + Lambda

In a separate terminal:

```bash
make sam-local-api
```

API base URL:

```text
http://127.0.0.1:3000
```

### 3. Send HTTP request to create transaction

```bash
curl -i -X POST "http://127.0.0.1:3000/transactions" \
	-H "Content-Type: application/json" \
	-d '{
		"walletId": "wallet-1",
		"walletName": "Main Wallet",
		"categoryId": "category-1",
		"categoryName": "Food",
		"description": "Lunch",
		"currency": "THB",
		"imageUrl": "",
		"type": "expense",
		"amount": 120.5,
		"date": "2026-06-13"
	}'
```

Expected response: `201 Created`

### 4. Query data through Lambda endpoint

```bash
curl -s "http://127.0.0.1:3000/transactions?fromDate=2026-06-01&toDate=2026-06-30" | jq
```

### 5. Verify records directly in DynamoDB Local (optional)

```bash
AWS_ACCESS_KEY_ID=local AWS_SECRET_ACCESS_KEY=local AWS_DEFAULT_REGION=ap-southeast-1 \
aws dynamodb scan --table-name satang-dynamodb --endpoint-url http://localhost:8000
```

### 6. Stop local services

```bash
make local-dynamodb-down
```

To stop SAM local API, press `Ctrl+C` in the terminal running `make sam-local-api`.
