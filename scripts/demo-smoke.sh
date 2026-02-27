#!/usr/bin/env bash
set -euo pipefail

COMPOSE_FILE="${COMPOSE_FILE:-deploy/docker/docker-compose.yml}"
DC="docker compose -f ${COMPOSE_FILE}"

echo "=== DEMO SMOKE: bring stack up ==="
${DC} up -d --build

echo "=== DEMO SMOKE: wait for services to be healthy ==="
deadline=$((SECONDS+180))
while true; do
  bad_count="$(${DC} ps --format json | python3 - <<'PY'
import json,sys
items=json.load(sys.stdin)
bad=[x for x in items if x.get("State","").lower()=="unhealthy"]
print(len(bad))
PY
)"
  if [ "${bad_count}" = "0" ]; then
    echo "All containers not reporting unhealthy."
    break
  fi
  if [ "${SECONDS}" -gt "${deadline}" ]; then
    echo "FAIL: containers still unhealthy"
    ${DC} ps
    exit 1
  fi
  sleep 3
done

echo "=== DEMO SMOKE: basic HTTP readiness ==="
curl -fsS http://localhost:8080/readyz >/dev/null
curl -fsS http://localhost:8085/readyz >/dev/null
curl -fsS http://localhost:8083/readyz >/dev/null
curl -fsS http://localhost:8084/readyz >/dev/null
curl -fsS http://localhost:8090/readyz >/dev/null || curl -fsS http://localhost:8090/healthz >/dev/null
curl -fsS http://localhost:3000/ >/dev/null

echo "=== DEMO SMOKE: Prometheus up==1 for core jobs ==="
python3 scripts/prom_up_check.py http://localhost:9090 gateway query alerts governance aggregator features inference simulator

echo "PASS: demo smoke complete"
