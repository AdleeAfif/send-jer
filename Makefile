.PHONY: help build run test clean docker-build docker-up docker-down

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build the application
	go build -o send-jer cmd/server/main.go

run: ## Run the application
	go run cmd/server/main.go

test: ## Run tests
	go test -v ./...

clean: ## Clean build artifacts
	rm -f send-jer send-jer.exe
	go clean

install: ## Install dependencies
	go mod download

tidy: ## Tidy go modules
	go mod tidy

docker-build: ## Build Docker image
	docker build -t send-jer:latest .

docker-up: ## Start Docker Compose services
	docker-compose up -d

docker-down: ## Stop Docker Compose services
	docker-compose down

docker-logs: ## View Docker logs
	docker-compose logs -f

lint: ## Run linter
	golangci-lint run

format: ## Format code
	go fmt ./...

dev: ## Run in development mode with auto-reload (requires air)
	air
