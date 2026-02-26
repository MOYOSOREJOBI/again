import hashlib
import json
import os
from pathlib import Path


def snapshot_hash(payload: dict) -> str:
    ordered = {k: payload[k] for k in sorted(payload.keys())}
    return hashlib.sha256(json.dumps(ordered, sort_keys=True, default=str).encode()).hexdigest()


def artifact_exists(path: str) -> bool:
    return bool(path) and Path(path).exists() and Path(path).is_file()


def artifact_hash(path: str) -> str | None:
    if not artifact_exists(path):
        return None
    h = hashlib.sha256()
    with open(path, 'rb') as f:
        for block in iter(lambda: f.read(65536), b''):
            h.update(block)
    return h.hexdigest()


def artifact_version(path: str, default: str = 'unknown') -> str:
    if not artifact_exists(path):
        return default
    base = os.path.basename(path)
    return base.rsplit('.', 1)[0] or default


def resolve_artifact(path: str, default_version: str) -> dict:
    return {
        'path': path,
        'available': artifact_exists(path),
        'artifact_hash': artifact_hash(path),
        'version': artifact_version(path, default_version),
    }
