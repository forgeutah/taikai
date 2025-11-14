.PHONY: help dev build test clean migrate-up migrate-down migrate-create docker-up docker-down install-deps

# Default target
.DEFAULT_GOAL := help

# Load environment variables
include .env
export

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

install-deps: ## Install Go dependencies
	@echo "Installing Go dependencies..."
	go get github.com/go-chi/chi/v5
	go get github.com/lib/pq
	go get github.com/jmoiron/sqlx
	go get github.com/pressly/goose/v3/cmd/goose
	go get github.com/hibiken/asynq
	go get github.com/golang-jwt/jwt/v5
	go get github.com/joho/godotenv
	go get golang.org/x/crypto/bcrypt
	go get github.com/google/uuid
	go mod tidy

install-tools: ## Install development tools
	@echo "Installing development tools..."
	go install github.com/pressly/goose/v3/cmd/goose@latest
	go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest

docker-up: ## Start Docker containers (PostgreSQL, Mailpit)
	@echo "Starting Docker containers..."
	@docker run -d --name taikai-postgres \
		-e POSTGRES_DB=taikai \
		-e POSTGRES_USER=taikai \
		-e POSTGRES_PASSWORD=taikai_dev_password \
		-p 5432:5432 \
		-v taikai-postgres-data:/var/lib/postgresql/data \
		postgres:17-alpine 2>/dev/null || echo "PostgreSQL container already running"
	@docker run -d --name taikai-mailpit \
		-p 1025:1025 \
		-p 8025:8025 \
		-e MP_SMTP_AUTH_ACCEPT_ANY=1 \
		-e MP_SMTP_AUTH_ALLOW_INSECURE=1 \
		axllent/mailpit:latest 2>/dev/null || echo "Mailpit container already running"
	@echo "Waiting for PostgreSQL to be ready..."
	@sleep 3
	@echo "Docker containers started!"
	@echo "PostgreSQL: localhost:5432"
	@echo "Mailpit Web UI: http://localhost:8025"

docker-down: ## Stop Docker containers
	@echo "Stopping Docker containers..."
	@docker stop taikai-postgres taikai-mailpit 2>/dev/null || true
	@docker rm taikai-postgres taikai-mailpit 2>/dev/null || true

docker-clean: ## Stop containers and remove volumes
	@echo "Cleaning Docker containers and volumes..."
	@docker stop taikai-postgres taikai-mailpit 2>/dev/null || true
	@docker rm taikai-postgres taikai-mailpit 2>/dev/null || true
	@docker volume rm taikai-postgres-data 2>/dev/null || true

migrate-up: ## Run database migrations
	@echo "Running database migrations..."
	goose -dir migrations postgres "$(DATABASE_URL)" up

migrate-down: ## Rollback last migration
	@echo "Rolling back last migration..."
	goose -dir migrations postgres "$(DATABASE_URL)" down

migrate-status: ## Check migration status
	@echo "Checking migration status..."
	goose -dir migrations postgres "$(DATABASE_URL)" status

migrate-create: ## Create a new migration file (usage: make migrate-create NAME=migration_name)
	@if [ -z "$(NAME)" ]; then \
		echo "Error: NAME is required. Usage: make migrate-create NAME=migration_name"; \
		exit 1; \
	fi
	@echo "Creating migration: $(NAME)"
	goose -dir migrations create $(NAME) sql

sqlc-generate: ## Generate Go code from SQL using sqlc
	@echo "Generating Go code with sqlc..."
	sqlc generate

dev: ## Run the development server
	@echo "Starting development server..."
	go run cmd/server/main.go

worker: ## Run the background worker
	@echo "Starting background worker..."
	go run cmd/worker/main.go

build: ## Build the application binaries
	@echo "Building application..."
	@mkdir -p bin
	go build -o bin/server cmd/server/main.go
	go build -o bin/worker cmd/worker/main.go
	go build -o bin/migrate cmd/migrate/main.go
	@echo "Binaries built in ./bin/"

test: ## Run tests
	@echo "Running tests..."
	go test -v -race -cover ./...

test-coverage: ## Run tests with coverage report
	@echo "Running tests with coverage..."
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

clean: ## Clean build artifacts
	@echo "Cleaning build artifacts..."
	rm -rf bin/
	rm -f coverage.out coverage.html

setup: docker-up migrate-up ## Complete setup: start Docker and run migrations
	@echo "Setup complete!"
	@echo "You can now run 'make dev' to start the development server"

seed: ## Seed the database with test data
	@echo "Seeding database..."
	go run cmd/migrate/seed.go

lint: ## Run linter
	@echo "Running linter..."
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed. Install it from https://golangci-lint.run/usage/install/"; \
	fi

fmt: ## Format Go code
	@echo "Formatting code..."
	go fmt ./...
	gofmt -s -w .

vet: ## Run go vet
	@echo "Running go vet..."
	go vet ./...

all: clean install-deps build test ## Clean, install deps, build, and test
	@echo "All tasks completed!"

.PHONY: docker-logs
docker-logs: ## Show Docker container logs
	docker-compose logs -f
