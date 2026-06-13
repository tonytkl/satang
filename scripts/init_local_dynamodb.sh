#!/usr/bin/env bash
set -euo pipefail

TABLE_NAME="${TABLE_NAME:-satang-dynamodb}"
DYNAMODB_ENDPOINT="${AWS_DYNAMODB_ENDPOINT:-http://localhost:8000}"
AWS_REGION="${AWS_REGION:-ap-southeast-1}"

export AWS_ACCESS_KEY_ID="${AWS_ACCESS_KEY_ID:-local}"
export AWS_SECRET_ACCESS_KEY="${AWS_SECRET_ACCESS_KEY:-local}"
export AWS_DEFAULT_REGION="$AWS_REGION"
# Prevent AWS CLI from reading malformed host-level profile files during local runs.
export AWS_SHARED_CREDENTIALS_FILE="${AWS_SHARED_CREDENTIALS_FILE:-/dev/null}"
export AWS_CONFIG_FILE="${AWS_CONFIG_FILE:-/dev/null}"

aws_local() {
  aws --no-cli-pager --cli-connect-timeout 3 --cli-read-timeout 3 "$@"
}

wait_for_dynamodb() {
  local attempts=20
  local delay_seconds=1

  for ((i=1; i<=attempts; i++)); do
    if aws_local dynamodb list-tables --endpoint-url "$DYNAMODB_ENDPOINT" >/dev/null 2>&1; then
      return 0
    fi
    sleep "$delay_seconds"
  done

  return 1
}

echo "Checking DynamoDB table: $TABLE_NAME at $DYNAMODB_ENDPOINT"

if ! wait_for_dynamodb; then
  echo "DynamoDB endpoint is not ready: $DYNAMODB_ENDPOINT"
  exit 1
fi

if aws_local dynamodb describe-table \
  --table-name "$TABLE_NAME" \
  --endpoint-url "$DYNAMODB_ENDPOINT" \
  >/dev/null 2>&1; then
  echo "Table already exists: $TABLE_NAME"
  exit 0
fi

echo "Creating table: $TABLE_NAME"
create_output="$(aws_local dynamodb create-table \
  --table-name "$TABLE_NAME" \
  --endpoint-url "$DYNAMODB_ENDPOINT" \
  --billing-mode PAY_PER_REQUEST \
  --attribute-definitions \
    AttributeName=PK,AttributeType=S \
    AttributeName=SK,AttributeType=S \
    AttributeName=GSI_PK,AttributeType=S \
    AttributeName=GSI_SK,AttributeType=S \
    AttributeName=GSI2_PK,AttributeType=S \
    AttributeName=GSI2_SK,AttributeType=S \
    AttributeName=GSI3_PK,AttributeType=S \
    AttributeName=GSI3_SK,AttributeType=S \
  --key-schema \
    AttributeName=PK,KeyType=HASH \
    AttributeName=SK,KeyType=RANGE \
  --global-secondary-indexes '[
    {
      "IndexName": "GSI1",
      "KeySchema": [
        {"AttributeName": "GSI_PK", "KeyType": "HASH"},
        {"AttributeName": "GSI_SK", "KeyType": "RANGE"}
      ],
      "Projection": {"ProjectionType": "ALL"}
    },
    {
      "IndexName": "GSI2",
      "KeySchema": [
        {"AttributeName": "GSI2_PK", "KeyType": "HASH"},
        {"AttributeName": "GSI2_SK", "KeyType": "RANGE"}
      ],
      "Projection": {"ProjectionType": "ALL"}
    },
    {
      "IndexName": "GSI3",
      "KeySchema": [
        {"AttributeName": "GSI3_PK", "KeyType": "HASH"},
        {"AttributeName": "GSI3_SK", "KeyType": "RANGE"}
      ],
      "Projection": {"ProjectionType": "ALL"}
    }
  ]' 2>&1)" || true

if [[ "$create_output" == *"ResourceInUseException"* ]]; then
  echo "Table already exists: $TABLE_NAME"
elif [[ -n "$create_output" ]]; then
  echo "$create_output"
fi

aws_local dynamodb wait table-exists \
  --table-name "$TABLE_NAME" \
  --endpoint-url "$DYNAMODB_ENDPOINT"

echo "DynamoDB table is ready: $TABLE_NAME"
