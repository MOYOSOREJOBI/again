# Testing Guide

## Frontend
- `cd web && npm ci`
- `npm run lint`
- `npm run typecheck`
- `npm run test`
- `npm run build`

## Go
- `go test ./...`

## Python inference
- `cd services/inference && python -m unittest discover -s tests -q`

## Integration
- `make integration-suite`

## Full Docker runtime validation
- Run from repo root (`come-main`) so compose paths resolve correctly.
- Use the committed validator script:
  - `bash scripts/final-validation.sh`
- If you are on zsh, do not paste shebang blocks directly in terminal; save to file and run with `bash`.

