output "dynamodb_table_arn" {
  description = "ARN of the DynamoDB table"
  value       = aws_dynamodb_table.dynamodb_table.arn
}

output "dynamodb_table_name" {
  description = "Name of the DynamoDB table"
  value       = aws_dynamodb_table.dynamodb_table.name
}

output "api_base_url" {
  description = "Base URL for Satang HTTP API"
  value       = aws_apigatewayv2_stage.default.invoke_url
}

output "create_transaction_endpoint" {
  description = "Endpoint URL for creating transactions"
  value       = "${aws_apigatewayv2_stage.default.invoke_url}/api/v1.0/transactions"
}