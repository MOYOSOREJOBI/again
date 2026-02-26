import hashlib
import json
import os


def snapshot_hash(payload: dict) -> str:
    ordered = {k: payload[k] for k in sorted(payload.keys())}
    return hashlib.sha256(json.dumps(ordered, sort_keys=True).encode()).hexdigest()


def artifact_hash(path: str) -> str:
    h = hashlib.sha256()
    with open(path, 'rb') as f:
        while True:
            block = f.read(65536)
            if not block:
                break
            h.update(block)
    return h.hexdigest()


def resolve_artifact(path: str) -> dict:
    if path and os.path.exists(path):
        return {'path': path, 'artifact_hash': artifact_hash(path), 'available': True}
    return {'path': '', 'artifact_hash': '', 'available': False}
