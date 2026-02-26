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

$COMPOSE up -d redpanda postgres redis
./scripts/migrate.sh
./scripts/create-topics.sh
$COMPOSE up -d --build gateway-api simulator aggregator features inference alerts governance query web prometheus grafana
make seed

curl -fsS http://localhost:8080/healthz >/dev/null
COOKIE_JAR=$(mktemp)
LOGIN_JSON='{"Email":"admin@sentinel.local","Password":"Sentinel#123"}'
curl -fsS -c "$COOKIE_JAR" -X POST http://localhost:8080/auth/login -H 'Content-Type: application/json' -d "$LOGIN_JSON" >/dev/null
CSRF=$(awk '/sentinel_csrf/ {print $7}' "$COOKIE_JAR" | tail -n1)
[ -n "$CSRF" ]
curl -fsS -b "$COOKIE_JAR" http://localhost:8085/queue >/dev/null
curl -fsS -b "$COOKIE_JAR" http://localhost:8085/trust >/dev/null
curl -fsS -b "$COOKIE_JAR" http://localhost:8085/world-map >/dev/null
rm -f "$COOKIE_JAR"

echo "Sentinel Demo Ready ($mode)"
