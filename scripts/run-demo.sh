#!/usr/bin/env bash
set -euo pipefail
mode=${1:-default}

wait_for() {
  local name=$1 url=$2 max=${3:-60}
  echo "[wait] $name -> $url"
  for i in $(seq 1 "$max"); do
    if curl -fsS "$url" >/dev/null 2>&1; then
      echo "[ok] $name"
      return 0
    fi
    sleep 2
  done
  echo "[fail] $name did not become ready: $url"
  return 1
}

if ! command -v docker >/dev/null 2>&1; then
  echo "SKIP: demo-smoke skipped because docker is not available in this environment"
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

wait_for gateway http://localhost:8080/healthz
wait_for query http://localhost:8085/readyz
wait_for alerts http://localhost:8083/readyz
wait_for governance http://localhost:8084/readyz
wait_for command-center http://localhost:3000/command-center 80

COOKIE_JAR=$(mktemp)
LOGIN_JSON='{"Email":"admin@sentinel.local","Password":"Sentinel#123"}'
curl -fsS -c "$COOKIE_JAR" -X POST http://localhost:8080/auth/login -H 'Content-Type: application/json' -d "$LOGIN_JSON" >/dev/null
CSRF=$(awk '/sentinel_csrf/ {print $7}' "$COOKIE_JAR" | tail -n1)
[ -n "$CSRF" ]
for ep in command-center queue trust world-map replay/demo governance/summary executive-summary cases; do
  curl -fsS -b "$COOKIE_JAR" "http://localhost:8085/$ep" >/dev/null
  echo "[ok] query/$ep"
done
rm -f "$COOKIE_JAR"

echo "[pass] Sentinel stack ready ($mode)"
./scripts/browser-validate.sh || { echo "[fail] browser validation"; exit 1; }
./scripts/capture-screenshots.sh || { echo "[fail] screenshot capture"; exit 1; }
echo "PASS: Sentinel Demo Ready ($mode)"
