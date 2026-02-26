# Sentinel

Sentinel is a streaming risk-operations workspace (ticks -> features -> scores -> incidents -> cases) with deterministic fallback scoring and explicit trust labeling.

## Implemented now (final closure state)

### Fully real in this repo
- Auth + JWT cookie flow with seeded users (`admin`, `analyst`, `viewer`) and `/me` role checks.
- Query/alerts APIs backing command center, queue, incident detail, trust, replay, world map, governance, executive, and case routes.
- Incident -> case promotion, case detail timeline, notes/evidence/disposition mutation paths.
- Deterministic fallback scoring with explicit fallback honesty metadata.

### Deterministic and useful (intentionally simplified)
- Replay is deterministic, metadata-derived reconstruction from persisted incident records.
- Governance/executive surfaces are aggregate-first summaries from seeded/demo data.

### Honest simplifications preserved
- No claim of full tick-perfect historical replay reconstruction.
- Governance is summary/reporting, not destructive governance-control execution.

## Browser validation

```bash
make browser-validate
# strict gate mode (fail if browser tooling unavailable)
REQUIRE_BROWSER=1 make browser-validate
```

What it checks:
- World-map region click/filter propagation into downstream fetches.
- Incident -> case promotion and resulting case-detail usability.
- Role gating (viewer read-only behavior) and backend `403` enforcement for forbidden promote/mutate calls.

Artifacts:
- `docs/screenshots/browser-validation.json`

Exit semantics:
- `PASS` -> exit 0
- `SKIP` -> exit 3 (or exit 1 in strict gate mode when `REQUIRE_BROWSER=1`)
- `FAIL` -> exit 1

## Screenshot generation

```bash
make screenshot-smoke
# strict gate mode (fail if browser tooling unavailable)
REQUIRE_BROWSER=1 make screenshot-smoke
```

Behavior:
- Captures after login using seeded admin credentials.
- Resolves dynamic IDs (incident/case/replay) before route capture.
- Emits PASS/SKIP/FAIL entries and writes `docs/screenshots/manifest.json`.
- Stores screenshots in `docs/screenshots/*.png`.

Routes captured:
- `/command-center`, `/queue`, `/incident/[id]`, `/case`, `/case/[id]`, `/trust`, `/replay/[job]`, `/governance`, `/executive`, `/about`, `/glossary`.

## Local demo flow (reviewer runbook)

1. `make dev-keys`
2. `make demo`
3. Sign in (`admin@sentinel.local` / `Sentinel#123`)
4. Open Command Center (`/command-center`)
5. Inspect Queue (`/queue`)
6. Open an Incident (`/incident/[id]`)
7. Promote to Case
8. Open Case Workspace (`/case`, `/case/[id]`)
9. Open Replay (`/replay/demo`)
10. Open Governance + Executive (`/governance`, `/executive`)
11. Review screenshot/browser artifacts under `docs/screenshots/`

## Release validation gate

```bash
make release-gate
# equivalent strict browser checks:
# REQUIRE_BROWSER=1 make browser-validate && REQUIRE_BROWSER=1 make screenshot-smoke && make verify-screenshots
```

Gate includes:
- Go tests + inference tests + web tests (`make unit`)
- Replay/security/demo smoke checks (including case workflow and query auth/governance RBAC integration checks)
- Strict browser validation
- Strict screenshot generation
- Manifest contract validation (`make verify-screenshots`)

## Role/RBAC notes
- Governance summary endpoint is admin-only (`governance:read`).
- Viewer/analyst can access core read surfaces but are denied governance summary APIs by backend enforcement.

## Environment notes
- Browser artifact generation requires a runnable local stack.
- If local Playwright is unavailable, scripts attempt Docker Playwright runner.
- Docker-gated checks may SKIP only when Docker is genuinely unavailable.

## Remaining intentional limitations
- Replay remains deterministic derived / metadata-first, not full tick-perfect recompute.
- Fallback scoring remains primary path when model artifacts are unavailable.
