# Traceability Matrix (S1–S10, G0–G8)

## Success Bars

| ID | Weight | Score | Evidence |
|---|---:|---|---|
| S1 Compose healthchecks + `service_healthy` correctness | 12 | PARTIAL | `deploy/docker/docker-compose.yml` includes healthchecks and `depends_on: condition: service_healthy`; runtime not validated because Docker unavailable. |
| S2 Prometheus `/metrics` scraping correctness + `up==1` | 10 | PARTIAL | `/metrics` routes wired in services and scrape config in `prometheus/prometheus.yml`; runtime query not executable (no Docker). |
| S3 Grafana provisioning-as-code | 6 | PARTIAL | Provisioning files exist in `grafana/provisioning/**`; runtime load not validated (no Docker). |
| S4 Web console quality (globe + charts + SSE) | 12 | PARTIAL | Components/routes exist and web build/tests pass; no full runtime warmup validation. |
| S5 i18n (17 locales + RTL + missing keys) | 8 | PASS | `npm run check:i18n` passes for 17 locales; RTL logic in `web/lib/i18n.ts`. |
| S6 Real Redis-backed cache + CSRF + rate limiting | 16 | PARTIAL | Redis-backed implementations exist with in-memory fallback; login attempt limiter remains `sync.Map` in gateway. |
| S7 Replay parity via shared pipeline logic | 12 | PARTIAL | Replay uses shared `internal/pipeline/scoring`, but replay still uses bespoke direct DB path for pipeline stages. |
| S8 Health/readiness/root routes observability wiring | 10 | PARTIAL | Source wiring present in cmd services; runtime HTTP checks blocked by absent Docker. |
| S9 Smoke/gate operator automation | 8 | PARTIAL | Scripts exist; cannot execute Docker-dependent demos in this environment. |
| S10 Playwright e2e (5 scenarios, runnable) | 6 | FAIL | `npx playwright test` fails with npm 403; Docker runner fails (`docker: command not found`). |

## Must-Pass Gates

| Gate | Result | Evidence |
|---|---|---|
| G0 Repo checks (`make doctor`, `make lint`, `make test`) | PASS | Commands succeeded; doctor produced warnings only. |
| G1 Demo bringup (`make demo`) | FAIL | Docker unavailable (`docker: command not found`). |
| G2 Compose health (`docker compose ps` healthy) | FAIL | Docker unavailable. |
| G3 Root routes return 200 JSON | FAIL | Runtime service stack not started (Docker unavailable). |
| G4 Prometheus `up==1` via API | FAIL | Runtime Prometheus not available (Docker unavailable). |
| G5 Grafana provisioning loaded | FAIL | Runtime Grafana not available (Docker unavailable). |
| G6 Web command center renders globe/charts/data after warmup | FAIL | No runtime environment with backend + warm data. |
| G7 i18n 17 locales + no missing keys + Arabic RTL | PASS | i18n check pass + RTL code path exists. |
| G8 Truthfulness (Redis real, replay parity, Playwright runnable) | FAIL | Playwright not runnable; login limiter still in-memory. |

## Command Log Index

- `logs/repo_state.txt`
- `logs/make_doctor.txt`
- `logs/make_lint.txt`
- `logs/make_test.txt`
- `logs/web_npm_ci.txt`
- `logs/web_build.txt`
- `logs/web_test.txt`
- `logs/web_i18n_check.txt`
- `logs/inference_pytest.txt`
- `logs/playwright.txt`
- `logs/playwright_docker.txt`
- `logs/docker_available.txt`
