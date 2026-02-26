#!/usr/bin/env bash
set -euo pipefail
OUT_DIR=${1:-docs/screenshots}
mkdir -p "$OUT_DIR"
MANIFEST="$OUT_DIR/manifest.json"
if ! command -v node >/dev/null 2>&1; then
  echo '[{"route":"all","status":"SKIP","reason":"node unavailable"}]' > "$MANIFEST"
  echo "SKIP: node unavailable"
  exit 0
fi
if node -e "require('playwright')" >/dev/null 2>&1; then
  OUT_DIR="$OUT_DIR" node scripts/capture-screenshots.mjs
  exit $?
fi
if command -v docker >/dev/null 2>&1; then
  docker run --rm --network host -v "$PWD":/work -w /work -e OUT_DIR="$OUT_DIR" mcr.microsoft.com/playwright:v1.55.0-jammy node scripts/capture-screenshots.mjs
  exit $?
fi
echo '[{"route":"all","status":"SKIP","reason":"playwright missing locally and docker unavailable"}]' > "$MANIFEST"
echo "SKIP: playwright missing locally and docker unavailable"
exit 0
