#!/usr/bin/env bash
set -euo pipefail
if ! command -v go >/dev/null 2>&1; then
  echo "SKIP: go not available"
  exit 0
fi
go test ./internal/logic -run TestFeature -count=1
