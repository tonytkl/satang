resource "aws_lambda_function" "get_list_transactions" {
  function_name = "satang-get-list-transactions"
  role          = var.lambda_role_arn
  runtime       = "provided.al2023"
  handler       = "bootstrap"
  filename      = "../../aws/lambda/get_list_transactions.zip"

  source_code_hash = filebase64sha256("../../aws/lambda/get_list_transactions.zip")

  environment {
    variables = {
      TABLE_NAME = var.table_name
    }
  }
}
