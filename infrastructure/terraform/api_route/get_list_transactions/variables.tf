variable "api_id" {
  description = "HTTP API Gateway ID"
  type        = string
}

variable "api_execution_arn" {
  description = "HTTP API Gateway execution ARN"
  type        = string
}

variable "integration_uri" {
  description = "Lambda invoke ARN used as API integration URI"
  type        = string
}

variable "integration_method" {
  description = "HTTP method used by API Gateway when invoking integration"
  type        = string
  default     = "POST"
}

variable "route_key" {
  description = "Route key in format '<METHOD> <PATH>'"
  type        = string
}

variable "lambda_function_name" {
  description = "Lambda function name allowed to be invoked by API Gateway"
  type        = string
}

variable "lambda_permission_statement_id" {
  description = "Unique statement ID for Lambda invoke permission"
  type        = string
}
