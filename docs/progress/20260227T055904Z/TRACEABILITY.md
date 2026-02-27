# Traceability

## Commands executed
- repo state + diff stat: `git branch --show-current`, `git rev-parse HEAD`, `git log --oneline -n 25`, `git diff --stat 25dd2d8..HEAD`
- baseline: `make doctor`, `make lint`, `make test`
- web: `(cd web && npm ci)`, `(cd web && npm run build)`, `(cd web && npm test -- --runInBand)`
- inference: `(cd services/inference && pytest -q)`
- playwright: `(cd web && npx playwright test)`, fallback `./scripts/playwright-docker.sh`

## S-score mapping
- S1: baseline commands pass; `make demo` still unproven in this environment.
- S2-S5: runtime/docker proofs unavailable (no docker binary).
- S6: web build/tests pass; runtime warm-data behavior unproven.
- S7: i18n checks and rtl logic evidence available.
- S8: Redis-backed cache/csrf/ratelimit/login limiter present in code/tests.
- S9: replay uses shared scoring package; full runtime replay audit verify unproven here.
- S10: playwright not runnable (npm 403 + no docker fallback).

## Gate mapping
- G0 PASS: doctor/lint/test passed.
- G1-G6 FAIL: docker unavailable, runtime not executable.
- G7 PASS: i18n checks pass.
- G8 FAIL: Playwright runnable proof missing.
