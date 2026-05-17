output "get_transaction_invoke_arn" {
  description = "Invoke ARN for get transaction Lambda"
  value       = aws_lambda_function.get_transaction.invoke_arn
}

output "get_transaction_function_name" {
  description = "Function name for get transaction Lambda"
  value       = aws_lambda_function.get_transaction.function_name
}
