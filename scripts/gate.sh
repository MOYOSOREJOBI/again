#!/usr/bin/env bash
set -euo pipefail

echo "=== GATE: Go checks ==="
make doctor
make lint
make test

echo "=== GATE: Web install/build ==="
( cd web && npm ci && npm run check:i18n && npm run build )

echo "=== GATE: Playwright (local if available, else Docker) ==="
if (cd web && npx playwright --version >/dev/null 2>&1); then
  if ! ( cd web && npx playwright test ); then
    echo "Local Playwright execution failed; trying Docker runner."
    ./scripts/playwright-docker.sh
  fi
else
  echo "Local Playwright not available; using Docker runner."
  ./scripts/playwright-docker.sh
fi

echo "=== GATE: Demo smoke (requires Docker) ==="
if command -v docker >/dev/null 2>&1 && docker info >/dev/null 2>&1; then
  ./scripts/demo-smoke.sh
else
  echo "SKIP: Docker not available. Run ./scripts/demo-smoke.sh on a Docker-capable machine."
fi

echo "PASS: gate complete"
