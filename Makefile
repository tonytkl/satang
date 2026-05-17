LAMBDA_MAIN_FILES := $(shell find app/lambda -type f -name main.go)
LAMBDA_ARTIFACT_DIR := aws/lambda
LAMBDA_BUILD_DIR := $(LAMBDA_ARTIFACT_DIR)/build

.PHONY: all build-lambdas zip-lambdas clean

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