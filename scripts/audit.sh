#!/usr/bin/env bash
set -euo pipefail

TS="$(date -u +%Y%m%dT%H%M%SZ)"
OUT="docs/audit/${TS}"
LOGS="${OUT}/logs"
mkdir -p "$LOGS"

run_and_log() {
  local name="$1"; shift
  echo "\$ $*" | tee "${LOGS}/${name}.txt"
  if "$@" >>"${LOGS}/${name}.txt" 2>&1; then
    return 0
  fi
  return 1
}

# Phase 0 baseline commands (always logged)
doctor_ok=0
lint_ok=0
test_ok=0
run_and_log make_doctor make doctor && doctor_ok=1
run_and_log make_lint make lint && lint_ok=1
run_and_log make_test make test && test_ok=1

# Main gate execution
gate_ok=0
run_and_log gate ./scripts/gate.sh && gate_ok=1

# Parse explicit gate summary signal if present
playwright_ok=0
runtime_ok=0
if summary_line=$(rg -n "GATE_SUMMARY" "${LOGS}/gate.txt" -N 2>/dev/null | tail -n1); then
  if echo "$summary_line" | rg -q "playwright=1"; then playwright_ok=1; fi
  if echo "$summary_line" | rg -q "runtime=1"; then runtime_ok=1; fi
fi

docker_ok=0
if command -v docker >/dev/null 2>&1 && docker info >/dev/null 2>&1; then
  docker_ok=1
fi

# Success bars weights (exact rubric)
w_s1=12; w_s2=10; w_s3=6; w_s4=12; w_s5=8; w_s6=16; w_s7=12; w_s8=10; w_s9=8; w_s10=6

# PASS bars as strict booleans
s1=$([ "$doctor_ok" -eq 1 ] && [ "$lint_ok" -eq 1 ] && [ "$test_ok" -eq 1 ] && [ "$gate_ok" -eq 1 ] && echo 1 || echo 0)
s2=$([ "$runtime_ok" -eq 1 ] && echo 1 || echo 0)
s3=$([ "$runtime_ok" -eq 1 ] && echo 1 || echo 0)
s4=$([ "$runtime_ok" -eq 1 ] && echo 1 || echo 0)
s5=$([ "$runtime_ok" -eq 1 ] && echo 1 || echo 0)
s6=$([ "$runtime_ok" -eq 1 ] && echo 1 || echo 0)
s7=$([ "$gate_ok" -eq 1 ] && echo 1 || echo 0)
s8=$([ "$gate_ok" -eq 1 ] && echo 1 || echo 0)
s9=$([ "$gate_ok" -eq 1 ] && echo 1 || echo 0)
s10=$([ "$playwright_ok" -eq 1 ] && echo 1 || echo 0)

score=$((s1*w_s1 + s2*w_s2 + s3*w_s3 + s4*w_s4 + s5*w_s5 + s6*w_s6 + s7*w_s7 + s8*w_s8 + s9*w_s9 + s10*w_s10))
status="FAIL"
if [ "$score" -eq 100 ]; then
  status="PASS"
fi
runtime="NONE"
if [ "$runtime_ok" -eq 1 ]; then
  runtime="FULL"
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
- gate_ok=${gate_ok}
- docker_ok=${docker_ok}
- runtime_ok=${runtime_ok}
- playwright_ok=${playwright_ok}

Primary evidence: logs/gate.txt
MD

cat > "${OUT}/TRACEABILITY.md" <<MD
- baseline: make doctor / make lint / make test
- gate: ./scripts/gate.sh
- docker runtime smoke: ./scripts/demo-smoke.sh
- playwright docker runner: ./scripts/playwright-docker.sh
- prometheus check: ./scripts/prom_up_check.py
- logs dir: ${LOGS}
MD

echo "COMPLETION: ${score}%"
echo "STATUS: ${status}"
echo "RUNTIME COVERAGE: ${runtime}"
echo "Artifacts: ${OUT}"

[ "$status" = "PASS" ]
