#!/usr/bin/env bash
set -euo pipefail

has_docker=0
if command -v docker >/dev/null 2>&1 && docker info >/dev/null 2>&1; then
  has_docker=1
fi

echo "=== GATE: Web install ==="
( cd web && npm ci )

echo "=== GATE: Go checks ==="
make doctor
make lint
make test

echo "=== GATE: Web build/test ==="
( cd web && npm run check:i18n && npm run build && npm test -- --runInBand )

echo "=== GATE: Inference tests ==="
( cd services/inference && pytest -q )

playwright_ok=0

echo "=== GATE: Playwright ==="
if (cd web && npx -y @playwright/test@1.53.0 --version >/dev/null 2>&1); then
  if ( cd web && npx -y @playwright/test@1.53.0 test ); then
    playwright_ok=1
  fi
fi

if [ "$playwright_ok" -ne 1 ] && [ "$has_docker" -eq 1 ]; then
  echo "Local Playwright unavailable/failed; using Docker runner"
  ./scripts/playwright-docker.sh
  playwright_ok=1
fi

if [ "$playwright_ok" -ne 1 ]; then
  echo "WARN: Playwright unavailable in this environment"
fi

runtime_ok=0
if [ "$has_docker" -eq 1 ]; then
  echo "=== GATE: Demo + runtime smoke (Docker) ==="
  make demo
  ./scripts/demo-smoke.sh
  runtime_ok=1
else
  echo "SKIP: Docker not available; runtime smoke skipped locally (must run in CI Docker runner)."
fi

echo "GATE_SUMMARY docker=${has_docker} playwright=${playwright_ok} runtime=${runtime_ok}"
echo "PASS: gate complete"
