LAMBDA_MAIN_FILES := $(shell find app/lambda -type f -name main.go)
LAMBDA_ARTIFACT_DIR := aws/lambda
LAMBDA_BUILD_DIR := $(LAMBDA_ARTIFACT_DIR)/build

.PHONY: all build-lambdas zip-lambdas clean

SAM_TEMPLATE := template.sam.yaml
SAM_BUILD_TEMPLATE := .aws-sam/build/template.yaml
LOCAL_API_PORT := 3000
LOCAL_DOCKER_COMPOSE_FILE := docker-compose.local.yml
SAM_LOCAL_REGION := ap-southeast-1
SAM_LOCAL_ENV := AWS_SHARED_CREDENTIALS_FILE=/dev/null AWS_CONFIG_FILE=/dev/null AWS_REGION=$(SAM_LOCAL_REGION) DOCKER_HOST=$$(docker context inspect $$(docker context show) --format '{{.Endpoints.docker.Host}}')

all: zip-lambdas

build-lambdas:
	@mkdir -p "$(LAMBDA_BUILD_DIR)"
	@set -e; \
	for main in $(LAMBDA_MAIN_FILES); do \
		rel=$${main#app/}; \
		pkg=$$(dirname "$$rel"); \
		name=$$(basename "$$pkg"); \
		out_dir="$(LAMBDA_BUILD_DIR)/$$name"; \
		mkdir -p "$$out_dir"; \
		echo "Building $$name from ./app/$$pkg"; \
		(cd app && GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o "../$$out_dir/bootstrap" "./$$pkg"); \
	done

zip-lambdas: build-lambdas
	@set -e; \
	for dir in "$(LAMBDA_BUILD_DIR)"/*; do \
		[ -d "$$dir" ] || continue; \
		name=$$(basename "$$dir"); \
		echo "Zipping $$name to $(LAMBDA_ARTIFACT_DIR)/$$name.zip"; \
		(cd "$$dir" && zip -q -FS "../../$$name.zip" bootstrap); \
	done

clean:
	@rm -rf "$(LAMBDA_BUILD_DIR)"
	@rm -f "$(LAMBDA_ARTIFACT_DIR)"/*.zip

.PHONY: local-dynamodb-up local-dynamodb-init local-dynamodb-down sam-build sam-local-api sam-local-stop local-up

local-dynamodb-up:
	docker compose -f $(LOCAL_DOCKER_COMPOSE_FILE) up -d dynamodb-local

local-dynamodb-init:
	bash scripts/init_local_dynamodb.sh

local-dynamodb-down:
	docker compose -f $(LOCAL_DOCKER_COMPOSE_FILE) stop dynamodb-local

sam-build:
	$(SAM_LOCAL_ENV) sam build --template-file $(SAM_TEMPLATE) --region $(SAM_LOCAL_REGION)

sam-local-api:
	$(SAM_LOCAL_ENV) sam local start-api \
		--template $(SAM_BUILD_TEMPLATE) \
		--host 0.0.0.0 \
		--port $(LOCAL_API_PORT) \
		--region $(SAM_LOCAL_REGION) \
		--add-host host.docker.internal:host-gateway

local-up: local-dynamodb-up local-dynamodb-init sam-build
	@echo "Local dependencies are ready."
	@echo "Run 'make sam-local-api' to start API Gateway + Lambda locally on port $(LOCAL_API_PORT)."