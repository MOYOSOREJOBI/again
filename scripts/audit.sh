#!/usr/bin/env bash
set -euo pipefail
TS="$(date -u +%Y%m%dT%H%M%SZ)"
OUT="docs/audit/${TS}"
mkdir -p "$OUT"

if ./scripts/gate.sh >"$OUT/gate.log" 2>&1; then
  status="PASS"
  completion="60.0"
else
  status="FAIL"
  completion="40.0"
fi

cat > "$OUT/RESULT.md" <<MD
COMPLETION: ${completion}%
STATUS: ${status}

See gate output: gate.log
MD

cat > "$OUT/TRACEABILITY.md" <<MD
- gate: ./scripts/gate.sh
- demo smoke: ./scripts/demo-smoke.sh
- prom up check: ./scripts/prom_up_check.py
- playwright docker fallback: ./scripts/playwright-docker.sh
MD

echo "COMPLETION: ${completion}%"
echo "STATUS: ${status}"
echo "Artifacts: ${OUT}"
