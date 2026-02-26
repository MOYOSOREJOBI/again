# Sentinel

Sentinel is a streaming market-risk workspace: ticks -> candles -> features -> scores -> alerts -> incidents -> cases.

## What is currently implemented
- Go microservice pipeline with Kafka/Redpanda, Postgres/Timescale, Redis/Memcached, Prometheus/Grafana.
- Auth via gateway cookies + JWT claims, RBAC roles (`viewer`, `analyst`, `admin`), append-only audit chain.
- Incident-first alerting (clustered incidents, incident queue, incident detail).
- Case workflow endpoints (create/read/update case, notes, evidence, disposition).
- Deterministic fallback scoring in inference with explicit model-degraded markers.
- Route-based Next.js operator pages (`/command-center`, `/queue`, `/incident/[id]`, `/case/[id]`, `/trust`, `/governance`, `/executive`, `/replay/[job]`).
- World-map surface powered by real `/world-map` aggregates (country/region/sector/industry filters).

## Model stack reality (v1)
- Hot path today is deterministic fallback scoring (artifact-aware).
- Output includes anomaly, escalation probability, composite risk, priority score, recommended action, rank reason, explanation payload, and lineage fields.
- If model artifacts are missing, outputs explicitly mark fallback/degraded mode (`model_unavailable=true`, reduced confidence).

## Replay reality (v1)
- Replay is metadata/status driven (job IDs and status summary), not full lane recomputation yet.
- UI labels replay as partial where appropriate.

## Roles
- Viewer: read-only surfaces.
- Analyst: triage and case mutations.
- Admin: analyst actions plus governance controls.

## Run locally
```bash
make doctor
make dev-keys
make demo
```

## Test from repo root
```bash
make unit
make verify
```

`make unit` runs Go tests, inference Python tests, and web tests from repo root.

## Current limitations
- Replay rendering is metadata-first and partial.
- World map is a performant 2D table-driven interaction surface; 3D globe is not shipped.
- Learned LightGBM paths are scaffolded via metadata contract but fallback scoring remains default in this build.
