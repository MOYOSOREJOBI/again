#!/usr/bin/env bash
set -euo pipefail
if ! command -v node >/dev/null 2>&1; then
  echo "SKIP: node unavailable"
  exit 0
fi
if node -e "require('playwright')" >/dev/null 2>&1; then
  node scripts/browser-validate.mjs
  exit $?
fi
if command -v docker >/dev/null 2>&1; then
  docker run --rm --network host -v "$PWD":/work -w /work mcr.microsoft.com/playwright:v1.55.0-jammy node scripts/browser-validate.mjs
  exit $?
fi
echo "SKIP: playwright missing locally and docker unavailable"
exit 0
