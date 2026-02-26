# Sentinel

Sentinel is a streaming market-risk workspace: ticks -> candles -> features -> scores -> alerts -> incidents -> cases.

## What changed in this pass
- Added a first-class **Case Workspace** (`/case`) and deeper case timeline/actions in `/case/[id]`.
- Added query-side case APIs (`/cases`, `/case/{id}`) and replay timeline enrichment (`/replay/{job}` now includes deterministic timeline entries).
- Tightened Docker/demo readiness and smoke behavior (`scripts/run-demo.sh`) with explicit wait + PASS/SKIP/FAIL semantics.
- Added screenshot capture helper (`scripts/capture-screenshots.sh`) for Command Center, Queue, Incident, Trust, Replay, Governance, and About.
- Productized Governance/Executive/About/Glossary pages with clearer role behavior and honest scope copy.

## What is currently implemented
- Go microservice pipeline with Kafka/Redpanda, Postgres/Timescale, Redis, Prometheus/Grafana.
- Auth via gateway cookies + JWT claims, RBAC roles (`viewer`, `analyst`, `admin`), append-only audit chain.
- Incident-first alerting and investigation workflow (incidents + cases).
- Deterministic fallback scoring in inference with explicit model-degraded markers.
- Route-based Next.js operator pages (`/command-center`, `/queue`, `/incident/[id]`, `/case`, `/case/[id]`, `/trust`, `/governance`, `/executive`, `/replay/[job]`, `/about`, `/glossary`).
- World-map surface powered by real `/world-map` aggregates.

## Model stack reality (v1)
- Hot path today is deterministic fallback scoring (artifact-aware).
- Output includes anomaly, escalation probability, composite risk, priority score, recommended action, rank reason, explanation payload, and lineage fields.
- If model artifacts are missing, outputs explicitly mark fallback/degraded mode.

## Replay reality (v1.1)
- Replay remains metadata-first, but now includes a deterministic derived timeline lane from stored records.
- UI labels replay limitations explicitly; no fake full historical tick-perfect recompute is claimed.

## Roles
- Viewer: read-only analytical and case surfaces.
- Analyst: triage incidents and mutate case workflow.
- Admin: analyst actions + governance surfaces.

## Run locally
```bash
make doctor
make dev-keys
make demo
```

## Screenshot generation
```bash
# Requires running web app at localhost:3000
./scripts/capture-screenshots.sh
# outputs docs/screenshots/*.png
```

## Browser validation flow
1. `make demo`
2. Sign in as seeded admin (`admin@sentinel.local` / `Sentinel#123`)
3. Validate pages: command-center, queue, incident, case, trust, replay, governance, executive, about.
4. Validate world map click -> region filter reflected in queue/command center.

## Demo flow
1. Command Center global posture.
2. Queue ranked triage list.
3. Incident detail explanation and case promotion.
4. Case Workspace (notes/evidence/status/disposition).
5. Replay deterministic timeline lane.
6. Governance + Executive + About/Glossary trust narrative.

## Test from repo root
```bash
make unit
make replay-test
make security-test   # may SKIP without Docker
make demo-smoke      # may SKIP without Docker
make verify
```

## Current limitations
- Docker-dependent integration checks will SKIP in environments without Docker.
- Replay is still not full tick-perfect recompute.
- Deterministic fallback scoring remains the default path in this build.
