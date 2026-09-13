output "invoke_arn" {
  description = "Invoke ARN for the Lambda function"
  value       = aws_lambda_function.this.invoke_arn
}

output "function_name" {
  description = "Function name for the Lambda function"
  value       = aws_lambda_function.this.function_name
}
