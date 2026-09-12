resource "aws_apigatewayv2_api" "satang_api" {
  name          = "satang-api"
  protocol_type = "HTTP"
}

locals {
  api_routes = {
    create_transaction = {
      method                  = "POST"
      path                    = "/api/v1.0/transactions"
      integration_uri         = module.lambda_create_transaction.create_transaction_invoke_arn
      lambda_function_name    = module.lambda_create_transaction.create_transaction_function_name
      permission_statement_id = "AllowExecutionFromAPIGatewayCreateTransaction"
    }
    get_transaction = {
      method                  = "GET"
      path                    = "/api/v1.0/transactions/{transaction_id}"
      integration_uri         = module.lambda_get_transaction.get_transaction_invoke_arn
      lambda_function_name    = module.lambda_get_transaction.get_transaction_function_name
      permission_statement_id = "AllowExecutionFromAPIGatewayGetTransaction"
    }
    list_transactions = {
      method                  = "GET"
      path                    = "/api/v1.0/transactions"
      integration_uri         = module.lambda_list_transactions.list_transactions_invoke_arn
      lambda_function_name    = module.lambda_list_transactions.list_transactions_function_name
      permission_statement_id = "AllowExecutionFromAPIGatewayListTransactions"
    }
  }
}

module "api_routes" {
  for_each = local.api_routes

  source = "./api_route/create_transaction"

  api_id                         = aws_apigatewayv2_api.satang_api.id
  api_execution_arn              = aws_apigatewayv2_api.satang_api.execution_arn
  integration_uri                = each.value.integration_uri
  integration_method             = each.value.method
  route_key                      = "${each.value.method} ${each.value.path}"
  lambda_function_name           = each.value.lambda_function_name
  lambda_permission_statement_id = each.value.permission_statement_id
}

resource "aws_apigatewayv2_stage" "default" {
  api_id      = aws_apigatewayv2_api.satang_api.id
  name        = "$default"
  auto_deploy = true
}
