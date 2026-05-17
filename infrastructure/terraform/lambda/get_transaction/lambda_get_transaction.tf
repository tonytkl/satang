resource "aws_lambda_function" "get_transaction" {
  function_name = "satang-get-transaction"
  role          = var.lambda_role_arn
  runtime       = "provided.al2023"
  handler       = "bootstrap"
  filename      = "../../aws/lambda/get_transaction.zip"

  source_code_hash = filebase64sha256("../../aws/lambda/get_transaction.zip")

  environment {
    variables = {
      TABLE_NAME = var.table_name
    }
  }
}
