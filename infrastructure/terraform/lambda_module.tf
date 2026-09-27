locals {
  transaction_functions = {
    # Transactions
    create_transaction = {
      function_name = "satang-create-transaction"
      artifact_path = "../../aws/lambda/create_transaction.zip"
    }
    get_transaction = {
      function_name = "satang-get-transaction"
      artifact_path = "../../aws/lambda/get_transaction.zip"
    }
    list_transactions = {
      function_name = "satang-list-transactions"
      artifact_path = "../../aws/lambda/list_transactions.zip"
    }
  }
  wallet_functions = {
    # Wallets
    create_wallet = {
      function_name = "satang-create-wallet"
      artifact_path = "../../aws/lambda/create_wallet.zip"
    }
    get_wallet = {
      function_name = "satang-get-wallet"
      artifact_path = "../../aws/lambda/get_wallet.zip"
    }
    list_wallets = {
      function_name = "satang-list-wallets"
      artifact_path = "../../aws/lambda/list_wallets.zip"
    }
    edit_wallet = {
      function_name = "satang-edit-wallet"
      artifact_path = "../../aws/lambda/edit_wallet.zip"
    }
  }
  lambda_functions = merge(local.transaction_functions, local.wallet_functions)
}

module "lambdas" {
  for_each = local.lambda_functions

  source          = "./lambda"
  lambda_role_arn = aws_iam_role.lambda_role.arn
  table_name      = aws_dynamodb_table.dynamodb_table.name
  function_name   = each.value.function_name
  artifact_path   = each.value.artifact_path

  depends_on = [
    aws_iam_role_policy_attachment.lambda_basic_execution,
    aws_iam_role_policy.lambda_dynamodb_access
  ]
}

