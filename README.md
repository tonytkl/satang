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
│   ├── tests/
│   ├── docker/
│   ├── docker-compose.yml
│   └── go.mod
├── infrastructure/
│   └── terraform/
├── aws/
└── README.md
```

## Prerequisites

- Go `1.25.7+`
- Docker (for local DynamoDB)
- AWS credentials/profile configured (for real AWS access)
- Terraform `1.5+` (for infrastructure deployment)

## Quick Start

1. Start local DynamoDB:

```bash
cd app
docker compose up -d
```

2. Run Go tests:

```bash
cd app
go test ./...
```

3. (Optional) Provision DynamoDB in AWS with Terraform:

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
