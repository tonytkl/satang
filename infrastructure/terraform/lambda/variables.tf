variable "function_name" {
  description = "Lambda function name"
  type        = string
}

variable "artifact_path" {
  description = "Path to the Lambda zip artifact"
  type        = string
}

variable "lambda_role_arn" {
  description = "IAM role ARN used by the Lambda function"
  type        = string
}

variable "table_name" {
  description = "DynamoDB table name exposed to Lambda"
  type        = string
}
