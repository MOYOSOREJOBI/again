#!/usr/bin/env bash
set -euo pipefail

COMPOSE_FILE="${COMPOSE_FILE:-deploy/docker/docker-compose.yml}"
DC="docker compose -f ${COMPOSE_FILE}"

TS_DIR="${TS_DIR:-docs/audit/_latest}"
PROOF_DIR="${TS_DIR}/proofs"
LOG_DIR="${TS_DIR}/logs"
mkdir -p "${PROOF_DIR}" "${LOG_DIR}"

log() { tee -a "${LOG_DIR}/runtime-proof.log"; }

echo "=== clean up ===" | log
${DC} down -v --remove-orphans >>"${LOG_DIR}/compose_down.log" 2>&1 || true

echo "=== up ===" | log
${DC} up -d --build >>"${LOG_DIR}/compose_up.log" 2>&1

echo "=== wait healthy (robust) ===" | log
deadline=$((SECONDS+300))
while true; do
  # robust: inspect health directly (no reliance on compose json formatting)
  bad=$(${DC} ps -q | while read -r cid; do
    st=$(docker inspect -f '{{if .State.Health}}{{.State.Health.Status}}{{else}}{{.State.Status}}{{end}}' "$cid")
    name=$(docker inspect -f '{{.Name}}' "$cid" | sed 's#^/##')
    if [ "$st" != "healthy" ] && [ "$st" != "running" ]; then
      echo "$name:$st"
    fi
  done | wc -l | tr -d ' ')
  if [ "${bad}" = "0" ]; then break; fi
  if [ "${SECONDS}" -gt "${deadline}" ]; then
    echo "FAIL: containers not healthy" | log
    ${DC} ps >>"${LOG_DIR}/compose_ps.log" 2>&1 || true
    ${DC} logs --tail=300 >>"${LOG_DIR}/compose_logs_tail.log" 2>&1 || true
    exit 1
  fi
  sleep 3
done

${DC} ps >>"${LOG_DIR}/compose_ps.log" 2>&1
touch "${PROOF_DIR}/S2_compose_healthy.ok"

echo "=== root/ready/metrics ===" | log
curl -fsS http://localhost:8080/ >/dev/null
curl -fsS http://localhost:8085/ >/dev/null
curl -fsS http://localhost:8083/ >/dev/null
curl -fsS http://localhost:8084/ >/dev/null
touch "${PROOF_DIR}/S3_root_routes.ok"

curl -fsS http://localhost:8080/readyz >/dev/null
curl -fsS http://localhost:8085/readyz >/dev/null
curl -fsS http://localhost:8083/readyz >/dev/null
curl -fsS http://localhost:8084/readyz >/dev/null
curl -fsS http://localhost:8080/metrics >/dev/null
curl -fsS http://localhost:8085/metrics >/dev/null
curl -fsS http://localhost:8083/metrics >/dev/null
curl -fsS http://localhost:8084/metrics >/dev/null

echo "=== web reachable ===" | log
curl -fsS http://localhost:3000/ >/dev/null

echo "=== data warm <= 90s ===" | log
deadline=$((SECONDS+90))
while true; do
  if curl -fsS http://localhost:8085/debug/seed-status | tee "${LOG_DIR}/seed-status.json" | python3 - <<'PY'
import json,sys
d=json.load(sys.stdin)
req=["raw_ticks","candles","features","scores","alerts","incidents"]
ok=all(int(d.get(k,0))>0 for k in req)
sys.exit(0 if ok else 1)
PY
  then break; fi
  if [ "${SECONDS}" -gt "${deadline}" ]; then
    echo "FAIL: warm data not present" | log
    exit 1
  fi
  sleep 2
done
touch "${PROOF_DIR}/S6_data_warm.ok"

echo "=== Prometheus up==1 ===" | log
python3 scripts/prom_up_check.py http://localhost:9090 --from-config prometheus/prometheus.yml \
  | tee "${LOG_DIR}/prom_up_check.txt"
touch "${PROOF_DIR}/S4_prom_up.ok"

echo "=== Grafana provision proof via HTTP API ===" | log
# Grafana Health API is /api/health; use basic auth if configured.
# Prefer creds from compose env; if not present, default admin:admin.
GF_USER="${GF_USER:-admin}"
GF_PASS="${GF_PASS:-admin}"
curl -fsS "http://${GF_USER}:${GF_PASS}@localhost:3001/api/health" | tee "${LOG_DIR}/grafana_health.json" >/dev/null
# Confirm dashboards exist (search API):
curl -fsS "http://${GF_USER}:${GF_PASS}@localhost:3001/api/search?type=dash-db" | tee "${LOG_DIR}/grafana_search.json" >/dev/null
python3 - <<'PY'
import json
d=json.load(open("docs/audit/_latest/logs/grafana_search.json"))
assert isinstance(d,list) and len(d)>0, "no dashboards found"
PY
touch "${PROOF_DIR}/S5_grafana_provisioned.ok"

echo "PASS runtime-proof" | log
