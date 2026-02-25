COMPOSE=docker compose -f deploy/docker/docker-compose.yml
FAST_COMPOSE=docker compose -f deploy/docker/docker-compose.yml -f deploy/docker/docker-compose.fast.yml

.PHONY: help doctor lint dev-keys up down migrate topics seed demo demo-fast test integration-test integration-suite logs status

help: ## Show available targets
	@echo "Sentinel Platform - Available targets:"
	@echo ""
	@echo "  make doctor           Check prerequisites (docker, openssl, ports)"
	@echo "  make lint             Run Go formatting and Python syntax checks"
	@echo "  make dev-keys         Generate JWT RSA key pair for development"
	@echo "  make up               Start all services with docker compose"
	@echo "  make up-fast          Start services in fast mode (fewer symbols)"
	@echo "  make down             Stop all services and remove volumes"
	@echo "  make migrate          Run database migrations"
	@echo "  make topics           Create Kafka topics in Redpanda"
	@echo "  make seed             Seed demo user accounts"
	@echo "  make demo             Full demo: keys + build + migrate + topics + seed"
	@echo "  make demo-fast        Fast demo with reduced symbols"
	@echo "  make test             Run Go unit tests"
	@echo "  make integration-test Run integration tests (requires docker)"
	@echo "  make logs             Tail logs from all services"
	@echo "  make status           Show running container status"
	@echo ""

doctor: ## Check prerequisites
	./scripts/doctor.sh

lint: ## Run linters
	./scripts/lint.sh

dev-keys: ## Generate JWT dev keys
	./scripts/dev-keys.sh

up: ## Start all services
	$(COMPOSE) up -d --build

up-fast: ## Start in fast mode
	$(FAST_COMPOSE) up -d --build

down: ## Stop all services
	$(COMPOSE) down -v

migrate: ## Run database migrations
	./scripts/migrate.sh

topics: ## Create Kafka topics
	./scripts/create-topics.sh

seed: ## Seed demo users
	POSTGRES_URL=postgres://sentinel:sentinel@localhost:5432/sentinel?sslmode=disable go run ./scripts/seed-users.go

demo: ## Run full demo
	./scripts/run-demo.sh default
	@echo ""
	@echo "=== Sentinel Demo Ready ==="
	@echo "  Web Dashboard:  http://localhost:3000"
	@echo "  Gateway API:    http://localhost:8080"
	@echo "  Query API:      http://localhost:8085"
	@echo "  Alerts SSE:     http://localhost:8083/sse/alerts"
	@echo "  Grafana:        http://localhost:3001"
	@echo "  Prometheus:     http://localhost:9090"
	@echo ""
	@echo "  Login:  admin@sentinel.local / Sentinel#123"
	@echo "=========================="

demo-fast: ## Run fast demo
	./scripts/run-demo.sh fast
	@echo "Sentinel Demo Fast Ready"

test: ## Run unit tests
	go test -race -count=1 ./...

integration-test: ## Run integration tests
	./scripts/integration-test.sh

logs: ## Tail all service logs
	$(COMPOSE) logs -f --tail=50

status: ## Show container status
	$(COMPOSE) ps


integration-suite: ## Run phase2 integration checks
	./integration/e2e_pipeline_test.sh
	./integration/idempotency_test.sh
	./integration/replay_equivalence_test.sh
	./integration/startup_ordering_test.sh
	./integration/dependency_failure_test.sh
