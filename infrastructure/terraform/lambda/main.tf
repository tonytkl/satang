resource "aws_lambda_function" "this" {
  function_name = var.function_name
  role          = var.lambda_role_arn
  runtime       = "provided.al2023"
  handler       = "bootstrap"
  filename      = var.artifact_path

  source_code_hash = filebase64sha256(var.artifact_path)

  environment {
    variables = {
      TABLE_NAME = var.table_name
    }
  }
}
