import json, os

def load_model():
    path = os.getenv('MODEL_CONFIG_PATH', '')
    if not path:
        return {'name': 'baseline', 'version': 'baseline-v1'}
    with open(path, 'r', encoding='utf-8') as f:
        m = json.load(f)
    if 'version' not in m:
        m['version'] = 'custom-v1'
    return m
