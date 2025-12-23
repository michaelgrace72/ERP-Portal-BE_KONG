.PHONY: help build up down logs ps infra blue green deploy-blue deploy-green health-check-blue health-check-green clean

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build Portal service image
	docker build -t localhost:8001/portal-be:latest .

infra: ## Start infrastructure services (PostgreSQL, Kong, Redis, RabbitMQ, PgAdmin)
	docker compose -f docker-compose.local.yml up -d portal-db kong-database kong-migrations kong redis rabbitmq pgadmin
	@echo "✅ Infrastructure started"
	@echo "- Portal DB: localhost:5432 (Internal)"
	@echo "- PgAdmin: localhost:5050 (admin@admin.com / admin)"
	@echo "- Kong Proxy: localhost:3600"
	@echo "- Kong Admin: localhost:3602"
	@echo "- Redis: localhost:3606"
	@echo "- RabbitMQ: localhost:3608 (UI)"

blue: ## Start/deploy blue instance
	docker compose up -d portal-be-blue
	@echo "✅ Blue instance started on port 3502"

green: ## Start/deploy green instance
	docker compose up -d portal-be-green
	@echo "✅ Green instance started on port 3503"

deploy-blue: build ## Build and deploy to blue
	docker compose up -d portal-be-blue
	@echo "✅ Blue deployment complete"

deploy-green: build ## Build and deploy to green
	docker compose up -d portal-be-green
	@echo "✅ Green deployment complete"

up: ## Start all services (infrastructure + blue)
	docker compose up -d

down: ## Stop all services
	docker compose down

logs: ## Show logs for all services
	docker compose logs -f

logs-blue: ## Show logs for blue instance
	docker compose logs -f portal-be-blue

logs-green: ## Show logs for green instance
	docker compose logs -f portal-be-green

ps: ## Show running containers
	docker compose ps

health-check-blue: ## Check blue instance health
	@echo "Checking Blue instance..."
	@curl -s http://localhost:3502/health || echo "❌ Blue: DOWN"

health-check-green: ## Check green instance health
	@echo "Checking Green instance..."
	@curl -s http://localhost:3503/health || echo "❌ Green: DOWN"

health-check: ## Check all services health
	@echo "=== Infrastructure Health ==="
	@curl -s http://localhost:3602/ | grep -q "version" && echo "✅ Kong: UP" || echo "❌ Kong: DOWN"
	@curl -s http://localhost:3604 > /dev/null && echo "✅ Konga: UP" || echo "❌ Konga: DOWN"
	@echo ""
	@echo "=== Application Health ==="
	@curl -s http://localhost:3502/health > /dev/null && echo "✅ Blue (3502): UP" || echo "❌ Blue: DOWN"
	@curl -s http://localhost:3503/health > /dev/null && echo "✅ Green (3503): UP" || echo "❌ Green: DOWN"

clean: ## Remove stopped containers and volumes
	docker compose down -v
	docker system prune -f

restart-blue: ## Restart blue instance
	docker compose restart portal-be-blue

restart-green: ## Restart green instance
	docker compose restart portal-be-green

stop-blue: ## Stop blue instance
	docker compose stop portal-be-blue

stop-green: ## Stop green instance
	docker compose stop portal-be-green
