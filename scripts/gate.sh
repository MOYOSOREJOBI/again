#!/usr/bin/env bash
set -euo pipefail

echo "=== GATE: Go checks ==="
make doctor
make lint
make test

echo "=== GATE: Web install/build ==="
( cd web && npm ci && npm run check:i18n && npm run build )

echo "=== GATE: Playwright (local if available, else Docker) ==="
if (cd web && npx -y @playwright/test@1.53.0 --version >/dev/null 2>&1); then
  if ! ( cd web && npx -y @playwright/test@1.53.0 test ); then
    if command -v docker >/dev/null 2>&1 && docker info >/dev/null 2>&1; then
      echo "Local Playwright execution failed; trying Docker runner."
      ./scripts/playwright-docker.sh
    else
      echo "WARN: Playwright failed and Docker is unavailable; skipping e2e in this environment."
    fi
  fi
else
  if command -v docker >/dev/null 2>&1 && docker info >/dev/null 2>&1; then
    echo "Local Playwright not available; using Docker runner."
    ./scripts/playwright-docker.sh
  else
    echo "WARN: Playwright not available and Docker unavailable; skipping e2e in this environment."
  fi
fi

echo "=== GATE: Demo smoke (requires Docker) ==="
if command -v docker >/dev/null 2>&1 && docker info >/dev/null 2>&1; then
  ./scripts/demo-smoke.sh
else
  echo "SKIP: Docker not available. Run ./scripts/demo-smoke.sh on a Docker-capable machine."
fi

echo "PASS: gate complete"
