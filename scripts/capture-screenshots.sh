#!/usr/bin/env bash
set -euo pipefail
OUT_DIR=${1:-docs/screenshots}
BASE=${BASE_URL:-http://localhost:3000}
mkdir -p "$OUT_DIR"
if ! python3 - <<'PY' >/dev/null 2>&1
import playwright
PY
then
  echo "SKIP: playwright python package not available"
  exit 0
fi
python3 - <<PY
from playwright.sync_api import sync_playwright
base = '${BASE}'
out = '${OUT_DIR}'
routes = [
 ('command-center','/command-center'),
 ('queue','/queue'),
 ('incident','/incident/1'),
 ('trust','/trust'),
 ('replay','/replay/demo'),
 ('governance','/governance'),
 ('about','/about'),
]
with sync_playwright() as p:
    browser = p.chromium.launch()
    page = browser.new_page(viewport={'width': 1600, 'height': 1000})
    for name, path in routes:
        try:
            page.goto(base + path, wait_until='networkidle', timeout=20000)
            page.screenshot(path=f"{out}/{name}.png", full_page=True)
            print(f"PASS: {name} -> {out}/{name}.png")
        except Exception as e:
            print(f"FAIL: {name} -> {e}")
    browser.close()
PY
