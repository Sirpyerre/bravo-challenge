.PHONY: help docker-up docker-down run build test migrate seed logs clean

BACKEND_DIR=./backend
FRONTEND_DIR=./frontend

# Colors for output
GREEN=\033[0;32m
YELLOW=\033[0;33m
NC=\033[0m # No Color

help:
	@echo "$(GREEN)Bravo Project - Available Commands$(NC)"
	@echo ""
	@echo "$(YELLOW)Docker$(NC)"
	@echo "  make docker-up          Start all services (PostgreSQL, RabbitMQ, Redis, WireMock)"
	@echo "  make docker-down        Stop all services"
	@echo "  make docker-logs        View Docker logs"
	@echo ""
	@echo "$(YELLOW)Backend$(NC)"
	@echo "  make run                Run backend in development"
	@echo "  make build              Build backend binary"
	@echo "  make test               Run tests"
	@echo "  make migrate            Run database migrations"
	@echo "  make seed               Seed database with sample data"
	@echo ""
	@echo "$(YELLOW)Frontend$(NC)"
	@echo "  make frontend-install   Install frontend dependencies"
	@echo "  make frontend-start     Start frontend dev server"
	@echo "  make frontend-build     Build frontend for production"
	@echo ""
	@echo "$(YELLOW)Development$(NC)"
	@echo "  make clean              Clean build artifacts and temp files"
	@echo "  make help               Show this help message"

# Docker commands

docker-build: ## Build the API Docker image
	docker build -t bravo-api:latest -f $(BACKEND_DIR)/Dockerfile $(BACKEND_DIR)
	@echo "$(GREEN)✓ Docker image built: bravo-api:latest$(NC)"

docker-up:
	docker-compose up -d
	@echo "$(GREEN)✓ Services are up$(NC)"
	@echo "  PostgreSQL: localhost:5432"
	@echo "  RabbitMQ:   localhost:5672 (Admin: http://localhost:15672)"
	@echo "  Redis:      localhost:6379"
	@echo "  WireMock:   http://localhost:8080"

docker-down:
	docker-compose down
	@echo "$(GREEN)✓ Services are down$(NC)"

docker-logs:
	docker-compose logs -f

# Backend commands
run: docker-up
	set -a && . ./.env && set +a && cd $(BACKEND_DIR) && go run ./cmd/

build:
	cd $(BACKEND_DIR) && CGO_ENABLED=1 go build -o ./bin/bravo-api ./cmd/

test:
	cd $(BACKEND_DIR) && go test -v -race -cover ./...

swagger:
	cd $(BACKEND_DIR) && swag init --dir ./cmd --generalInfo docs.go --output ./docs --parseDependency --parseInternal

# Kubernetes
k8s-up:
	kubectl apply -k k8s/base/

k8s-down:
	kubectl delete namespace bravo --ignore-not-found

k8s-status:
	kubectl get all -n bravo

migrate:
	cd $(BACKEND_DIR) && psql -h localhost -U postgres -d bravo -f migrations/001_init.sql
	@echo "$(GREEN)✓ Migrations completed$(NC)"

seed:
	cd $(BACKEND_DIR) && go run ./scripts/seed.go

# Frontend commands
frontend-install:
	cd $(FRONTEND_DIR) && npm install

frontend-start: frontend-install
	cd $(FRONTEND_DIR) && npm start

frontend-build:
	cd $(FRONTEND_DIR) && npm run build

# Cleaning
clean:
	rm -rf $(BACKEND_DIR)/bin
	rm -rf $(FRONTEND_DIR)/build
	rm -rf $(FRONTEND_DIR)/dist
	rm -rf $(BACKEND_DIR)/*.log
	docker-compose down -v
	@echo "$(GREEN)✓ Clean completed$(NC)"

# Development workflow
dev: docker-up
	@echo "$(GREEN)Starting development environment...$(NC)"
	@echo "Backend will start in 5 seconds..."
	@sleep 5
	cd $(BACKEND_DIR) && go run ./cmd/main.go &
	@echo "$(GREEN)Backend started. Starting frontend...$(NC)"
	@sleep 2
	cd $(FRONTEND_DIR) && npm start

# Check if services are healthy
health-check:
	@echo "Checking services..."
	@pg_isready -h localhost -p 5432 || echo "PostgreSQL: ✗"
	@echo "PostgreSQL: ✓"
	@curl -s http://localhost:15672 > /dev/null && echo "RabbitMQ: ✓" || echo "RabbitMQ: ✗"
	@curl -s http://localhost:6379 > /dev/null && echo "Redis: ✓" || echo "Redis: ✗"
	@curl -s http://localhost:8080 > /dev/null && echo "WireMock: ✓" || echo "WireMock: ✗"

# Development info
info:
	@echo "$(YELLOW)Bravo Development Environment$(NC)"
	@echo ""
	@echo "URLs:"
	@echo "  Frontend:     http://localhost:3000"
	@echo "  Backend API:  http://localhost:8080"
	@echo "  RabbitMQ:     http://localhost:15672 (guest/guest)"
	@echo "  PostgreSQL:   localhost:5432 (postgres/postgres)"
	@echo "  Redis:        localhost:6379"
	@echo "  WireMock:     http://localhost:8080"
	@echo ""
	@echo "Useful commands:"
	@echo "  make docker-up          - Start services"
	@echo "  make run                - Start backend"
	@echo "  make frontend-start     - Start frontend"
	@echo "  make test               - Run tests"
	@echo "  make migrate            - Run migrations"
