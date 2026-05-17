variable "lambda_role_arn" {
  description = "IAM role ARN used by the create transaction Lambda"
  type        = string
}

variable "table_name" {
  description = "DynamoDB table name exposed to Lambda"
  type        = string
}
