#!/usr/bin/env bash
set -euo pipefail

TS="$(date -u +%Y%m%dT%H%M%SZ)"
TS_DIR="docs/audit/${TS}"
PROOF_DIR="${TS_DIR}/proofs"
LOG_DIR="${TS_DIR}/logs"
mkdir -p "${PROOF_DIR}" "${LOG_DIR}"

ln -sfn "${TS}" docs/audit/_latest

set +e
TS_DIR="${TS_DIR}" ./scripts/gate.sh 2>&1 | tee "${LOG_DIR}/gate.log"
GATE_EC=${PIPESTATUS[0]}
set -e

mkdir -p "${PROOF_DIR}" "${LOG_DIR}"

has() {
  [ -f "${PROOF_DIR}/$1" ] && echo 1 || echo 0
}

s2="$(has S2_compose_healthy.ok)"
s3="$(has S3_root_routes.ok)"
s4="$(has S4_prom_up.ok)"
s5="$(has S5_grafana_provisioned.ok)"
s6="$(has S6_data_warm.ok)"
s10="$(has S10_playwright.ok)"

passed=$((s2+s3+s4+s5+s6+s10))
score=$((passed*100/6))
status="FAIL"
runtime="NONE"
if [ "${passed}" -eq 6 ]; then
  status="PASS"
  runtime="FULL"
fi

cat > "${TS_DIR}/RESULT.md" <<MD
COMPLETION: ${score}%
STATUS: ${status}
RUNTIME COVERAGE: ${runtime}
VERDICT: ${verdict}

S2=${s2} S3=${s3} S4=${s4} S5=${s5} S6=${s6} S10=${s10}

GATE_EXIT=${GATE_EC}
PROOFS=${PROOF_DIR}
LOGS=${LOG_DIR}
MD

cat > "${TS_DIR}/TRACEABILITY.md" <<MD
# Traceability

- ts_dir: ${TS_DIR}
- ran: TS_DIR=${TS_DIR} ./scripts/gate.sh
- scoring source: proof markers in ${PROOF_DIR}
- required markers:
  - S2_compose_healthy.ok
  - S3_root_routes.ok
  - S4_prom_up.ok
  - S5_grafana_provisioned.ok
  - S6_data_warm.ok
  - S10_playwright.ok
- logs: ${LOG_DIR}
MD

ln -sfn "${TS}" docs/audit/_latest
cat "${TS_DIR}/RESULT.md"

[ "${score}" -eq 100 ]
