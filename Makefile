COMPOSE=docker compose -f deploy/docker/docker-compose.yml
FAST_COMPOSE=docker compose -f deploy/docker/docker-compose.yml -f deploy/docker/docker-compose.fast.yml

.PHONY: help doctor lint dev-keys up down migrate topics seed demo demo-fast test unit integration-test integration-suite replay-test security-test demo-smoke verify logs status

help: ## Show available targets
	@echo "Sentinel Platform - Available targets:"
	@echo "  make unit             Run Go + Python + Web tests from repo root"
	@echo "  make verify           Run unit + security + replay + demo smoke"

doctor:
	./scripts/doctor.sh

lint:
	./scripts/lint.sh

dev-keys:
	./scripts/dev-keys.sh

up:
	$(COMPOSE) up -d --build

up-fast:
	$(FAST_COMPOSE) up -d --build

down:
	$(COMPOSE) down -v

migrate:
	./scripts/migrate.sh

topics:
	./scripts/create-topics.sh

seed:
	POSTGRES_URL=postgres://sentinel:sentinel@localhost:5432/sentinel?sslmode=disable go run ./scripts/seed-users.go

demo:
	./scripts/run-demo.sh default

demo-fast:
	./scripts/run-demo.sh fast

test:
	go test -race -count=1 ./...

unit:
	go test ./...
	PYTHONPATH=services/inference python3 -m pytest -q services/inference/tests services/inference/test_app.py
	cd web && npm test -- --runInBand

integration-test:
	./scripts/integration-test.sh

integration-suite:
	./integration/e2e_pipeline_test.sh
	./integration/idempotency_test.sh
	./integration/replay_equivalence_test.sh
	./integration/startup_ordering_test.sh
	./integration/dependency_failure_test.sh

replay-test:
	./integration/replay_equivalence_test.sh

security-test:
	./integration/dependency_failure_test.sh

demo-smoke:
	./integration/e2e_pipeline_test.sh

verify:
	$(MAKE) unit
	$(MAKE) security-test
	$(MAKE) replay-test
	$(MAKE) demo-smoke

logs:
	$(COMPOSE) logs -f --tail=50

status:
	$(COMPOSE) ps
