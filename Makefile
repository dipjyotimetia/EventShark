.PHONY: test test-unit test-integration test-coverage build clean code-gen schema-gen lint fmt vet security-check help

# Default target
.DEFAULT_GOAL := help

# Test targets
test: test-unit test-integration ## Run all tests

test-unit: ## Run unit tests
	@echo "Running unit tests..."
	go test ./pkg/... -v -race -coverprofile=coverage.out

test-integration: ## Run integration tests
	@echo "Running integration tests..."
	go test ./tests/integration/... -v --tags=integration -count=1 --json > test-report.json

test-coverage: test-unit ## Run tests with coverage report
	@echo "Generating coverage report..."
	go tool cover -html=coverage.out -o coverage.html
	go tool cover -func=coverage.out

# Build targets
build: ## Build and start services with Docker Compose
	@echo "Building and starting services..."
	docker compose up --build -d

build-local: ## Build the application locally
	@echo "Building application locally..."
	go build -o bin/event-shark ./cmd

# Development targets
run: ## Run the application locally
	@echo "Running application locally..."
	go run ./cmd

dev: ## Start development environment
	@echo "Starting development environment..."
	docker compose up kafka -d
	@echo "Waiting for Kafka to be ready..."
	sleep 10
	$(MAKE) run

# Clean targets
clean: ## Clean up Docker containers and volumes
	@echo "Cleaning up..."
	docker compose down --rmi all --volumes --remove-orphans
	rm -f coverage.out coverage.html test-report.json
	rm -rf bin/

# Code generation targets
code-gen: ## Generate Go code from Avro schemas
	@echo "Generating Go code..."
	go generate ./...

schema-gen: ## Generate JSON schemas from Avro schemas
	@echo "Generating JSON schemas..."
	go run script/avsc2json/main.go schema/avro/expense.avsc > docker/schema/expense.json
	go run script/avsc2json/main.go schema/avro/payment.avsc > docker/schema/payment.json
	go run script/avsc2json/main.go schema/avro/transaction.avsc > docker/schema/transaction.json

# Quality checks
lint: ## Run linting
	@echo "Running linters..."
	@which golangci-lint > /dev/null || (echo "golangci-lint not found. Installing..." && go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
	golangci-lint run --timeout 5m

fmt: ## Format code
	@echo "Formatting code..."
	go fmt ./...
	@which goimports > /dev/null || (echo "goimports not found. Installing..." && go install golang.org/x/tools/cmd/goimports@latest)
	goimports -w .

vet: ## Run go vet
	@echo "Running go vet..."
	go vet ./...

security-check: ## Run security checks
	@echo "Running security checks..."
	@which gosec > /dev/null || (echo "gosec not found. Installing..." && go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest)
	gosec ./...

# Dependencies
deps: ## Download dependencies
	@echo "Downloading dependencies..."
	go mod download
	go mod tidy

deps-update: ## Update dependencies
	@echo "Updating dependencies..."
	go get -u ./...
	go mod tidy

# Performance tests
perf-test: ## Run performance tests
	@echo "Running performance tests..."
	cd tests/performance && npm test

# Docker targets
docker-build: ## Build Docker image
	@echo "Building Docker image..."
	docker build -t event-shark:latest .

docker-run: ## Run Docker container
	@echo "Running Docker container..."
	docker run -p 8083:8083 event-shark:latest

# Health checks
health-check: ## Check service health
	@echo "Checking service health..."
	curl -f http://localhost:8083/health || exit 1
	curl -f http://localhost:8083/health/ready || exit 1
	curl -f http://localhost:8083/health/live || exit 1

# Utility targets
logs: ## Show Docker compose logs
	docker compose logs -f

restart: ## Restart services
	docker compose restart

stop: ## Stop services
	docker compose stop

# Quality gate - runs all quality checks
quality-gate: fmt lint vet security-check test-unit ## Run all quality checks

# CI/CD target
ci: deps quality-gate test-integration ## Run CI pipeline

# Help target
help: ## Show this help message
	@echo "EventShark Makefile"
	@echo "==================="
	@echo ""
	@echo "Available targets:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)
	@echo ""
	@echo "Examples:"
	@echo "  make build          # Build and start with Docker Compose"
	@echo "  make test           # Run all tests"
	@echo "  make dev            # Start development environment"
	@echo "  make quality-gate   # Run all quality checks"
	@echo "  make ci             # Run CI pipeline"