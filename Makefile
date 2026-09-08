.PHONY: run build test clean docker-build docker-run help

# Default target
help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

run: ## Run the application in development mode
	go run cmd/server/main.go

build: ## Build the application
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bin/server ./cmd/server

test: ## Run tests
	go test -v ./...

clean: ## Clean build files
	rm -rf bin/

install-deps: ## Install/update dependencies
	go mod tidy
	go mod download

docker-build: ## Build Docker image
	docker build -t ecommerce-api:latest .

docker-run: ## Run Docker container
	docker-compose up --build

docker-stop: ## Stop Docker containers
	docker-compose down

lint: ## Run golangci-lint
	golangci-lint run

format: ## Format Go code
	go fmt ./...

# Production targets
prod-run: ## Run in production mode (requires SSL certs)
	PRODUCTION=true go run cmd/server/main.go

# Development targets
dev-setup: ## Setup development environment
	cp .env.example .env
	@echo "Please edit .env file and add your STRIPE_SECRET_KEY"

# Security check
security-check: ## Check for security vulnerabilities
	go list -json -m all | nancy sleuth