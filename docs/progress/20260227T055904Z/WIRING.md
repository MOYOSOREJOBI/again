# Wiring integrity

## Services → topics
- simulator produces `raw.ticks`.
- aggregator consumes `raw.ticks`, produces `derived.candles`.
- features consumes `derived.candles`, produces `derived.features` and `dq.metrics`.
- alerts consumes `derived.scores`, produces `alerts.created`.

## Services → tables
- aggregator writes `raw_ticks`, `candles`.
- alerts reads/writes `scores`, `alerts`, `incidents`, `cases`, `case_actions`, `incident_alert_links`.
- query reads `raw_ticks`, `candles`, `features`, `scores`, `alerts`, `incidents`, `cases`, `model_registry`, `replay_jobs`, `replay_runs`, `instrument_metadata`.
- governance writes/reads `model_registry`, `model_deployments`, `replay_jobs`.
- replay writes `candles`, `features`, `scores`, `replay_runs`; reads `raw_ticks`, `scores`, `replay_jobs`.

## Web → endpoints
- gateway-api: `/auth/login`, `/auth/logout`, `/me`.
- query: `/command-center`, `/queue`, `/incident/{id}`, `/trust`, `/replay/{job}`, `/world-map`, `/governance/summary`, `/executive-summary`, `/cases`, `/case/{id}`.
- alerts: `/cases/{id}`, `/cases/{id}/notes`, `/cases/{id}/evidence`, `/cases/{id}/disposition`, `/incidents/{id}/promote-case`.
- next api sse: `/api/sse/alerts`.

## Replay → pipeline packages
- replay runner imports and uses `internal/pipeline/scoring.ScoreAndSeverity`.

## Security state location
- cache: Redis via `internal/rediskv` with sync.Map fallback.
- csrf: Redis via `csrfRedis` with sync.Map fallback.
- rate limit: Redis via `rlRedis` with sync.Map fallback.
- login limiter: Redis via `loginRedis` with sync.Map fallback.
