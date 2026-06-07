output "get_list_transactions_invoke_arn" {
  description = "Invoke ARN for get list transactions Lambda"
  value       = aws_lambda_function.get_list_transactions.invoke_arn
}

output "get_list_transactions_function_name" {
  description = "Function name for get list transactions Lambda"
  value       = aws_lambda_function.get_list_transactions.function_name
}
