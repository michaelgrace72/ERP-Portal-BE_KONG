.PHONY: help migrate-up migrate-down migrate-force migrate-version migrate-create migrate-legacy-up migrate-legacy-down migrate-legacy-fresh migrate-seed build run docker-up docker-down docker-logs

help: ## Display this help message
	@echo "Available commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

# Docker commands
docker-up: ## Start development environment with Kong (PostgreSQL, Kong, Redis, RabbitMQ)
	@echo "Starting development environment with Kong Gateway..."
	@docker compose up -d
	@echo "Waiting for services to be ready..."
	@sleep 10
	@echo "Services started successfully!"
	@echo "PostgreSQL: localhost:5432"
	@echo "Kong Proxy: http://localhost:8000 (HTTPS: 8443)"
	@echo "Kong Admin API: http://localhost:8001 (HTTPS: 8444)"
	@echo "Redis: localhost:6379"
	@echo "RabbitMQ: localhost:5672 (Management UI: http://localhost:15672)"

kong-up: ## Start Kong Gateway with all services
	@echo "Starting Kong Gateway environment..."
	@docker compose -f docker-compose.kong.yml up -d
	@echo "Waiting for Kong to be ready..."
	@sleep 10
	@echo "Running Kong setup script..."
	@./scripts/setup-kong.sh
	@echo ""
	@echo "✅ Kong is ready!"
	@echo "   Kong Proxy: http://localhost:8000"
	@echo "   Kong Admin: http://localhost:8001"

kong-down: ## Stop Kong Gateway environment
	@echo "Stopping Kong Gateway..."
	@docker compose -f docker-compose.kong.yml down

kong-logs: ## View Kong logs
	@docker compose -f docker-compose.kong.yml logs -f kong

kong-reset: ## Reset Kong configuration (remove all services/routes)
	@echo "Resetting Kong configuration..."
	@docker compose -f docker-compose.kong.yml down -v
	@docker compose -f docker-compose.kong.yml up -d
	@sleep 10
	@./scripts/setup-kong.sh

docker-down: ## Stop development environment
	@echo "Stopping development environment..."
	@docker-compose down

docker-logs: ## View logs from development services
	@docker-compose logs -f

docker-clean: ## Stop and remove all containers, volumes, and networks
	@echo "Cleaning up development environment..."
	@docker-compose down -v
	@echo "Cleanup completed!"

# Kong specific commands
kong-status: ## Check Kong Gateway status
	@echo "Checking Kong status..."
	@curl -s http://localhost:8001/status | jq

kong-consumers: ## List all Kong consumers
	@echo "Listing Kong consumers..."
	@curl -s http://localhost:8001/consumers | jq

kong-services: ## List all Kong services
	@echo "Listing Kong services..."
	@curl -s http://localhost:8001/services | jq

kong-routes: ## List all Kong routes
	@echo "Listing Kong routes..."
	@curl -s http://localhost:8001/routes | jq

# Golang-migrate commands (recommended for production)
migrate-up: ## Run all pending migrations
	@echo "Running migrations..."
	@go run cmd/migrate/main.go up

migrate-down: ## Rollback the last migration
	@echo "Rolling back last migration..."
	@go run cmd/migrate/main.go down

migrate-force: ## Force migration to a specific version (usage: make migrate-force VERSION=1)
	@echo "Forcing migration to version $(VERSION)..."
	@go run cmd/migrate/main.go force $(VERSION)

migrate-version: ## Show current migration version
	@go run cmd/migrate/main.go version

migrate-create: ## Create a new migration file (usage: make migrate-create NAME=create_users_table)
	@echo "Creating migration: $(NAME)"
	@go run cmd/migrate/main.go create $(NAME)

# Seeder commands
seed: ## Run database seeders (roles and permissions)
	@echo "Running database seeders..."
	@go run cmd/seed/main.go cmd/seed/copy_roles.go

copy-roles: ## Copy system roles to a tenant (usage: make copy-roles TENANT_ID=1)
	@if [ -z "$(TENANT_ID)" ]; then \
		echo "Error: TENANT_ID is required"; \
		echo "Usage: make copy-roles TENANT_ID=1"; \
		exit 1; \
	fi
	@echo "Copying system roles to tenant $(TENANT_ID)..."
	@go run cmd/copy-roles/main.go --tenant-id=$(TENANT_ID)

# Legacy GORM migration commands (for development)
migrate-legacy-up: ## Run GORM auto-migrations (development only)
	@echo "Running GORM migrations..."
	@go run cmd/migrate/main.go migrate

migrate-legacy-down: ## Rollback GORM migrations (development only)
	@echo "Rolling back GORM migrations..."
	@go run cmd/migrate/main.go rollback

migrate-legacy-fresh: ## Drop all tables and recreate (development only)
	@echo "Running fresh GORM migrations..."
	@go run cmd/migrate/main.go fresh

# Build and run commands
build: ## Build the application
	@echo "Building application..."
	@go build -o bin/server cmd/server/main.go

run: ## Run the application
	@echo "Running application..."
	@go run cmd/server/main.go

# Development helpers
install-migrate-cli: ## Install golang-migrate CLI tool
	@echo "Installing golang-migrate CLI..."
	@go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

test: ## Run tests
	@echo "Running tests..."
	@go test ./... -v

clean: ## Clean build artifacts
	@echo "Cleaning..."
	@rm -rf bin/
	@rm -rf tmp/
