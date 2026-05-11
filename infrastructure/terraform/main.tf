terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

provider "aws" {
  region = var.region

  skip_credentials_validation = lower(var.environment) == "local"
  skip_metadata_api_check     = lower(var.environment) == "local"
  skip_requesting_account_id  = lower(var.environment) == "local"

  endpoints {
    dynamodb   = lower(var.environment) == "local" ? lookup(var.local_endpoints, "dynamodb", null) : null
    lambda     = lower(var.environment) == "local" ? lookup(var.local_endpoints, "lambda", null) : null
    apigateway = lower(var.environment) == "local" ? lookup(var.local_endpoints, "apigateway", null) : null
  }

  access_key = lower(var.environment) == "local" ? "test" : var.access_key
  secret_key = lower(var.environment) == "local" ? "test" : var.secret_key
}