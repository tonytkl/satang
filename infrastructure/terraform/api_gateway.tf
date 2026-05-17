resource "aws_apigatewayv2_api" "satang_api" {
  name          = "satang-api"
  protocol_type = "HTTP"
}

module "api_route_create_transaction" {
  source = "./api_route/create_transaction"

  api_id                         = aws_apigatewayv2_api.satang_api.id
  api_execution_arn              = aws_apigatewayv2_api.satang_api.execution_arn
  integration_uri                = module.lambda_create_transaction.create_transaction_invoke_arn
  integration_method             = "POST"
  route_key                      = "POST /api/v1.0/transactions"
  lambda_function_name           = module.lambda_create_transaction.create_transaction_function_name
  lambda_permission_statement_id = "AllowExecutionFromAPIGatewayCreateTransaction"
}

module "api_route_get_transaction" {
  source = "./api_route/get_transaction"

  api_id                         = aws_apigatewayv2_api.satang_api.id
  api_execution_arn              = aws_apigatewayv2_api.satang_api.execution_arn
  integration_uri                = module.lambda_get_transaction.get_transaction_invoke_arn
  integration_method             = "GET"
  route_key                      = "GET /api/v1.0/transactions/{transaction_id}"
  lambda_function_name           = module.lambda_get_transaction.get_transaction_function_name
  lambda_permission_statement_id = "AllowExecutionFromAPIGatewayGetTransaction"
}

resource "aws_apigatewayv2_stage" "default" {
  api_id      = aws_apigatewayv2_api.satang_api.id
  name        = "$default"
  auto_deploy = true
}
