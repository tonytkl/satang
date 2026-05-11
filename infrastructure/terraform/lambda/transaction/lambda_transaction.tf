resource "aws_lambda_function" "create_transaction" {
  function_name = "satang-create-transaction"
  role          = var.lambda_role_arn
  runtime       = "provided.al2023"
  handler       = "bootstrap"
  filename      = var.create_transaction_lambda_zip_path

  source_code_hash = filebase64sha256(var.create_transaction_lambda_zip_path)

  environment {
    variables = {
      TABLE_NAME = var.table_name
    }
  }
}