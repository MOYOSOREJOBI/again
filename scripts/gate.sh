#!/usr/bin/env bash
set -euo pipefail

echo "=== GATE: Repo checks ==="
make doctor
make lint
make test

echo "=== GATE: Runtime proof (Docker required) ==="
if command -v docker >/dev/null 2>&1 && docker info >/dev/null 2>&1; then
  ./scripts/runtime-proof.sh
else
  echo "FAIL: Docker missing; runtime proof cannot be proven."
  exit 1
fi

echo "=== GATE: Playwright (Docker compose service) ==="
./scripts/playwright-docker.sh

echo "PASS: gate complete"
