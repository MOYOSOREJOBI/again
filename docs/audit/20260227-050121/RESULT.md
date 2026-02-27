PROGRESS: 51%
GRADE: F
STATUS: FAIL
RUNTIME COVERAGE: NONE

## A) Score Breakdown (S1–S10)

| ID | Area | Weight | Result | Points |
|---|---|---:|---|---:|
| S1 | Compose healthchecks + service_healthy | 12 | PARTIAL | 6 |
| S2 | Prometheus metrics scraping + up==1 | 10 | PARTIAL | 5 |
| S3 | Grafana provisioning-as-code | 6 | PARTIAL | 3 |
| S4 | Web operator console (globe + charts + SSE) | 12 | PARTIAL | 6 |
| S5 | i18n (17 locales, no missing keys, Arabic RTL) | 8 | PASS | 8 |
| S6 | Redis-backed cache + CSRF + rate limit | 16 | PARTIAL | 8 |
| S7 | Replay parity with shared pipeline logic | 12 | PARTIAL | 6 |
| S8 | Service health/readiness/root observability wiring | 10 | PARTIAL | 5 |
| S9 | Demo/gate automation | 8 | PARTIAL | 4 |
| S10 | Playwright e2e coverage/runnability | 6 | FAIL | 0 |

**Total progress = 51%.**

## B) Must-Pass Gates (G0–G8)

| Gate | Result | Evidence |
|---|---|---|
| G0 Repo checks | PASS | `make doctor`, `make lint`, `make test` succeeded (doctor warnings only). |
| G1 Demo bringup (`make demo`) | FAIL | Docker absent (`docker: command not found`). |
| G2 Compose health | FAIL | Cannot run `docker compose ps`. |
| G3 Root routes 200 JSON | FAIL | Services not started due Docker absence. |
| G4 Prometheus `up==1` | FAIL | Prometheus not running due Docker absence. |
| G5 Grafana provisioning loaded | FAIL | Grafana not running due Docker absence. |
| G6 Web command center runtime validation | FAIL | No full stack runtime to validate warm-data rendering. |
| G7 i18n validation | PASS | `npm run check:i18n` passed, 17 locales and key parity check in place; Arabic RTL code present. |
| G8 Truthfulness | FAIL | Playwright not runnable (npm 403 + no Docker), and login-attempt state remains in-memory `sync.Map`. |

## C) Repo Scrape Summary

Artifacts generated under `inventory/`:
- `files.txt` full tracked file list
- `go_packages.txt` package inventory
- `go_functions.txt` Go function inventory
- `routes_<service>.md` for each cmd service
- `backend_routes.txt` consolidated route scrape
- `web_routes.md` app/API route inventory
- `web_exports.txt` web export index
- `feature_locations.txt` globe/chart/SSE/i18n location evidence
- `migrations.txt` SQL migration list

Notable findings:
- Service root/health/ready/metrics handlers are wired across cmd services.
- Web contains SSE route and client hooks.
- i18n inventory includes 17 locale files and central locale utility.

## D) Wiring Integrity (services→topics, services→tables, web→endpoints)

See `WIRING.md` for full matrix. Key points:
- Kafka chain exists (`raw.ticks` → `derived.candles` → `derived.features`/`dq.metrics`; alerts consumes `derived.scores`, emits `alerts.created`).
- DB table wiring exists for ingest, scoring, incidents/cases, query, governance, replay.
- Web API calls mostly map to existing gateway/query/alerts routes.
- Replay uses shared scoring package but still performs bespoke stage writes directly in replay runner.

## E) Bug / Risk List (ranked)

1. **Critical:** Runtime observability/health gates unverified because Docker unavailable in audit environment.
   - Root cause: no `docker` binary in PATH.
   - Smallest fix plan: rerun this audit on Docker-capable runner and execute mandatory compose/prom/grafana/http checks.

2. **High:** Playwright e2e not runnable in current environment.
   - Root cause: `npx playwright test` blocked by npm 403; Docker fallback cannot run without Docker.
   - Smallest fix plan: preinstall Playwright dependencies in CI image or allow npm registry access; keep docker fallback for local parity.

3. **Medium:** “Redis security state real” is incomplete for login-attempt limiter.
   - Root cause: `cmd/gateway-api/main.go` still uses `sync.Map` (`loginAttempts`) with no Redis backing.
   - Smallest fix plan: migrate login-attempt state to `internal/rediskv` with local fallback and TTL.

4. **Medium:** Replay parity is only partial.
   - Root cause: replay shares scoring function but not full live pipeline wiring.
   - Smallest fix plan: extract reusable pipeline stage funcs for candle/feature/score writes and invoke from both live and replay paths.

## F) What’s Missing + Next Steps

### To reach ~60%
- Run Docker-backed mandatory gates (G1–G5 evidence collection).
- Restore Playwright runnability in CI/container.

### To reach ~80%
- Move login limiter state to Redis-backed shared store.
- Add/verify 5 Playwright scenarios explicitly mapped to product flows.
- Validate web command center runtime with warm data (screenshots + assertions).

### To reach ~95%
- Complete replay/live parity by reusing identical pipeline stages end-to-end.
- Add runtime checks in gate script for Prometheus `up==1` and Grafana provisioning status.
- Add failing tests for any remaining parity/security edge cases.

## G) Audit Artifacts Created

- `docs/audit/20260227-050121/RESULT.md`
- `docs/audit/20260227-050121/TRACEABILITY.md`
- `docs/audit/20260227-050121/WIRING.md`
- `docs/audit/20260227-050121/inventory/files.txt`
- `docs/audit/20260227-050121/inventory/go_packages.txt`
- `docs/audit/20260227-050121/inventory/go_functions.txt`
- `docs/audit/20260227-050121/inventory/routes_aggregator.md`
- `docs/audit/20260227-050121/inventory/routes_alerts.md`
- `docs/audit/20260227-050121/inventory/routes_features.md`
- `docs/audit/20260227-050121/inventory/routes_gateway-api.md`
- `docs/audit/20260227-050121/inventory/routes_governance.md`
- `docs/audit/20260227-050121/inventory/routes_query.md`
- `docs/audit/20260227-050121/inventory/routes_simulator.md`
- `docs/audit/20260227-050121/inventory/backend_routes.txt`
- `docs/audit/20260227-050121/inventory/web_routes.md`
- `docs/audit/20260227-050121/inventory/web_exports.txt`
- `docs/audit/20260227-050121/inventory/feature_locations.txt`
- `docs/audit/20260227-050121/inventory/migrations.txt`
- `docs/audit/20260227-050121/logs/*.txt`
