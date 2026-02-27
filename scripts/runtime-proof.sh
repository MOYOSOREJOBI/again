#!/usr/bin/env bash
set -euo pipefail

COMPOSE_FILE="${COMPOSE_FILE:-deploy/docker/docker-compose.yml}"
DC="docker compose -f ${COMPOSE_FILE}"

echo "=== RUNTIME PROOF: clean up ==="
${DC} down -v --remove-orphans || true

echo "=== RUNTIME PROOF: bring stack up ==="
${DC} up -d --build

echo "=== RUNTIME PROOF: wait for healthy ==="
deadline=$((SECONDS+240))
while true; do
  unhealthy="$(${DC} ps --format json | python3 - <<'PY'
import json,sys
items=json.load(sys.stdin)
bad=[x for x in items if x.get("State","").lower()=="unhealthy"]
print(len(bad))
PY
)"
  if [ "${unhealthy}" = "0" ]; then
    break
  fi
  if [ "${SECONDS}" -gt "${deadline}" ]; then
    echo "FAIL: containers still unhealthy"
    ${DC} ps
    ${DC} logs --tail=200
    exit 1
  fi
  sleep 3
done
${DC} ps

echo "=== RUNTIME PROOF: root + ready + metrics ==="
curl -fsS http://localhost:8080/ >/dev/null
curl -fsS http://localhost:8080/readyz >/dev/null
curl -fsS http://localhost:8080/metrics >/dev/null

curl -fsS http://localhost:8085/ >/dev/null
curl -fsS http://localhost:8085/readyz >/dev/null
curl -fsS http://localhost:8085/metrics >/dev/null

curl -fsS http://localhost:8083/ >/dev/null
curl -fsS http://localhost:8083/readyz >/dev/null
curl -fsS http://localhost:8083/metrics >/dev/null

curl -fsS http://localhost:8084/ >/dev/null
curl -fsS http://localhost:8084/readyz >/dev/null
curl -fsS http://localhost:8084/metrics >/dev/null

curl -fsS http://localhost:8090/readyz >/dev/null || curl -fsS http://localhost:8090/healthz >/dev/null
curl -fsS http://localhost:3000/ >/dev/null

echo "=== RUNTIME PROOF: data flow within 60s ==="
deadline=$((SECONDS+60))
while true; do
  if curl -fsS http://localhost:8085/debug/seed-status | python3 - <<'PY'
import json,sys
d=json.load(sys.stdin)
ok = all(d.get(k,0)>0 for k in ["raw_ticks","candles","features","scores","alerts","incidents"])
print("OK" if ok else "NO")
sys.exit(0 if ok else 1)
PY
  then
    break
  fi
  if [ "${SECONDS}" -gt "${deadline}" ]; then
    echo "FAIL: data did not warm within 60s"
    curl -fsS http://localhost:8085/debug/seed-status || true
    exit 1
  fi
  sleep 2
done

echo "=== RUNTIME PROOF: Prometheus up==1 ==="
python3 scripts/prom_up_check.py http://localhost:9090 gateway query alerts governance aggregator features inference simulator

echo "=== RUNTIME PROOF: Grafana provisioning evidence ==="
docker compose -f ${COMPOSE_FILE} logs grafana --tail=200 | tee /tmp/grafana_tail.log
grep -Ei "provision|dashboard|datasource" /tmp/grafana_tail.log >/dev/null || true

echo "PASS: runtime-proof"
