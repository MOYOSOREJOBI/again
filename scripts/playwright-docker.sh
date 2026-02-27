#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
PW_IMAGE="${PW_IMAGE:-mcr.microsoft.com/playwright:v1.53.0-noble}"

echo "=== Playwright Docker runner ==="
echo "Image: ${PW_IMAGE}"

docker run --rm -t \
  -v "${ROOT}:/repo" \
  -w /repo/web \
  -e CI=1 \
  -e PLAYWRIGHT_SKIP_BROWSER_DOWNLOAD=1 \
  -e npm_config_registry=https://registry.npmjs.org/ \
  "${PW_IMAGE}" bash -lc '
    set -euo pipefail
    node -v
    npm -v
    npm ci
    npx playwright --version
    npx playwright test
  '

if [ -n "${AUDIT_OUT:-}" ]; then
  mkdir -p "${AUDIT_OUT}/proof"
  date -u +%Y-%m-%dT%H:%M:%SZ > "${AUDIT_OUT}/proof/s10_playwright.ok"
fi
