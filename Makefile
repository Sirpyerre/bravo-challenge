.PHONY: help run build test swagger migrate clean \
        docker-up docker-down \
        minikube-build minikube-deploy minikube-reload \
        k8s-status k8s-down \
        monitoring-install monitoring-forward monitoring-down

BACKEND_DIR=./backend
GREEN=\033[0;32m
YELLOW=\033[0;33m
NC=\033[0m

help:
	@echo "$(YELLOW)Local$(NC)"
	@echo "  make run              Levantar infra + correr backend"
	@echo "  make build            Compilar binario"
	@echo "  make test             Correr tests"
	@echo "  make swagger          Regenerar docs Swagger"
	@echo "  make migrate          Aplicar migraciones SQL"
	@echo ""
	@echo "$(YELLOW)Docker$(NC)"
	@echo "  make docker-up        Levantar servicios (PG, Redis, RabbitMQ, WireMock)"
	@echo "  make docker-down      Detener servicios"
	@echo ""
	@echo "$(YELLOW)Minikube$(NC)"
	@echo "  make minikube-deploy      Build imagen + configmaps + apply manifiestos"
	@echo "  make minikube-reload      Rebuild imagen y reiniciar pods API"
	@echo "  make k8s-status           Ver estado de pods en namespace bravo"
	@echo "  make k8s-down             Eliminar namespace bravo"
	@echo ""
	@echo "$(YELLOW)Monitoring$(NC)"
	@echo "  make monitoring-install   Instalar Prometheus + Grafana via Helm"
	@echo "  make monitoring-forward   Port-forward Grafana:3000 y Prometheus:9090"
	@echo "  make monitoring-down      Desinstalar stack de monitoreo"

# ── Local ──────────────────────────────────────────────────────────────────────

run: docker-up
	set -a && . ./.env && set +a && cd $(BACKEND_DIR) && go run ./cmd/

build:
	cd $(BACKEND_DIR) && CGO_ENABLED=0 go build -o ./bin/bravo-api ./cmd/

test:
	cd $(BACKEND_DIR) && go test -v -race -cover ./...

swagger:
	cd $(BACKEND_DIR) && swag init --dir ./cmd --generalInfo docs.go --output ./docs --parseDependency --parseInternal

migrate:
	psql -h localhost -U postgres -d bravo -f $(BACKEND_DIR)/migrations/001_init.sql
	psql -h localhost -U postgres -d bravo -f $(BACKEND_DIR)/migrations/002_add_role_to_users.sql
	psql -h localhost -U postgres -d bravo -f $(BACKEND_DIR)/migrations/003_pg_notify_trigger.sql

clean:
	rm -rf $(BACKEND_DIR)/bin
	docker-compose down -v

# ── Docker ─────────────────────────────────────────────────────────────────────

docker-up:
	docker-compose up -d

docker-down:
	docker-compose down

# ── Minikube ───────────────────────────────────────────────────────────────────

minikube-build:
	eval $$(minikube docker-env) && \
	docker build -t bravo-api:latest -f $(BACKEND_DIR)/Dockerfile $(BACKEND_DIR)
	@echo "$(GREEN)✓ bravo-api:latest cargada en minikube$(NC)"

minikube-deploy: minikube-build
	kubectl create configmap wiremock-mappings \
		--from-file=wiremock/mappings/ -n bravo \
		--dry-run=client -o yaml | kubectl apply -f -
	kubectl apply -k k8s/base/
	@echo "$(GREEN)✓ Cluster actualizado$(NC)"

minikube-reload: minikube-build
	kubectl rollout restart deployment/bravo-api -n bravo

k8s-status:
	kubectl get pods -n bravo

k8s-down:
	kubectl delete namespace bravo --ignore-not-found

k8s-migrate: ## Correr migraciones en el PostgreSQL del cluster
	kubectl cp $(BACKEND_DIR)/migrations/001_init.sql bravo/postgres-0:/tmp/001_init.sql
	kubectl cp $(BACKEND_DIR)/migrations/002_add_role_to_users.sql bravo/postgres-0:/tmp/002_add_role_to_users.sql
	kubectl cp $(BACKEND_DIR)/migrations/003_pg_notify_trigger.sql bravo/postgres-0:/tmp/003_pg_notify_trigger.sql
	kubectl exec -n bravo postgres-0 -- psql -U postgres -d bravo -f /tmp/001_init.sql
	kubectl exec -n bravo postgres-0 -- psql -U postgres -d bravo -f /tmp/002_add_role_to_users.sql
	kubectl exec -n bravo postgres-0 -- psql -U postgres -d bravo -f /tmp/003_pg_notify_trigger.sql

k8s-port-forward: ## Acceso local sin tunnel: API en localhost:8000
	kubectl port-forward svc/bravo-api-service 8000:8000 -n bravo

# ── Monitoring (Helm) ───────────────────────────────────────────────────────

monitoring-install: ## Instalar kube-prometheus-stack en namespace monitoring
	helm repo add prometheus-community https://prometheus-community.github.io/helm-charts 2>/dev/null || true
	helm repo update
	kubectl create namespace monitoring --dry-run=client -o yaml | kubectl apply -f -
	helm upgrade --install kube-prometheus-stack prometheus-community/kube-prometheus-stack \
		--namespace monitoring \
		--values k8s/monitoring/values.yaml \
		--wait --timeout 5m
	@echo "$(GREEN)✓ Monitoring stack instalado$(NC)"
	@echo "  Agrega a /etc/hosts:  127.0.0.1 grafana.local"
	@echo "  O usa: make monitoring-forward"

monitoring-forward: ## Port-forward Grafana (3000) y Prometheus (9090) en background
	@echo "$(GREEN)Grafana   -> http://localhost:3000  (admin / admin)$(NC)"
	@echo "$(GREEN)Prometheus-> http://localhost:9090$(NC)"
	kubectl port-forward svc/kube-prometheus-stack-grafana 3000:80 -n monitoring &
	kubectl port-forward svc/kube-prometheus-stack-prometheus 9090:9090 -n monitoring

monitoring-down: ## Desinstalar kube-prometheus-stack y borrar namespace monitoring
	helm uninstall kube-prometheus-stack -n monitoring --ignore-not-found
	kubectl delete namespace monitoring --ignore-not-found
	@echo "$(GREEN)✓ Monitoring stack eliminado$(NC)"
