# Go Weather Lambda Makefile

.PHONY: help build test clean deploy local-build docker-build docker-dev docker-sam

# Variables
BINARY_NAME=weather-lambda
HISTORY_BINARY_NAME=weather-history-api
BUILD_DIR=bin
HISTORY_BUILD_DIR=bin-history
DOCKER_COMPOSE=docker compose
SAM_TEMPLATE=template.yaml

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# Build targets
build: ## Build the Lambda function binary (requires Go 1.23+)
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
		-ldflags='-w -s -extldflags "-static"' \
		-a -installsuffix cgo \
		-o $(BUILD_DIR)/$(BINARY_NAME) \
		./cmd/weather-lambda

build-docker: deps-docker ## Build the Lambda function using Docker
	@echo "Building $(BINARY_NAME) with Docker..."
	@$(DOCKER_COMPOSE) exec dev sh -c "mkdir -p $(BUILD_DIR) && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -buildvcs=false -ldflags='-w -s -extldflags \"-static\"' -a -installsuffix cgo -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/weather-lambda"

build-history: ## Build the Weather History API Lambda function binary (requires Go 1.23+)
	@echo "Building $(HISTORY_BINARY_NAME)..."
	@mkdir -p $(HISTORY_BUILD_DIR)
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
		-ldflags='-w -s -extldflags "-static"' \
		-a -installsuffix cgo \
		-o $(HISTORY_BUILD_DIR)/bootstrap \
		./cmd/weather-history-api

build-history-docker: deps-docker ## Build the Weather History API Lambda function using Docker
	@echo "Building $(HISTORY_BINARY_NAME) with Docker..."
	@$(DOCKER_COMPOSE) exec dev sh -c "mkdir -p $(HISTORY_BUILD_DIR) && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -buildvcs=false -ldflags='-w -s -extldflags \"-static\"' -a -installsuffix cgo -o $(HISTORY_BUILD_DIR)/bootstrap ./cmd/weather-history-api"

build-all: build-docker build-history-docker ## Build all Lambda functions using Docker

local-build: ## Build for local development
	@echo "Building $(BINARY_NAME) for local development..."
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/weather-lambda

# Docker targets
docker-build: ## Build using Docker
	@echo "Building with Docker..."
	@docker build -t weather-lambda .

docker-dev: ## Start development environment
	@echo "Starting development environment..."
	@$(DOCKER_COMPOSE) up -d dev
	@echo "Development environment is ready!"
	@echo "Enter development container: docker compose exec dev bash"

docker-sam: ## Start SAM CLI environment
	@echo "Starting SAM CLI environment..."
	@$(DOCKER_COMPOSE) up -d sam
	@echo "SAM environment is ready!"
	@echo "Enter SAM container: docker compose exec sam bash"

docker-stop: ## Stop all Docker services
	@echo "Stopping Docker services..."
	@$(DOCKER_COMPOSE) down

# Test targets
test: ## Run tests
	@echo "Running tests..."
	@go test -v ./...

test-aws: ## Run tests against AWS resources
	@echo "Running tests against AWS..."
	@echo "Note: Requires valid AWS credentials and deployed resources"
	@go test -v -tags=integration ./tests/...

# Lint and format
lint: ## Run linter
	@echo "Running linter..."
	@golangci-lint run

fmt: ## Format code
	@echo "Formatting code..."
	@go fmt ./...

# SAM targets
sam-build: build-all ## Build with SAM
	@echo "Building with SAM..."
	@$(DOCKER_COMPOSE) exec dev sam build

sam-invoke-aws: ## Invoke deployed Lambda function in AWS
	@echo "Invoking deployed Lambda function..."
	@./scripts/validate-env.sh
	@$(DOCKER_COMPOSE) exec dev sh -c ". ./.env && FUNC_ARN=$$(/usr/bin/aws cloudformation describe-stacks --stack-name \$$STACK_NAME --region \$$AWS_REGION --query 'Stacks[0].Outputs[?OutputKey==\`WeatherLambdaFunction\`].OutputValue' --output text) && echo \"Function ARN: \$$FUNC_ARN\" && /usr/bin/aws lambda invoke --region \$$AWS_REGION --function-name \"\$$FUNC_ARN\" response.json && cat response.json"

test-history-api: ## Test the Weather History API endpoint (requires API key)
	@echo "Testing Weather History API with API key..."
	@./scripts/validate-env.sh
	@$(DOCKER_COMPOSE) exec dev sh -c ". ./.env && echo \"Using API Key: \$$API_KEY_VALUE\" && curl -s -H \"X-API-Key: \$$API_KEY_VALUE\" \"\$$API_GATEWAY_URL/weather/history?period=6h\""

test-history-api-no-auth: ## Test API without authentication (should fail with 403)
	@echo "Testing Weather History API without API key (expecting 403)..."
	@./scripts/validate-env.sh
	@$(DOCKER_COMPOSE) exec dev sh -c ". ./.env && curl -s -w \"HTTP Status: %{http_code}\n\" \"\$$API_GATEWAY_URL/weather/history?period=6h\""

test-rate-limiting: ## Test API rate limiting (should show throttling after 10 requests/sec)
	@echo "Testing rate limiting (10 requests/sec limit, 50 burst)..."
	@./scripts/validate-env.sh
	@$(DOCKER_COMPOSE) exec dev sh -c ". ./.env && for i in \$$(seq 1 15); do echo \"Request \$$i:\"; curl -s -w \"HTTP Status: %{http_code}\n\" -H \"X-API-Key: \$$API_KEY_VALUE\" \"\$$API_GATEWAY_URL/weather/history?period=6h\" | head -1; done"

test-aws-cli: ## Test AWS CLI in container
	@echo "Testing AWS CLI..."
	@./scripts/validate-env.sh
	@$(DOCKER_COMPOSE) exec dev sh -c ". ./.env && /usr/bin/aws --version && /usr/bin/aws sts get-caller-identity --region \$$AWS_REGION"

test-lambda-direct: ## Test Lambda function using known function name
	@echo "Testing Lambda function directly..."
	@./scripts/validate-env.sh
	@$(DOCKER_COMPOSE) exec dev sh -c ". ./.env && FUNC_NAME=$$(/usr/bin/aws cloudformation describe-stacks --stack-name \$$STACK_NAME --region \$$AWS_REGION --query 'Stacks[0].Outputs[?OutputKey==\`WeatherLambdaFunction\`].OutputValue' --output text) && /usr/bin/aws lambda invoke --region \$$AWS_REGION --function-name \"\$$FUNC_NAME\" response.json && cat response.json"

get-api-key: ## Get the API key for Weather History API
	@echo "Retrieving API key from CloudFormation stack..."
	@./scripts/validate-env.sh
	@$(DOCKER_COMPOSE) exec dev sh -c ". ./.env && echo \"API Key ID: \$$API_KEY_ID\" && echo \"API Key Value: \$$API_KEY_VALUE\""

monitor-api-usage: ## Monitor API Gateway usage statistics
	@echo "Monitoring API usage..."
	@./scripts/validate-env.sh
	@$(DOCKER_COMPOSE) exec dev sh -c ". ./.env && /usr/bin/aws apigateway get-usage --usage-plan-id \$$USAGE_PLAN_ID --start-date 2025-08-19 --end-date 2025-08-26 --region \$$AWS_REGION"

sam-deploy: sam-build ## Deploy to AWS (guided)
	@echo "Deploying to AWS with guided setup..."
	@$(DOCKER_COMPOSE) exec dev sam deploy --guided --region $${AWS_REGION:-ap-northeast-1}

sam-deploy-dev: sam-build ## Deploy to development (requires WEATHER_API_KEY env var)
	@echo "Deploying to development environment..."
	@./scripts/validate-env.sh
	@$(DOCKER_COMPOSE) exec dev sh -c ". ./.env && sam deploy \
		--region \$$AWS_REGION \
		--resolve-s3 \
		--stack-name weather-lambda-dev \
		--parameter-overrides WeatherAPIKey=\"\$$WEATHER_API_KEY\" CityName=\"\$$CITY_NAME\" Environment=dev \
		--capabilities CAPABILITY_IAM \
		--no-confirm-changeset"

sam-deploy-prod: sam-build ## Deploy to production (requires WEATHER_API_KEY env var)
	@echo "Deploying to production..."
	@./scripts/validate-env.sh
	@$(DOCKER_COMPOSE) exec dev sam deploy \
		--region $${AWS_REGION:-ap-northeast-1} \
		--resolve-s3 \
		--stack-name weather-lambda-prod \
		--parameter-overrides WeatherAPIKey="$$WEATHER_API_KEY" CityName="$${CITY_NAME:-Tokyo}" Environment=prod \
		--capabilities CAPABILITY_IAM \
		--no-confirm-changeset

# Utility targets
clean: ## Clean build artifacts
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)
	@rm -rf $(HISTORY_BUILD_DIR)
	@rm -rf .aws-sam

deps: ## Download and tidy dependencies
	@echo "Tidying dependencies..."
	@go mod tidy
	@echo "Downloading dependencies..."
	@go mod download
	@go mod verify

deps-update: ## Update dependencies
	@echo "Updating dependencies..."
	@go get -u ./...
	@go mod tidy

deps-docker: ## Tidy dependencies using Docker dev environment
	@echo "Tidying dependencies in Docker..."
	@$(DOCKER_COMPOSE) exec dev go mod tidy

# Environment setup
setup-env: ## Copy .env.example to .env
	@if [ ! -f .env ]; then \
		cp .env.example .env; \
		echo "Created .env file. Please update with your values."; \
	else \
		echo ".env file already exists."; \
	fi

# AWS helpers
aws-verify: ## Verify AWS credentials and connectivity
	@echo "Verifying AWS credentials..."
	@aws sts get-caller-identity
	@echo "AWS connectivity verified!"

# Complete development setup
dev-setup: setup-env docker-dev deps-docker ## Complete development environment setup
	@echo "Development environment setup complete!"
	@echo "Deploy to AWS with 'make sam-deploy' then test with 'make sam-invoke-aws'"

# Production deployment checklist
deploy-checklist: ## Show deployment checklist
	@echo "Pre-deployment checklist:"
	@echo "1. Set WEATHER_API_KEY environment variable"
	@echo "2. Configure AWS credentials"
	@echo "3. Review template.yaml parameters"
	@echo "4. Run 'make test' to ensure tests pass"
	@echo "5. Run 'make sam-deploy' for guided deployment"