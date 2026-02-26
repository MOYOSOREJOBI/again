import os
import pickle
from .artifacts import resolve_artifact


def _load(path_env: str, model_name: str, default_version: str):
    path = os.getenv(path_env, '')
    meta = resolve_artifact(path, default_version)
    if not meta['available']:
        return None, {'model_name': model_name, 'model_version': meta['version'], 'artifact_hash': meta['artifact_hash'], 'fallback_mode': True, 'missing_artifact': path or path_env}
    with open(path, 'rb') as f:
        model = pickle.load(f)
    return model, {'model_name': model_name, 'model_version': meta['version'], 'artifact_hash': meta['artifact_hash'], 'fallback_mode': False, 'missing_artifact': ''}


def load_anomaly_model():
    return _load('ANOMALY_MODEL_PATH', 'anomaly_model', 'anomaly-fallback-v1')


def load_escalation_model():
    return _load('ESCALATION_MODEL_PATH', 'escalation_model', 'escalation-fallback-v1')


def load_ranking_model():
    return _load('RANKING_MODEL_PATH', 'ranking_model', 'ranking-fallback-v1')


def load_model_bundle() -> dict:
    anomaly_model, anomaly_meta = load_anomaly_model()
    escalation_model, escalation_meta = load_escalation_model()
    ranking_model, ranking_meta = load_ranking_model()
    return {
        'anomaly': {'model': anomaly_model, **anomaly_meta},
        'escalation': {'model': escalation_model, **escalation_meta},
        'ranking': {'model': ranking_model, **ranking_meta},
    }
