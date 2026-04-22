# GitHub Copilot Custom Instructions

## Project Overview
This project is the backend service for an expense tracking application.

The system is designed to be:
- Scalable
- Serverless
- Cost-efficient
- Cloud-native (AWS-first)

## Architecture Guidelines

### Cloud Provider
- Use AWS serverless architecture
- Prefer managed services over self-hosted solutions
- Free-tier plan eligible is preferable for cost efficiency, but not mandatory

### Compute
- All compute logic must be implemented using AWS Lambda
- Lambdas should be:
  - Stateless
  - Idempotent where possible
  - Optimized for cold start performance

### Infrastructure as Code
- Infrastructure must be defined using Terraform
- Follow best practices:
  - Modular design
  - Reusable components
  - Environment separation (dev, staging, prod)

### Application Deployment
- Lambda functions are written in Go (Golang)
- Use AWS CDK for Go when defining Lambda-related constructs if needed
- Ensure build and packaging steps are compatible with AWS Lambda runtime

### Data Storage
- Use Amazon DynamoDB as the primary database
- Follow DynamoDB best practices:
  - Single-table design preferred
  - Optimize for access patterns
  - Use GSIs where necessary
  - Avoid joins; design for denormalized data

## Project Structure

The repository must follow this structure:

/app
  └── (Golang application code for Lambda functions)

/infrastructure
  └── (Terraform code for AWS resources)

### /app Guidelines
- Written in Go
- Organized by domain or feature (e.g., transactions, users)
- Follow clean architecture principles:
  - handler (Lambda entrypoint)
  - service (business logic)
  - repository (data access)
- Use interfaces for testability

### /infrastructure Guidelines
- Terraform code only
- Structure:
  - modules/ for reusable components
  - env/ or environments/ for environment-specific configs
- Manage:
  - Lambda functions
  - API Gateway
  - DynamoDB tables
  - IAM roles and policies

## API Design
- Use API Gateway in front of Lambda
- Follow RESTful conventions
- JSON-based request/response

## Security
- Apply least privilege principle for IAM roles
- Never hardcode secrets
- Use AWS Secrets Manager or Parameter Store if needed

## Observability
- Enable:
  - CloudWatch Logs
  - Structured logging (JSON preferred)
- Include meaningful log messages for debugging

## Coding Standards

### Golang
- Follow idiomatic Go practices
- Keep functions small and focused
- Handle errors explicitly
- Avoid global state

### General
- Prefer simplicity over over-engineering
- Write clear, maintainable code
- Add comments only when necessary

## Testing
- Write unit tests for business logic
- Mock external dependencies (e.g., DynamoDB)

## Copilot Behavior Instructions
When generating code:
- Always align with serverless architecture
- Prefer DynamoDB-compatible patterns
- Generate Terraform for infrastructure unless specified otherwise
- Keep Go code modular and production-ready
- Avoid introducing unnecessary frameworks or heavy dependencies
- Update READEME.md with any new features or architectural changes