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

  attribute {
    name = "GSI2_PK"
    type = "S"
  }

  attribute {
    name = "GSI2_SK"
    type = "S"
  }

  attribute {
    name = "GSI3_PK"
    type = "S"
  }

  attribute {
    name = "GSI3_SK"
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
  global_secondary_index {
    name               = "GSI2"
    hash_key           = "GSI2_PK"
    range_key          = "GSI2_SK"
    projection_type    = "ALL"
    read_capacity      = 2
    write_capacity     = 2
  }
  global_secondary_index {
    name               = "GSI3"
    hash_key           = "GSI3_PK"
    range_key          = "GSI3_SK"
    projection_type    = "ALL"
    read_capacity      = 2
    write_capacity     = 2
  }
}