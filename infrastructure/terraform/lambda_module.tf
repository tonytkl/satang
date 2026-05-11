module "lambda_transaction" {
  source = "./lambda/transaction"

  create_transaction_lambda_zip_path = var.create_transaction_lambda_zip_path
  lambda_role_arn                    = aws_iam_role.create_transaction_lambda_role.arn
  table_name                         = aws_dynamodb_table.dynamodb_table.name

  depends_on = [
    aws_iam_role_policy_attachment.create_transaction_lambda_basic_execution,
    aws_iam_role_policy.create_transaction_lambda_dynamodb_access
  ]
}
