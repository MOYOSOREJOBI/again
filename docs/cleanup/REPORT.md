# Cleanup Report

## A) Deleted, archived, kept

### Deleted from git tracking (generated noise)
- `docs/audit/<timestamp>/**`
- `docs/progress/<timestamp>/**`

Rationale: generated snapshots and logs belong in CI artifacts, not source control.

### Kept in git (source-of-truth docs)
- `docs/audit/README.md`
- `docs/progress/README.md`
- existing core docs at repo root and `docs/` (architecture/runbook/API/etc.)

### Ignore policy added
- `docs/audit/**`
- `docs/progress/**`
- exceptions for canonical READMEs.

## B) Deadcode report summary
- Attempted deadcode tooling install and `go mod tidy`.
- Both blocked by network/proxy restrictions in this environment.
- No code removed without proof.

Details: `docs/cleanup/DEADCODE.md`.

## C) Web unused deps removed list
- None removed in this pass due environment restrictions and lack of reliable dependency-audit tool execution.
- `npm ci` still fails with registry/proxy 403 in this environment.

## D) Script behavior rules now
- `runtime-proof.sh` emits machine-readable lines:
  - `STATUS=PASS|FAIL|SKIP`
  - `REASON=...`
  - `PROOFS_DIR=...`
  - `LOGS_DIR=...`
- Docker missing => explicit SKIP marker file (`SKIP_DOCKER.txt`) and exit code 2 (not fake pass).
- `gate.sh` propagates PASS/FAIL/SKIP consistently.
- `audit.sh` records PASS/FAIL/SKIP and reason in `RESULT.md` plus machine-readable status lines.

## E) Remaining known gaps
1. Runtime completeness still requires Docker-capable runner.
2. `npm ci` blocked by registry proxy policy in current environment.
3. Deadcode and dependency-prune automation blocked by Go proxy restrictions.

## F) Next priorities
1. Run cleanup validation in Docker-capable CI runner and capture runtime markers.
2. Fix network policy for npm/go proxy access in CI.
3. Re-run deadcode pass and remove proven unreachable functions.
4. Add non-interactive web lint strategy (ESLint config or documented TypeScript-only lint policy).
5. Add CI check to reject tracked timestamped audit/progress artifacts.
