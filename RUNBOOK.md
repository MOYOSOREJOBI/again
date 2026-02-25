# Sentinel Runbook

## Bring up stack
`docker compose -f deploy/docker/docker-compose.yml up -d --build`

## Verify readiness
- gateway: `/readyz`
- query: `/readyz`
- alerts/features/aggregator/governance: `/readyz`
- inference: `/readyz`

## Audit chain verification
Call `GET /audit/verify` as admin via gateway.

## Failure triage
1. Check `docker compose ps` and healthchecks.
2. Check migration and topics-init jobs exited successfully.
3. Check service logs for dependency failures.
