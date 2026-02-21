.PHONY: help proto build test clean docker-build docker-up docker-down install-tools

# Variables
PROTO_DIR := proto
SERVICES := auth integration workflow consent notification
GO_VERSION := 1.21

help: ## Show this help message
	@echo 'NeighbourHood Microservices - Available Commands:'
	@echo ''
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-25s %s\n", $$1, $$2}' $(MAKEFILE_LIST)
	@echo ''

install-tools: ## Install required development tools
	@echo "Installing Protocol Buffer compiler and plugins..."
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	@echo "Installing other tools..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "Tools installed successfully!"

proto: ## Generate Go code from proto files
	@echo "Generating protobuf code..."
	@mkdir -p proto/gen/go
	protoc --go_out=proto/gen/go --go_opt=module=neighbourhood/proto/gen/go \
		--go-grpc_out=proto/gen/go --go-grpc_opt=module=neighbourhood/proto/gen/go \
		$(PROTO_DIR)/*.proto
	@echo "Protobuf code generated successfully!"

proto-clean: ## Clean generated proto files
	@echo "Cleaning generated proto files..."
	rm -rf proto/gen
	@echo "Generated proto files cleaned!"

deps: ## Download Go module dependencies
	@echo "Downloading dependencies..."
	go mod download
	go mod tidy
	@echo "Dependencies downloaded!"

build: ## Build legacy monolith
	@echo "Building NeighbourHood monolith..."
	@cd cmdapi && go build -o ../bin/neighbourhood main.go
	@echo "Build complete: bin/neighbourhood"

build-services: proto ## Build all microservices
	@echo "Building all microservices..."
	@mkdir -p bin
	@echo "Building auth service..."
	@go build -o bin/auth-service ./services/auth/cmd/server
	@echo "Note: Other services (integration, workflow, consent, notification) structure not yet created"
	@echo "Build complete: bin/auth-service"

build-auth: proto ## Build auth microservice
	@echo "Building auth service..."
	@mkdir -p bin
	@go build -o bin/auth-service ./services/auth/cmd/server
	@echo "Build complete: bin/auth-service"

build-integration: proto ## Build integration microservice
	@echo "Building integration service..."
	@mkdir -p bin
	@go build -o bin/integration-service ./services/integration/cmd/server
	@echo "Build complete: bin/integration-service"

build-all: proto ## Build all microservices
	@echo "Building all microservices..."
	@mkdir -p bin
	@go build -o bin/auth-service ./services/auth/cmd/server
	@go build -o bin/integration-service ./services/integration/cmd/server
	@echo "Build complete!"

run-auth: build-auth ## Run auth microservice
	@echo "Starting auth service..."
	@./bin/auth-service

run-integration: build-integration ## Run integration microservice
	@echo "Starting integration service..."
	@./bin/integration-service

run: ## Run legacy monolith
	@echo "Starting NeighbourHood monolith..."
	@cd cmdapi && go run main.go

test: ## Run all tests
	@echo "Running tests..."
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Tests completed! Coverage report: coverage.html"

test-unit: ## Run unit tests only
	@echo "Running unit tests..."
	go test -v -short ./...

lint: ## Run linter
	@echo "Running linter..."
	golangci-lint run ./...
	@echo "Linting completed!"

format: ## Format code
	@echo "Formatting code..."
	go fmt ./...
	@echo "Code formatted!"

clean: ## Clean build artifacts
	@echo "Cleaning..."
	@rm -rf bin/
	@rm -f coverage.out coverage.html
	@echo "Clean complete"

docker-build: ## Build Docker image for monolith
	@echo "Building Docker image..."
	docker build -t neighbourhood:latest .
	@echo "Docker image built successfully!"

docker-up: ## Start Docker containers
	@echo "Starting Docker containers..."
	@docker-compose up -d
	@echo "Containers started"

docker-down: ## Stop Docker containers
	@echo "Stopping Docker containers..."
	@docker-compose down
	@echo "Containers stopped"

docker-logs: ## View Docker logs
	@docker-compose logs -f

migrate-up: ## Run database migrations
	@echo "Running migrations..."
	@psql -h localhost -U postgres -d neighbourhood -f internal/models/schema.sql
	@echo "Migrations complete"

migrate-down: ## Rollback migrations (manual process)
	@echo "Please manually rollback migrations"

install-deps: ## Install Go dependencies
	@echo "Installing dependencies..."
	@go mod download
	@go mod tidy
	@echo "Dependencies installed"

install-tools: ## Install development tools
	@echo "Installing development tools..."
	@go install github.com/cosmtrek/air@latest
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go install golang.org/x/tools/cmd/goimports@latest
	@echo "Tools installed"

setup: install-deps ## Initial project setup
	@echo "Setting up project..."
	@cp -n .env.example .env || true
	@mkdir -p bin
	@echo "Setup complete. Please configure your .env file"

check: lint test ## Run linter and tests

all: clean build test ## Clean, build, and test
