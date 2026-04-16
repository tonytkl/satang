variable "region" {
  description = "AWS region to deploy resources"
  type        = string
  default     = "ap-southeast-1"
}

variable "table_name" {
  description = "Name of the DynamoDB table"
  type        = string
  default     = "satang-dynamodb"
}