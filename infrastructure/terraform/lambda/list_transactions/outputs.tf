output "list_transactions_invoke_arn" {
  description = "Invoke ARN for list transactions Lambda"
  value       = aws_lambda_function.list_transactions.invoke_arn
}

output "list_transactions_function_name" {
  description = "Function name for list transactions Lambda"
  value       = aws_lambda_function.list_transactions.function_name
}
