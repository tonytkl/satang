# Satang Infrastructure

The `infrastructure` directory contains Terraform code for provisioning Satang AWS resources.

Current scope:
- One DynamoDB table with composite primary key (`PK`, `SK`)
- Three global secondary indexes (`GSI1`, `GSI2`, `GSI3`)

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

## Variables

Defined in `variables.tf`:

- `region` (default: `ap-southeast-1`)
- `table_name` (default: `satang-dynamodb`)

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
