# Wiring Integrity Matrix

## 1) Services → Kafka topics

| Service | Produces | Consumes | Evidence |
|---|---|---|---|
| simulator | `raw.ticks` | — | `cmd/simulator/main.go` kafka produce call |
| aggregator | `derived.candles` | `raw.ticks` | `cmd/aggregator/main.go` kafka consumer + producer |
| features | `derived.features`, `dq.metrics` | `derived.candles` | `cmd/features/main.go` kafka consumer + producer |
| alerts | `alerts.created` | `derived.scores` | `cmd/alerts/main.go` kafka consumer + producer |

## 2) Services → DB tables

| Service/Package | Reads | Writes | Evidence |
|---|---|---|---|
| aggregator | — | `raw_ticks`, `candles` | SQL `INSERT` in `cmd/aggregator/main.go` |
| alerts | `incidents`, `cases` | `scores`, `alerts`, `incident_alert_links`, `cases`, `case_actions` | SQL in `cmd/alerts/main.go` |
| query | `raw_ticks`, `candles`, `features`, `scores`, `alerts`, `incidents`, `cases`, `model_registry`, `replay_jobs`, `replay_runs`, `instrument_metadata` | — | SQL `SELECT` in `cmd/query/main.go` + `internal/query/*.go` |
| governance | `model_deployments`, `model_registry` | `model_registry`, `model_deployments`, `replay_jobs` | SQL in `cmd/governance/main.go` |
| replay | `raw_ticks`, `scores`, `replay_jobs` | `candles`, `features`, `scores`, `replay_runs`, `replay_jobs` | SQL in `internal/replay/*.go` |

## 3) Web → API endpoints called (and existence)

| Web call | Target | Exists server-side? | Evidence |
|---|---|---|---|
| `/me`, `/auth/login`, `/auth/logout` | gateway-api `:8080` | Yes | `web/lib/api.ts` + `cmd/gateway-api/main.go` |
| `/command-center`, `/queue`, `/incident/{id}`, `/trust`, `/replay/{job}`, `/world-map`, `/governance/summary`, `/executive-summary`, `/cases`, `/case/{id}` | query `:8085` | Yes | `web/lib/api.ts` + `cmd/query/main.go` |
| `/cases/{id}`, `/cases/{id}/notes`, `/cases/{id}/evidence`, `/cases/{id}/disposition`, `/incidents/{id}/promote-case` | alerts `:8083` | Yes | `web/lib/api.ts` + `cmd/alerts/main.go` |
| `/api/sse/alerts` | Next.js API route | Yes | `web/lib/stream.ts`, `web/lib/useSSE.ts`, `web/app/api/sse/alerts/route.ts` |

## 4) Replay parity proof

| Check | Finding | Evidence |
|---|---|---|
| Replay uses shared scoring package | Yes | `internal/replay/runner.go` imports `sentinel/internal/pipeline/scoring` and calls `scoring.ScoreAndSeverity(...)` |
| Replay reuses full live pipeline components | No (partial) | Replay still writes directly to tables in `internal/replay/runner.go` rather than using end-to-end live service modules |

## 5) Security state location (Redis vs in-memory)

| Concern | Redis-backed | In-memory fallback | Finding |
|---|---|---|---|
| Cache (`internal/cache`) | Yes via `internal/rediskv` | Yes via `sync.Map` | Acceptable fallback pattern |
| CSRF tokens (`internal/middleware/csrf.go`) | Yes via `csrfRedis` | Yes via `sync.Map` | Acceptable fallback pattern |
| Generic rate limit (`internal/middleware/ratelimit.go`) | Yes via `rlRedis` | Yes via `sync.Map` | Acceptable fallback pattern |
| Login attempt limiter (`cmd/gateway-api/main.go`) | No | Yes (`sync.Map`) | Gap vs strict “Redis security state real” expectation |
