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

variable "create_transaction_lambda_zip_path" {
  description = "Path to the create-transaction Lambda zip artifact"
  type        = string
  default     = "../../aws/lambda/create_transaction.zip"
}

variable "environment" {
  description = "Deployment environment. Use 'local' to target local service endpoints"
  type        = string
  default     = "local"
}

variable "local_endpoints" {
  description = "Service endpoints used when environment is set to local"
  type        = map(string)
  default = {
    dynamodb   = "http://localhost:8000"
    lambda     = "http://localhost:4566"
    apigateway = "http://localhost:4566"
  }
}

variable "access_key" {
  description = "AWS access key"
  type        = string
  default     = ""
  sensitive   = true
}

variable "secret_key" {
  description = "AWS secret key"
  type        = string
  default     = ""
  sensitive   = true
}