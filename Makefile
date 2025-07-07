GOPATH:=$(shell go env GOPATH)
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
BINARY_NAME=greenlync-api-gateway
BINARY_UNIX=$(BINARY_NAME)_unix

# Build info
VERSION := $(shell git describe --tags --always --dirty)
COMMIT := $(shell git rev-parse --short HEAD)
DATE := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS := -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)

# Colors for output
GREEN=\033[0;32m
YELLOW=\033[1;33m
RED=\033[0;31m
NC=\033[0m # No Color

.DEFAULT_GOAL := help

## Development Environment Setup
.PHONY: help
help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  $(GREEN)%-20s$(NC) %s\n", $$1, $$2}' $(MAKEFILE_LIST)

.PHONY: setup
setup: ## Install development dependencies
	@echo "$(GREEN)Installing development dependencies...$(NC)"
	@go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	@go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	@go install github.com/swaggo/swag/cmd/swag@latest
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go install github.com/securecodewarrior/sast-scan-runner@latest
	@go install honnef.co/go/tools/cmd/staticcheck@latest
	@go install github.com/fzipp/gocyclo/cmd/gocyclo@latest
	@echo "$(GREEN)Development dependencies installed!$(NC)"

.PHONY: init
init: setup ## Initialize project dependencies
	@echo "$(GREEN)Initializing project...$(NC)"
	@go get -u google.golang.org/protobuf@v1.26.0 
	@go mod download
	@go mod tidy
	@echo "$(GREEN)Project initialized!$(NC)"

## Code Generation & Protobuf
.PHONY: proto
proto: ## Generate protobuf code
	@echo "$(GREEN)Generating protobuf code...$(NC)"
	@protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		proto/greenlync/*.proto
	@echo "$(GREEN)Protobuf code generated!$(NC)"

.PHONY: swagger
swagger: ## Generate Swagger documentation
	@echo "$(GREEN)Generating Swagger documentation...$(NC)"
	@swag fmt -g cmd/main.go
	@swag init -g cmd/main.go
	@echo "$(GREEN)Swagger documentation generated!$(NC)"

## Code Quality & Testing
.PHONY: format
format: ## Format Go code
	@echo "$(GREEN)Formatting code...$(NC)"
	@go fmt ./...
	@goimports -w .
	@echo "$(GREEN)Code formatted!$(NC)"

.PHONY: lint
lint: ## Run linters
	@echo "$(GREEN)Running linters...$(NC)"
	@golangci-lint run ./...
	@staticcheck ./...
	@echo "$(GREEN)Linting completed!$(NC)"

.PHONY: security
security: ## Run security checks
	@echo "$(GREEN)Running security checks...$(NC)"
	@gosec ./...
	@echo "$(GREEN)Security checks completed!$(NC)"

.PHONY: complexity
complexity: ## Check code complexity
	@echo "$(GREEN)Checking code complexity...$(NC)"
	@gocyclo -over 15 .
	@echo "$(GREEN)Complexity check completed!$(NC)"

.PHONY: test
test: ## Run tests
	@echo "$(GREEN)Running tests...$(NC)"
	@go test -v ./... -cover -race
	@echo "$(GREEN)Tests completed!$(NC)"

.PHONY: test-coverage
test-coverage: ## Run tests with coverage report
	@echo "$(GREEN)Running tests with coverage...$(NC)"
	@go test -v ./... -coverprofile=coverage.out -covermode=atomic
	@go tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)Coverage report generated: coverage.html$(NC)"

.PHONY: benchmark
benchmark: ## Run benchmarks
	@echo "$(GREEN)Running benchmarks...$(NC)"
	@go test -bench=. -benchmem ./...
	@echo "$(GREEN)Benchmarks completed!$(NC)"

.PHONY: check
check: format lint security test ## Run all code quality checks

## Dependency Management
.PHONY: update
update: ## Update dependencies
	@echo "$(GREEN)Updating dependencies...$(NC)"
	@go get -u ./...
	@go mod tidy
	@echo "$(GREEN)Dependencies updated!$(NC)"

.PHONY: tidy
tidy: ## Clean up dependencies
	@echo "$(GREEN)Tidying dependencies...$(NC)"
	@go mod tidy
	@go mod verify
	@echo "$(GREEN)Dependencies tidied!$(NC)"

.PHONY: vendor
vendor: ## Vendor dependencies
	@echo "$(GREEN)Vendoring dependencies...$(NC)"
	@go mod vendor
	@echo "$(GREEN)Dependencies vendored!$(NC)"

## Build & Run
.PHONY: build
build: ## Build the application
	@echo "$(GREEN)Building application...$(NC)"
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
		-a -installsuffix cgo \
		-ldflags '-w -s $(LDFLAGS)' \
		-o ./cmd/$(BINARY_NAME) cmd/*.go
	@echo "$(GREEN)Application built: ./cmd/$(BINARY_NAME)$(NC)"

.PHONY: build-race
build-race: ## Build with race detection
	@echo "$(GREEN)Building with race detection...$(NC)"
	@CGO_ENABLED=1 go build -race -o ./cmd/$(BINARY_NAME) cmd/*.go
	@echo "$(GREEN)Race detection build completed!$(NC)"

.PHONY: run
run: ## Run the application
	@echo "$(GREEN)Starting application...$(NC)"
	@go run cmd/main.go

.PHONY: run-race
run-race: ## Run with race detection
	@echo "$(GREEN)Starting application with race detection...$(NC)"
	@CGO_ENABLED=1 go run -race cmd/main.go

.PHONY: debug
debug: ## Run with debug mode
	@echo "$(GREEN)Starting application in debug mode...$(NC)"
	@DEBUG=true go run cmd/main.go

## Database
.PHONY: seed
seed: ## Seed database
	@echo "$(GREEN)Seeding database...$(NC)"
	@go run data/*.go -user greenlync -password
	@echo "$(GREEN)Database seeded!$(NC)"

.PHONY: migrate
migrate: ## Run database migrations
	@echo "$(GREEN)Running database migrations...$(NC)"
	@go run cmd/main.go -migrate
	@echo "$(GREEN)Database migrations completed!$(NC)"

## Docker Operations
.PHONY: docker-build
docker-build: ## Build Docker image
	@echo "$(GREEN)Building Docker image...$(NC)"
	@docker build -t $(BINARY_NAME):dev .
	@echo "$(GREEN)Docker image built: $(BINARY_NAME):dev$(NC)"

.PHONY: docker-build-prod
docker-build-prod: ## Build production Docker image
	@echo "$(GREEN)Building production Docker image...$(NC)"
	@docker build -t $(BINARY_NAME):$(VERSION) -f Dockerfile.prod .
	@echo "$(GREEN)Production Docker image built: $(BINARY_NAME):$(VERSION)$(NC)"

.PHONY: docker-run
docker-run: ## Run Docker container
	@echo "$(GREEN)Running Docker container...$(NC)"
	@docker run -p 8888:8888 --env-file .env $(BINARY_NAME):dev

.PHONY: docker-clean
docker-clean: ## Clean Docker images and containers
	@echo "$(GREEN)Cleaning Docker images and containers...$(NC)"
	@docker system prune -f
	@docker rmi $(BINARY_NAME):dev 2>/dev/null || true
	@echo "$(GREEN)Docker cleanup completed!$(NC)"

## Docker Compose Operations
.PHONY: up
up: ## Start all services with monitoring
	@echo "$(GREEN)Starting all services...$(NC)"
	@docker-compose up -d
	@echo "$(GREEN)All services started!$(NC)"
	@echo "$(YELLOW)Services available at:$(NC)"
	@echo "  API Gateway: http://localhost:8888"
	@echo "  Prometheus: http://localhost:9090"
	@echo "  Grafana: http://localhost:3000 (admin/admin)"
	@echo "  Jaeger: http://localhost:16686"

.PHONY: down
down: ## Stop all services
	@echo "$(GREEN)Stopping all services...$(NC)"
	@docker-compose down
	@echo "$(GREEN)All services stopped!$(NC)"

.PHONY: logs
logs: ## Show logs for all services
	@docker-compose logs -f

.PHONY: logs-app
logs-app: ## Show application logs
	@docker-compose logs -f greenlync-api-gateway

.PHONY: stack-up
stack-up: ## Start dependency stack only
	@echo "$(GREEN)Starting dependency stack...$(NC)"
	@docker-compose -f stack.yaml up -d
	@echo "$(GREEN)Dependency stack started!$(NC)"

.PHONY: stack-down
stack-down: ## Stop dependency stack
	@echo "$(GREEN)Stopping dependency stack...$(NC)"
	@docker-compose -f stack.yaml down
	@echo "$(GREEN)Dependency stack stopped!$(NC)"

.PHONY: monitoring-up
monitoring-up: ## Start monitoring stack only
	@echo "$(GREEN)Starting monitoring stack...$(NC)"
	@docker-compose -f monitoring.yaml up -d
	@echo "$(GREEN)Monitoring stack started!$(NC)"
	@echo "$(YELLOW)Monitoring services:$(NC)"
	@echo "  Prometheus: http://localhost:9090"
	@echo "  Grafana: http://localhost:3000"
	@echo "  Jaeger: http://localhost:16686"

.PHONY: monitoring-down
monitoring-down: ## Stop monitoring stack
	@echo "$(GREEN)Stopping monitoring stack...$(NC)"
	@docker-compose -f monitoring.yaml down
	@echo "$(GREEN)Monitoring stack stopped!$(NC)"

## Process Management
.PHONY: kill
kill: ## Kill application running on port 8888
	@echo "$(GREEN)Killing application on port 8888...$(NC)"
	@kill -9 $$(lsof -t -i tcp:8888) 2>/dev/null || echo "No process running on port 8888"

.PHONY: ps
ps: ## Show running processes
	@echo "$(GREEN)Showing running processes...$(NC)"
	@lsof -i :8888 || echo "No processes running on port 8888"

## Maintenance
.PHONY: clean
clean: ## Clean build artifacts
	@echo "$(GREEN)Cleaning build artifacts...$(NC)"
	@go clean
	@rm -f $(BINARY_NAME)
	@rm -f ./cmd/$(BINARY_NAME)
	@rm -f coverage.out coverage.html
	@echo "$(GREEN)Cleanup completed!$(NC)"

.PHONY: reset
reset: clean ## Reset project state
	@echo "$(GREEN)Resetting project state...$(NC)"
	@docker-compose down -v --remove-orphans
	@docker system prune -f
	@rm -rf vendor/
	@echo "$(GREEN)Project reset completed!$(NC)"

## Release Management
.PHONY: release-dev
release-dev: check build docker-build ## Prepare development release
	@echo "$(GREEN)Development release prepared!$(NC)"

.PHONY: release-staging
release-staging: check build docker-build-prod ## Prepare staging release
	@echo "$(GREEN)Staging release prepared!$(NC)"

.PHONY: release-prod
release-prod: check build docker-build-prod ## Prepare production release
	@echo "$(GREEN)Production release prepared!$(NC)"
	@echo "$(YELLOW)Remember to tag the release: git tag -a v$(VERSION) -m 'Release $(VERSION)'$(NC)"

## Utilities
.PHONY: version
version: ## Show version information
	@echo "Version: $(VERSION)"
	@echo "Commit: $(COMMIT)"
	@echo "Date: $(DATE)"

.PHONY: env-check
env-check: ## Check environment variables
	@echo "$(GREEN)Checking environment configuration...$(NC)"
	@go run cmd/main.go -config-check || true

.PHONY: health
health: ## Check application health
	@echo "$(GREEN)Checking application health...$(NC)"
	@curl -f http://localhost:8888/api/v1/system/monitor/health || echo "$(RED)Application not responding$(NC)"

## Legacy targets (for backward compatibility)
.PHONY: build-app
build-app: build ## Legacy: Build application (use 'make build' instead)

.PHONY: docker-compose-up
docker-compose-up: up ## Legacy: Start services (use 'make up' instead)

.PHONY: docker-compose-down
docker-compose-down: down ## Legacy: Stop services (use 'make down' instead)

.PHONY: docker-stack-up
docker-stack-up: stack-up ## Legacy: Start stack (use 'make stack-up' instead)

.PHONY: docker-stack-down
docker-stack-down: stack-down ## Legacy: Stop stack (use 'make stack-down' instead)

.PHONY: race
race: run-race ## Legacy: Run with race detection (use 'make run-race' instead)

.PHONY: web
web: ## Legacy: Web assets placeholder
	@echo "$(YELLOW)Web assets build - configure your frontend build process$(NC)"

.PHONY: docs
docs: swagger ## Legacy: Documentation build
