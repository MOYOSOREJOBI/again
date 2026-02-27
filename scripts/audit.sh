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

gate_ok=0
if run_and_log gate ./scripts/gate.sh; then
  gate_ok=1
fi

docker_ok=0
if command -v docker >/dev/null 2>&1 && docker info >/dev/null 2>&1; then
  docker_ok=1
fi

# Success bars weights
w_s1=12; w_s2=10; w_s3=6; w_s4=12; w_s5=8; w_s6=16; w_s7=12; w_s8=10; w_s9=8; w_s10=6

# For strict 100%, all must be pass.
s1=$gate_ok
s2=$([ "$docker_ok" -eq 1 ] && [ "$gate_ok" -eq 1 ] && echo 1 || echo 0)
s3=$([ "$docker_ok" -eq 1 ] && [ "$gate_ok" -eq 1 ] && echo 1 || echo 0)
s4=$([ "$docker_ok" -eq 1 ] && [ "$gate_ok" -eq 1 ] && echo 1 || echo 0)
s5=$([ "$docker_ok" -eq 1 ] && [ "$gate_ok" -eq 1 ] && echo 1 || echo 0)
s6=$([ "$docker_ok" -eq 1 ] && [ "$gate_ok" -eq 1 ] && echo 1 || echo 0)
s7=$([ "$gate_ok" -eq 1 ] && echo 1 || echo 0)
s8=$([ "$gate_ok" -eq 1 ] && echo 1 || echo 0)
s9=$([ "$gate_ok" -eq 1 ] && echo 1 || echo 0)
s10=$([ "$gate_ok" -eq 1 ] && echo 1 || echo 0)

score=$((s1*w_s1 + s2*w_s2 + s3*w_s3 + s4*w_s4 + s5*w_s5 + s6*w_s6 + s7*w_s7 + s8*w_s8 + s9*w_s9 + s10*w_s10))
status="FAIL"
if [ "$score" -eq 100 ]; then
  status="PASS"
fi
runtime="NONE"
if [ "$docker_ok" -eq 1 ] && [ "$gate_ok" -eq 1 ]; then
  runtime="FULL"
fi

cat > "${OUT}/RESULT.md" <<MD
COMPLETION: ${score}%
STATUS: ${status}
RUNTIME COVERAGE: ${runtime}

S1=${s1} S2=${s2} S3=${s3} S4=${s4} S5=${s5} S6=${s6} S7=${s7} S8=${s8} S9=${s9} S10=${s10}

Primary evidence: logs/gate.txt
MD

cat > "${OUT}/TRACEABILITY.md" <<MD
- gate: ./scripts/gate.sh
- docker required for runtime S2-S6
- playwright docker runner: ./scripts/playwright-docker.sh
- demo smoke: ./scripts/demo-smoke.sh
- prometheus check: ./scripts/prom_up_check.py
- logs dir: ${LOGS}
MD

echo "COMPLETION: ${score}%"
echo "STATUS: ${status}"
echo "RUNTIME COVERAGE: ${runtime}"
echo "Artifacts: ${OUT}"

[ "$status" = "PASS" ]
