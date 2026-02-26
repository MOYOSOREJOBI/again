import hashlib
import json

def snapshot_hash(payload: dict) -> str:
    ordered = {k: payload[k] for k in sorted(payload.keys())}
    return hashlib.sha256(json.dumps(ordered, sort_keys=True).encode()).hexdigest()
