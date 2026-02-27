# Final Cleanup Report

## A) Deleted / archived / kept

### Deleted from tracking
- No source code directories were deleted in this cleanup pass.
- Duplicate `.gitignore` entries were removed to reduce maintenance noise.

### Archived / artifact-only
- `docs/audit/**` and `docs/progress/**` remain artifact-only via `.gitignore`, with canonical READMEs kept as source-of-truth.

### Kept
- Core app and service code under `cmd/`, `internal/`, `services/`, `web/`.
- Canonical audit/progress docs: `docs/audit/README.md`, `docs/progress/README.md`.

## B) Deadcode report summary
- `deadcode` tool execution was attempted but blocked by package proxy access restrictions in this environment.
- No dead-code deletions were made without machine proof.

## C) Web unused dependencies removed
- No dependency removals were committed in this pass because registry access restrictions prevented reliable `npm ci` + dependency-prune verification.
- Lint scope was expanded in `web/package.json` to cover `app`, `components`, `lib`, `e2e`, `scripts`.

## D) Script behavior rules (PASS/FAIL/SKIP)
- `runtime-proof.sh` emits machine-readable status lines and exits `2` when Docker is unavailable.
- `gate.sh` now runs `playwright-docker.sh` for S10 proof instead of re-running runtime-proof.
- SKIP behavior is explicit and produces proof markers (e.g., `SKIP_DOCKER.txt`).

## E) Remaining known gaps
- Full runtime proof requires Docker availability.
- Go deadcode pruning remains pending until module proxy access allows `deadcode` installation.
- Web dependency pruning remains pending until npm registry access allows deterministic reinstall + prune validation.

## F) Next engineering priorities
1. Run deadcode in CI (network-enabled) and prune confirmed unreachable functions.
2. Add `eslint` as a pinned dev dependency and enforce `npm run lint` in CI.
3. Add knip (or madge) report in CI artifact to track unused web files/deps.
4. Add `make cleanup-report` target to regenerate docs/cleanup outputs reproducibly.
5. Add a lightweight policy check that blocks conflict markers and generated artifact commits.
