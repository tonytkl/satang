data "aws_caller_identity" "current" {}

resource "aws_iam_role" "create_transaction_lambda_role" {
  name = "satang-create-transaction-lambda-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "lambda.amazonaws.com"
        }
      }
    ]
  })
}

resource "aws_iam_role_policy_attachment" "create_transaction_lambda_basic_execution" {
  role       = aws_iam_role.create_transaction_lambda_role.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
}

resource "aws_iam_role_policy" "create_transaction_lambda_dynamodb_access" {
  name = "satang-create-transaction-lambda-dynamodb-access"
  role = aws_iam_role.create_transaction_lambda_role.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "dynamodb:PutItem"
        ]
        Resource = [
          aws_dynamodb_table.dynamodb_table.arn
        ]
      }
    ]
  })
}


resource "aws_apigatewayv2_api" "satang_api" {
  name          = "satang-api"
  protocol_type = "HTTP"
}

resource "aws_apigatewayv2_integration" "create_transaction_lambda" {
  api_id                 = aws_apigatewayv2_api.satang_api.id
  integration_type       = "AWS_PROXY"
  integration_uri        = module.lambda_transaction.create_transaction_invoke_arn
  integration_method     = "POST"
  payload_format_version = "2.0"
}

resource "aws_apigatewayv2_route" "create_transaction" {
  api_id    = aws_apigatewayv2_api.satang_api.id
  route_key = "POST /api/v1.0/transactions"
  target    = "integrations/${aws_apigatewayv2_integration.create_transaction_lambda.id}"
}

resource "aws_apigatewayv2_stage" "default" {
  api_id      = aws_apigatewayv2_api.satang_api.id
  name        = "$default"
  auto_deploy = true
}

resource "aws_lambda_permission" "allow_apigateway_invoke_create_transaction" {
  statement_id  = "AllowExecutionFromAPIGatewayCreateTransaction"
  action        = "lambda:InvokeFunction"
  function_name = module.lambda_transaction.create_transaction_function_name
  principal     = "apigateway.amazonaws.com"
  source_arn    = "${aws_apigatewayv2_api.satang_api.execution_arn}/*/*"
}
