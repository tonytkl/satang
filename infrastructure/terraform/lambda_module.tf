module "lambda_create_transaction" {
  source          = "./lambda/create_transaction"
  lambda_role_arn = aws_iam_role.lambda_role.arn
  table_name      = aws_dynamodb_table.dynamodb_table.name

  depends_on = [
    aws_iam_role_policy_attachment.lambda_basic_execution,
    aws_iam_role_policy.lambda_dynamodb_access
  ]
}

module "lambda_get_transaction" {
  source          = "./lambda/get_transaction"
  lambda_role_arn = aws_iam_role.lambda_role.arn
  table_name      = aws_dynamodb_table.dynamodb_table.name

  depends_on = [
    aws_iam_role_policy_attachment.lambda_basic_execution,
    aws_iam_role_policy.lambda_dynamodb_access
  ]
}

