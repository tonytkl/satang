output "create_transaction_invoke_arn" {
  description = "Invoke ARN for create transaction Lambda"
  value       = aws_lambda_function.create_transaction.invoke_arn
}

output "create_transaction_function_name" {
  description = "Function name for create transaction Lambda"
  value       = aws_lambda_function.create_transaction.function_name
}
