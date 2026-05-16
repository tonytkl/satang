# Satang Infrastructure

The `infrastructure` directory contains Terraform code for provisioning Satang AWS resources.

Current scope:
- One DynamoDB table with composite primary key (`PK`, `SK`)
- Three global secondary indexes (`GSI1`, `GSI2`, `GSI3`)
- One Lambda function (`satang-create-transaction`)
- One API Gateway HTTP API endpoint (`POST /api/v1.0/transactions`)

Terraform files live in `infrastructure/terraform/`.

## Prerequisites

- Terraform `1.5+`
- AWS credentials configured in your shell/profile
- Access to the target AWS account and region

## Terraform Structure

```text
infrastructure/terraform/
├── main.tf            # Provider setup
├── variables.tf       # Input variables
├── dynamodb.tf        # DynamoDB table + GSIs
├── outputs.tf         # Exported outputs
└── terraform.tfvars   # Variable values (environment-specific)
```

## Configured Resources

### DynamoDB Table

Resource: `aws_dynamodb_table.dynamodb_table`

- Table name: from variable `table_name`
- Billing mode: `PROVISIONED`
- Read capacity: `2`
- Write capacity: `2`
- Hash key: `PK`
- Range key: `SK`

### Global Secondary Indexes

- `GSI1` on `GSI_PK` + `GSI_SK`
- `GSI2` on `GSI2_PK` + `GSI2_SK`
- `GSI3` on `GSI3_PK` + `GSI3_SK`

All GSIs use `projection_type = "ALL"` with provisioned read/write capacity `2`.

### Create Transaction API

Terraform provisions:

- IAM role and policies for Lambda execution + DynamoDB `PutItem`
- Lambda function using custom runtime (`provided.al2023`)
- API Gateway HTTP API with route `POST /api/v1.0/transactions`
- Lambda invoke permission for API Gateway

Lambda environment variable:

- `TABLE_NAME` (wired to provisioned DynamoDB table)

Lambda artifact location variable:

- `create_transaction_lambda_zip_path` (default: `../../aws/lambda/create_transaction.zip`)

Build the Lambda binary and zip artifact before `terraform apply`:

```bash
mkdir -p aws/lambda
cd app
GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -o ../aws/bootstrap ./cmd/lambda/create_transaction
cd ../aws
zip -j lambda/create_transaction.zip bootstrap
```

## Variables

Defined in `variables.tf`:

- `region` (default: `ap-southeast-1`)
- `table_name` (default: `satang-dynamodb`)
- `create_transaction_lambda_zip_path` (default: `../../aws/lambda/create_transaction.zip`)

Current values in `terraform.tfvars`:

```hcl
region     = "ap-southeast-1"
table_name = "satang-dynamodb"
```

## Deploy

From `infrastructure/terraform`:

```bash
terraform init
terraform fmt
terraform validate
terraform plan
terraform apply
```

## Destroy

```bash
terraform destroy
```

## Outputs

After apply, Terraform exports:
- `dynamodb_table_arn`
- `dynamodb_table_name`
- `api_base_url`
- `create_transaction_endpoint`
