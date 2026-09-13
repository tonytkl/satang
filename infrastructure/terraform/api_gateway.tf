resource "aws_apigatewayv2_api" "satang_api" {
  name          = "satang-api"
  protocol_type = "HTTP"
}

locals {
  transaction_routes = {
    create_transaction = {
      method                  = "POST"
      path                    = "/api/v1.0/transactions"
      integration_uri         = module.lambdas["create_transaction"].invoke_arn
      lambda_function_name    = module.lambdas["create_transaction"].function_name
      permission_statement_id = "AllowExecutionFromAPIGatewayCreateTransaction"
    }
    get_transaction = {
      method                  = "GET"
      path                    = "/api/v1.0/transactions/{transaction_id}"
      integration_uri         = module.lambdas["get_transaction"].invoke_arn
      lambda_function_name    = module.lambdas["get_transaction"].function_name
      permission_statement_id = "AllowExecutionFromAPIGatewayGetTransaction"
    }
    list_transactions = {
      method                  = "GET"
      path                    = "/api/v1.0/transactions"
      integration_uri         = module.lambdas["list_transactions"].invoke_arn
      lambda_function_name    = module.lambdas["list_transactions"].function_name
      permission_statement_id = "AllowExecutionFromAPIGatewayListTransactions"
    }
  }

  # Keep wallet routes as a separate map so they can be enabled when wallet
  # Lambda functions are added to local.lambda_functions.
  wallet_routes = {}

  api_routes = merge(local.transaction_routes, local.wallet_routes)
}

module "api_routes" {
  for_each = local.api_routes

  source = "./api_route/http_route"

  api_id                         = aws_apigatewayv2_api.satang_api.id
  api_execution_arn              = aws_apigatewayv2_api.satang_api.execution_arn
  integration_uri                = each.value.integration_uri
  integration_method             = "POST"
  route_key                      = "${each.value.method} ${each.value.path}"
  lambda_function_name           = each.value.lambda_function_name
  lambda_permission_statement_id = each.value.permission_statement_id
}

resource "aws_apigatewayv2_stage" "default" {
  api_id      = aws_apigatewayv2_api.satang_api.id
  name        = "$default"
  auto_deploy = true
}
