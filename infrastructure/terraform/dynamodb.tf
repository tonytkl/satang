resource "aws_dynamodb_table" "dynamodb_table" {
  name           = var.table_name
  billing_mode   = "PROVISIONED"
  read_capacity  = 2
  write_capacity = 2
  hash_key       = "PK"
  range_key      = "SK"

  attribute {
    name = "PK"
    type = "S"
  }

  attribute {
    name = "SK"
    type = "S"
  }

  attribute {
    name = "GSI_PK"
    type = "S"
  }

  attribute {
    name = "GSI_SK"
    type = "S"
  }

  global_secondary_index {
    name               = "GSI1"
    hash_key           = "GSI_PK"
    range_key          = "GSI_SK"
    projection_type    = "ALL"
    read_capacity      = 2
    write_capacity     = 2
  }
}