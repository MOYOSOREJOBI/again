#!/usr/bin/env bash
set -euo pipefail

TS="$(date -u +%Y%m%dT%H%M%SZ)"
OUT="docs/audit/${TS}"
LOGS="${OUT}/logs"
PROOF="${OUT}/proof"
mkdir -p "${LOGS}" "${PROOF}"

run_and_log() {
  local name="$1"; shift
  echo "$ $*" | tee "${LOGS}/${name}.txt"
  "$@" 2>&1 | tee -a "${LOGS}/${name}.txt"
}

doctor_ok=0
lint_ok=0
test_ok=0
demo_ok=0
gate_ok=0

run_and_log make_doctor make doctor && doctor_ok=1
run_and_log make_lint make lint && lint_ok=1
run_and_log make_test make test && test_ok=1

# demo is part of S1
if run_and_log make_demo make demo; then
  demo_ok=1
fi

if AUDIT_OUT="${OUT}" run_and_log gate ./scripts/gate.sh; then
  gate_ok=1
fi

mark_ok() {
  [ -f "$1" ] && echo 1 || echo 0
}

s1=$([ "$doctor_ok" -eq 1 ] && [ "$lint_ok" -eq 1 ] && [ "$test_ok" -eq 1 ] && [ "$demo_ok" -eq 1 ] && [ "$gate_ok" -eq 1 ] && echo 1 || echo 0)
s2=$(mark_ok "${PROOF}/s2_compose_healthy.ok")
s3=$(mark_ok "${PROOF}/s3_routes.ok")
s4=$(mark_ok "${PROOF}/s4_prometheus.ok")
s5=$(mark_ok "${PROOF}/s5_grafana.ok")
s6=$(mark_ok "${PROOF}/s6_dataflow.ok")
# S7/S8 are asserted by build+tests+gate success
s7=$([ "$gate_ok" -eq 1 ] && echo 1 || echo 0)
s8=$([ "$gate_ok" -eq 1 ] && echo 1 || echo 0)
s9=$(mark_ok "${PROOF}/s9_replay.ok")
s10=$(mark_ok "${PROOF}/s10_playwright.ok")

w_s1=12; w_s2=10; w_s3=6; w_s4=12; w_s5=8; w_s6=16; w_s7=12; w_s8=10; w_s9=8; w_s10=6
score=$((s1*w_s1 + s2*w_s2 + s3*w_s3 + s4*w_s4 + s5*w_s5 + s6*w_s6 + s7*w_s7 + s8*w_s8 + s9*w_s9 + s10*w_s10))

status="FAIL"
runtime="NONE"
if [ "$s2" -eq 1 ] && [ "$s3" -eq 1 ] && [ "$s4" -eq 1 ] && [ "$s5" -eq 1 ] && [ "$s6" -eq 1 ]; then
  runtime="FULL"
fi
if [ "$score" -eq 100 ]; then
  status="PASS"
fi

cat > "${OUT}/RESULT.md" <<MD
COMPLETION: ${score}%
STATUS: ${status}
RUNTIME COVERAGE: ${runtime}

S1=${s1} S2=${s2} S3=${s3} S4=${s4} S5=${s5} S6=${s6} S7=${s7} S8=${s8} S9=${s9} S10=${s10}

Signals:
- doctor_ok=${doctor_ok}
- lint_ok=${lint_ok}
- test_ok=${test_ok}
- demo_ok=${demo_ok}
- gate_ok=${gate_ok}
- proof_dir=${PROOF}
MD

cat > "${OUT}/TRACEABILITY.md" <<MD
# Traceability

- make doctor
- make lint
- make test
- make demo
- gate: ./scripts/gate.sh
- runtime: ./scripts/runtime-proof.sh
- playwright: ./scripts/playwright-docker.sh
- prom check: python3 scripts/prom_up_check.py http://localhost:9090 gateway query alerts governance aggregator features inference simulator
- logs: ${LOGS}
- proofs: ${PROOF}
MD

echo "COMPLETION: ${score}%"
echo "STATUS: ${status}"
echo "RUNTIME COVERAGE: ${runtime}"

[ "${status}" = "PASS" ]
