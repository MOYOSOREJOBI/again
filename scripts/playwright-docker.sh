#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
PW_IMAGE="${PW_IMAGE:-mcr.microsoft.com/playwright:v1.53.0-noble}"

echo "=== Playwright Docker runner ==="
echo "Image: ${PW_IMAGE}"

docker run --rm -t \
  --network host \
  -v "${ROOT}:/repo" \
  -w /repo/web \
  -e CI=true \
  -e PLAYWRIGHT_SKIP_BROWSER_DOWNLOAD=1 \
  "${PW_IMAGE}" bash -lc '
    set -euo pipefail
    node -v
    npm -v
    npm ci
    npx playwright --version
    npm run test:e2e
  '
