#!/usr/bin/env bash
set -euo pipefail
mode=${1:-default}
if ! command -v docker >/dev/null 2>&1; then
  echo "demo skipped: docker is not available in this environment"
  exit 0
fi
make dev-keys
if [ "$mode" = "fast" ]; then
  COMPOSE="docker compose -f deploy/docker/docker-compose.yml -f deploy/docker/docker-compose.fast.yml"
else
  COMPOSE="docker compose -f deploy/docker/docker-compose.yml"
fi

# Bring up infrastructure first so topics and migrations are deterministic.
$COMPOSE up -d redpanda postgres redis
./scripts/migrate.sh
./scripts/create-topics.sh

# Start application services only after prerequisites exist.
$COMPOSE up -d --build gateway-api simulator aggregator features inference alerts governance query web prometheus grafana
make seed
echo "Sentinel Demo Ready ($mode)"
