# Sentinel Audit Snapshot

This repository currently contains severe merge-corruption and compile-time failures across core backend, replay, and frontend modules.

Key blockers observed:
- Go build/tests fail due syntax corruption in `internal/replay/*` and duplicate/unfinished code in query modules.
- Frontend TypeScript build/tests fail due broken JSX/duplicate fragments in trust/map pages.
- Python tests partially pass but `readyz` tests fail due interface drift (`model_loaded` removed).
- Replay implementation writes placeholder recomputed scores (`0.0`, `stable`) rather than full pipeline parity.

See terminal evidence in this audit run for command outputs and failing suites.
