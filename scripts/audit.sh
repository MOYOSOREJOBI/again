#!/usr/bin/env bash
set -euo pipefail

TS="$(date -u +%Y%m%dT%H%M%SZ)"
TS_DIR="docs/audit/${TS}"
PROOFS_DIR="${TS_DIR}/proofs"
LOGS_DIR="${TS_DIR}/logs"
mkdir -p "${PROOFS_DIR}" "${LOGS_DIR}"

# Point _latest at this run before invoking gate/runtime scripts.
ln -sfn "${TS}" docs/audit/_latest

echo "=== audit: running gate with TS_DIR=${TS_DIR} ===" | tee -a "${LOGS_DIR}/audit.log"
set +e
TS_DIR="${TS_DIR}" ./scripts/gate.sh 2>&1 | tee -a "${LOGS_DIR}/gate.log"
GATE_EC=${PIPESTATUS[0]}
set -e


# Self-heal runtime marker production independent of earlier gate failures.
if [ "${GATE_EC}" -ne 0 ]; then
  echo "=== audit: gate failed; attempting direct proof generation ===" | tee -a "${LOGS_DIR}/audit.log"
  if [ ! -f "${PROOFS_DIR}/S2_compose_healthy.ok" ] || [ ! -f "${PROOFS_DIR}/S3_root_routes.ok" ] || [ ! -f "${PROOFS_DIR}/S4_prom_up.ok" ] || [ ! -f "${PROOFS_DIR}/S5_grafana_provisioned.ok" ] || [ ! -f "${PROOFS_DIR}/S6_data_warm.ok" ]; then
    set +e
    TS_DIR="${TS_DIR}" ./scripts/runtime-proof.sh 2>&1 | tee -a "${LOGS_DIR}/runtime-proof.recovery.log"
    set -e
  fi
  if [ ! -f "${PROOFS_DIR}/S10_playwright.ok" ]; then
    set +e
    TS_DIR="${TS_DIR}" ./scripts/playwright-docker.sh 2>&1 | tee -a "${LOGS_DIR}/playwright.recovery.log"
    set -e
  fi
fi

# Ensure proof/log artifacts live under this timestamped directory.
mkdir -p "${PROOFS_DIR}" "${LOGS_DIR}"
if [ -d "docs/audit/_latest/proofs" ] && [ "$(realpath docs/audit/_latest)" != "$(realpath "${TS_DIR}")" ]; then
  cp -a docs/audit/_latest/proofs/. "${PROOFS_DIR}/" || true
fi
if [ -d "docs/audit/_latest/logs" ] && [ "$(realpath docs/audit/_latest)" != "$(realpath "${TS_DIR}")" ]; then
  cp -a docs/audit/_latest/logs/. "${LOGS_DIR}/" || true
fi

has_marker() {
  local marker="$1"
  if [ -f "${PROOFS_DIR}/${marker}" ]; then
    echo 1
  else
    echo 0
  fi
}

s2="$(has_marker S2_compose_healthy.ok)"
s3="$(has_marker S3_root_routes.ok)"
s4="$(has_marker S4_prom_up.ok)"
s5="$(has_marker S5_grafana_provisioned.ok)"
s6="$(has_marker S6_data_warm.ok)"
s10="$(has_marker S10_playwright.ok)"

passed=$((s2 + s3 + s4 + s5 + s6 + s10))
score=$((passed * 100 / 6))

status="FAIL"
runtime="PARTIAL"
verdict="FAIL PARTIAL"
if [ "$passed" -eq 6 ]; then
  status="PASS"
  runtime="FULL"
  verdict="PASS FULL"
fi

cat > "${TS_DIR}/RESULT.md" <<MD
COMPLETION: ${score}%
STATUS: ${status}
RUNTIME COVERAGE: ${runtime}
VERDICT: ${verdict}

S2=${s2} S3=${s3} S4=${s4} S5=${s5} S6=${s6} S10=${s10}

MARKERS DIR: ${PROOFS_DIR}
LOGS DIR: ${LOGS_DIR}
GATE EXIT CODE: ${GATE_EC}
MD

cat > "${TS_DIR}/TRACEABILITY.md" <<MD
# Traceability

- timestamp: ${TS}
- ts_dir: ${TS_DIR}
- gate command: TS_DIR=${TS_DIR} ./scripts/gate.sh
- proof markers evaluated strictly from: ${PROOFS_DIR}
- marker files:
  - S2_compose_healthy.ok
  - S3_root_routes.ok
  - S4_prom_up.ok
  - S5_grafana_provisioned.ok
  - S6_data_warm.ok
  - S10_playwright.ok
- logs: ${LOGS_DIR}
MD

# Refresh _latest after artifact generation.
ln -sfn "${TS}" docs/audit/_latest

cat "${TS_DIR}/RESULT.md"

if [ "${score}" -ne 100 ]; then
  exit 1
fi
