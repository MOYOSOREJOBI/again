# TRACEABILITY (S1-S10)

| Requirement | Evidence | File(s) | Test(s)/Command(s) |
|---|---|---|---|
| S1 toolchain + demo | doctor/lint/test outputs + demo failure | `make_doctor.log`, `make_lint.log`, `make_test.log`, `make_demo.log` | `make doctor`, `make lint`, `make test`, `make demo` |
| S2 compose health | compose commands fail (`docker` missing) | `docker_up.log`, `docker_ps.log` | `docker compose ... up/ps` |
| S3 service root routes | curls captured but services unavailable | `curl_*` files under audit dir | `curl ... / /healthz /readyz /metrics` |
| S4 Prometheus up | API files empty/failed connection | `prom_targets.json`, `prom_up.json` | `curl http://localhost:9090/api/v1/...` |
| S5 Grafana provisioning | provisioning files present; runtime load unproven | `grafana/provisioning_files.txt`, `grafana/dashboard_files.txt`, `grafana.log` | `rg --files grafana/...`, `docker compose logs grafana` |
| S6 premium UI | static SVG map + basic cards, not premium globe/charts proof | `web/components/WorldRiskMap.tsx`, `web/app/command-center/page.tsx` | source scan (`rg`) |
| S7 i18n + RTL | only 4 locales in code; locale file checker finds none | `web/lib/i18n.ts`, `locale_key_check.txt` | `python check_locales.py` |
| S8 Redis security state | cache/rate-limit/csrf rely on sync.Map | `internal/cache/redis.go`, `internal/middleware/ratelimit.go`, `internal/middleware/csrf.go` | source scan (`rg`) |
| S9 replay parity | replay includes bespoke scoring function and persistence | `internal/replay/runner.go` | source scan (`rg`) |
| S10 playwright e2e | no playwright files; command fails | `playwright_files.txt`, `playwright_test.log` | `cd web && npx playwright test` |
