# Sentinel Codebase Status Summary

## What the codebase is and where it is today
Sentinel is a local, production-style streaming risk platform demo. It simulates market ticks, derives candles/features/scores, generates alerts, and exposes governance and audit capabilities through HTTP APIs and a Next.js dashboard.

Current architecture in practice:
- **Data plane**: simulator → aggregator → features → inference → alerts.
- **Control plane**: gateway-api (auth + audit verify), governance (model deploy + replay), query (read APIs).
- **Storage/infra**: Postgres/Timescale, Redpanda, Redis, optional Memcached, Prometheus, Grafana, and a web UI.

The repository appears actively hardened for demos: deterministic startup script ordering, readiness checks, idempotent inserts, and end-to-end local orchestration with `make demo`.

## What it does
### Core pipeline behavior
1. **Simulator** publishes market ticks to `raw.ticks`.
2. **Aggregator** consumes ticks, writes `raw_ticks` and `candles`, and emits `derived.candles`.
3. **Features** computes simple features from candles, stores `features`, and emits `derived.features` and data-quality metrics.
4. **Inference (Python)** consumes features, calculates a severity score, exposes `/metrics`, and emits `derived.scores`.
5. **Alerts** consumes scores, persists to `scores`, creates alerts for medium/high/critical, publishes `alerts.created`, appends audit records, and streams SSE updates.
6. **Query service** exposes `/alerts` and `/scores` from Postgres for UI consumption.

### Governance/auth/audit behavior
- **Gateway API** handles login/logout, identifies caller (`/me`), and provides admin audit-chain verification.
- **Governance API** supports model deploy and replay-start operations gated by JWT role permissions.
- **Audit log** is append-only with hash-chain verification and mutation-blocking triggers.

## Who it is for
- **Primary audience**: engineers, technical product teams, and architecture reviewers who want a realistic local demo of a risk/alert pipeline with governance controls.
- **Secondary audience**: security/compliance stakeholders who need visible RBAC and tamper-evident audit trails in a demo environment.

It is *not yet* positioned as a turnkey production SaaS; docs frame cloud deployment as follow-on work.

## Why this is unique
- Combines **streaming anomaly flow** and **governance primitives** in one local stack (not just pure analytics).
- Includes **audit hash-chain verification** endpoint and DB-level append-only protection.
- Uses **idempotency keys + unique constraints** across pipeline stages for replay-safe writes.
- Exposes both **operator APIs** and a **dashboard UX** with degraded-state messaging when data consistency breaks.

## What is needed to be fully complete / production-ready
1. **Richer automated validation**
   - Add true end-to-end integration assertions (ticks → persisted scores → alerts), not just infra+migrations.
   - Add contract/schema tests for Kafka payloads and API responses.
2. **Security hardening**
   - Tighten cookie security for non-local deployments (`Secure=true`, CSRF/session considerations).
   - Expand privileged-action audit detail and enforcement tests.
3. **Operational maturity**
   - Add stronger SLO-focused readiness checks (beyond static `ok` where applicable).
   - Expand observability coverage/alerts and define retention policies.
4. **Product depth**
   - Explainability, replay diff UX, suppression workflow controls, and incident timelines are the highest leverage differentiators.

## What is currently wrong or risky
- **Integration test scope is shallow**: current integration script validates infra + migration, but does not enforce end-to-end functional correctness.
- **Readiness depth varies by service**: several services return static `ok` for readiness and do not validate dependencies.
- **Demo-first posture**: cloud deployment/security/ops guidance exists but is intentionally not fully implemented in repo automation.
- **Optional components need intentionality**: Memcached is present as optional-cache profile but not central to documented core flow, so teams should decide to either operationalize or remove it.

## Bottom line
Sentinel is a strong, credible **local demo platform** with real strengths in idempotency, auditability, and multi-service orchestration. It is best viewed as an advanced reference implementation: impressive for demos and architecture validation, but still requiring deeper automated assurance and production hardening to be considered deployment-ready at scale.
