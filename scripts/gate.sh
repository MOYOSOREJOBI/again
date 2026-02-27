#!/usr/bin/env bash
set -euo pipefail

echo "=== GATE: Repo checks ==="
make doctor
make lint
make test

echo "=== GATE: Web build/tests ==="
( cd web && npm ci && npm run build && npm test -- --runInBand )

echo "=== GATE: Inference tests ==="
( cd services/inference && pytest -q )

echo "=== GATE: Playwright (Docker is canonical) ==="
if command -v docker >/dev/null 2>&1 && docker info >/dev/null 2>&1; then
  ./scripts/playwright-docker.sh
else
  echo "FAIL: Docker missing; Playwright cannot be proven. Install Docker or rely on CI."
  exit 1
fi

echo "=== GATE: Runtime proof (Docker required) ==="
./scripts/runtime-proof.sh

echo "PASS: gate complete"
