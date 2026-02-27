# Traceability Matrix (S1–S10)

| Requirement | Evidence Artifact(s) | Code/File Anchor(s) | Test/Command Evidence |
|---|---|---|---|
| S1 `make doctor && make lint && make test && make demo` pass | `make_doctor.txt`, `make_lint.txt`, `make_test.txt`, `docker_up.txt` | `Makefile` targets | `make doctor`, `make lint`, `make test`, `make demo` (runtime blocked for docker/demo) |
| S2 compose healthy | `docker_ps.txt` | `deploy/docker/docker-compose.yml` | `docker compose ... ps` |
| S3 root/health/ready/metrics | `curl_8080.txt`, `curl_8085.txt`, `curl_8083.txt`, `curl_8084.txt`, `curl_8090.txt` | `cmd/*/main.go`, `services/inference/app.py` | runtime curls failed in this environment |
| S4 Prometheus up=1 all jobs | `prom_targets.json`, `prom_up.json`, `prom_up_*.json` | `prometheus/prometheus.yml` | `curl ... /api/v1/query?query=up` |
| S5 Grafana provisioning loaded | `grafana_ds_files.txt`, `grafana_dashprov_files.txt`, `grafana_dash_files.txt`, `grafana_compose_mounts.txt`, `grafana.log` | `grafana/provisioning/*`, `deploy/docker/docker-compose.yml` | grafana logs unavailable without docker |
| S6 premium web UX (globe/charts/realtime) | `web_globe_chart_scan.txt`, `web_symbols.txt`, `web_root.txt`, `api_command-center.txt` | `web/components/WorldRiskMap.tsx`, `web/app/command-center/page.tsx` | no runtime e2e proof |
| S7 17 locales + Arabic RTL | `i18n_file.txt`, `i18n_locale_grep.txt`, `i18n_locales_list.txt`, `rtl_scan.txt`, `i18n_missing_keys.json` | `web/lib/i18n.ts` | locale inventory script in audit |
| S8 Redis-backed cache/RL/CSRF | `internal_cache_redis.go.txt`, `cache_security_scan.txt`, `csrf_mw.txt`, `ratelimit_mw.txt` | `internal/cache/redis.go`, `internal/middleware/csrf.go`, `internal/middleware/ratelimit.go` | source inspection |
| S9 replay parity + diff + audit verify | `replay_scan.txt`, `replay_files.txt`, `replay_first_file.txt`, `replay_equivalence_test.txt` | `internal/replay/*` | source inspection + integration script presence |
| S10 Playwright e2e required flows | `playwright.txt` | `web` tests inventory | `cd web && npx playwright test` failed due npm 403 |

## Must-Pass Gate Mapping
- G0 -> S1 artifacts.
- G1/G2 -> docker artifacts.
- G3 -> Prometheus artifacts.
- G4 -> Grafana artifacts.
- G5 -> web runtime/source artifacts.
- G6 -> i18n/rtl artifacts.
- G7 -> replay artifacts.
- G8 -> cache/security artifacts.
