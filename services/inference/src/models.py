import json
import os
from .artifacts import resolve_artifact


def load_model() -> dict:
    path = os.getenv('MODEL_CONFIG_PATH', '')
    artifact = resolve_artifact(path)
    if path and not artifact['available']:
        raise FileNotFoundError(path)
    if artifact['available']:
        with open(path, 'r', encoding='utf-8') as f:
            model = json.load(f)
        model.setdefault('name', 'configured-model')
        model.setdefault('version', 'custom-v1')
        model['artifact_hash'] = artifact['artifact_hash']
        model['artifact_available'] = True
        model['degraded_mode'] = False
        return model
    return {
        'name': 'baseline',
        'version': 'baseline-v1',
        'feature_set_version': 'v1',
        'training_data_window': 'local-seeded',
        'artifact_hash': '',
        'artifact_available': False,
        'degraded_mode': True,
    }
