from __future__ import annotations
import hashlib
import pickle
from pathlib import Path


class LoadedModel:
    def __init__(self, model=None, version="fallback_v1", artifact_hash="fallback", degraded=True):
        self.model = model
        self.version = version
        self.artifact_hash = artifact_hash
        self.degraded = degraded


def _sha256_file(path: Path) -> str:
    h = hashlib.sha256()
    with path.open("rb") as f:
        for chunk in iter(lambda: f.read(1024 * 1024), b""):
            h.update(chunk)
    return h.hexdigest()


def load_pickle_model(path: str, fallback_version: str) -> LoadedModel:
    p = Path(path)
    if not path or (not p.exists()) or p.is_dir():
        return LoadedModel(model=None, version=fallback_version, artifact_hash="missing", degraded=True)
    with p.open("rb") as f:
        model = pickle.load(f)
    return LoadedModel(model=model, version=p.stem, artifact_hash=_sha256_file(p), degraded=False)
