import json
import logging
import os
import signal
import sys
import threading
import time
from pathlib import Path

from flask import Flask, Response, jsonify
from kafka import KafkaConsumer, KafkaProducer
from prometheus_client import CONTENT_TYPE_LATEST, Counter, Gauge, generate_latest

SERVICE_DIR = Path(__file__).resolve().parent
if str(SERVICE_DIR) not in sys.path:
    sys.path.insert(0, str(SERVICE_DIR))

from src.artifacts import snapshot_hash
from src.explain import deterministic_explain
from src.fallbacks import safety_level_from_bounds
from src.models import load_pickle_model
from src.scoring import composite_risk, priority_score, recommended_action, score_anomaly, score_escalation

logging.basicConfig(level=logging.INFO, format='%(asctime)s %(levelname)s %(message)s')
logger = logging.getLogger(__name__)
app = Flask(__name__)
BROKER = os.getenv('KAFKA_BROKER', 'redpanda:9092')
CONSUMER_TOPIC = os.getenv('CONSUMER_TOPIC', 'derived.features')
PRODUCER_TOPIC = os.getenv('PRODUCER_TOPIC', 'derived.scores')

MESSAGES_CONSUMED = Counter('inference_messages_consumed_total', 'Total consumed feature messages')
MESSAGES_PRODUCED = Counter('inference_messages_produced_total', 'Total produced score messages')
PROCESSING_ERRORS = Counter('inference_processing_errors_total', 'Total inference processing errors')
LOOP_RUNNING = Gauge('inference_loop_running', 'Inference consumer loop running state (1/0)')

running = True
kafka_ready = False
model_bundle: dict | None = None
model_loaded = False
model_lock = threading.RLock()
last_model_refresh = 0.0


def _active_model_paths() -> tuple[str, str]:
    active_cfg_path = os.getenv("ACTIVE_MODELS_PATH", "")
    if active_cfg_path:
        try:
            with open(active_cfg_path, "r", encoding="utf-8") as f:
                cfg = json.load(f)
            anomaly_path = str(cfg.get("anomaly_model_path", "")).strip()
            escalation_path = str(cfg.get("escalation_model_path", "")).strip()
            if anomaly_path or escalation_path:
                return anomaly_path, escalation_path
        except Exception:
            logger.warning("failed to load active model config; using env model paths", exc_info=True)
    return os.getenv("ANOMALY_MODEL_PATH", ""), os.getenv("ESCALATION_MODEL_PATH", "")


def load_models() -> dict:
    anomaly_path, escalation_path = _active_model_paths()
    anomaly = load_pickle_model(anomaly_path, "anomaly_fallback_v1")
    escalation = load_pickle_model(escalation_path, "escalation_fallback_v1")
    return {"anomaly": anomaly, "escalation": escalation}


def refresh_models_if_due(force: bool = False) -> dict:
    global model_bundle, model_loaded, last_model_refresh
    now = time.time()
    interval_s = float(os.getenv("MODEL_RELOAD_INTERVAL_SECONDS", "30"))
    with model_lock:
        if not force and model_bundle is not None and (now-last_model_refresh) < max(interval_s, 1.0):
            return model_bundle
        model_bundle = load_models()
        model_loaded = model_bundle is not None
        last_model_refresh = now
        return model_bundle


def build_output(msg: dict, models: dict | None = None) -> dict:
    models = models or load_models()
    feats = msg.get('payload', {})
    symbol = msg.get('symbol', '')
    if not symbol or not isinstance(feats, dict):
        raise ValueError('malformed feature payload')
    vec = [feats.get(k, 0.0) for k in sorted(feats.keys())]
    row = dict(feats)
    row["vec"] = vec
    raw_if, anomaly_norm, degraded_if = score_anomaly(models["anomaly"].model, row)
    esc_p, confidence, band, degraded_es = score_escalation(models["escalation"].model, row, anomaly_norm, degraded_if)
    comp = composite_risk(anomaly_norm, esc_p, row)
    pri = priority_score(comp, band, confidence, row)
    dq = float(row.get("dq_penalty", 0.0))
    action = recommended_action(pri, confidence, dq)
    safety = safety_level_from_bounds(pri, dq, confidence)
    explain = deterministic_explain(row)

    fallback_mode = degraded_if or degraded_es
    out = {
        'symbol': symbol,
        'raw_anomaly_score': raw_if,
        'normalized_anomaly_score': anomaly_norm,
        'escalation_probability': esc_p,
        'confidence': confidence,
        'expected_severity_band': band,
        'priority_score': pri,
        'recommended_action': action,
        'composite_risk': comp,
        'safety_level': safety,
        'top_drivers': explain['top_drivers'],
        'explanation_text': explain['plain_language'],
        'feature_snapshot_hash': snapshot_hash(feats),
        'feature_set_version': str(msg.get('feature_set_version', feats.get('feature_set_version', 'v2'))),
        'model_version': f'{models["anomaly"].version},{models["escalation"].version}',
        'artifact_hash': f'{models["anomaly"].artifact_hash},{models["escalation"].artifact_hash}',
        'fallback_mode': fallback_mode,
        'dq_penalty': dq,
        'score': comp,
        'severity': band,
        'explanation': explain['plain_language'],
        'ts': time.time(),
    }
    return out




@app.get('/')
def root():
    return jsonify({'service': 'inference', 'version': 'dev', 'links': ['/healthz', '/readyz', '/metrics', '/docs']})

@app.get('/healthz')
def healthz():
    return 'ok'


@app.get('/readyz')
def readyz():
    bundle = refresh_models_if_due()
    ready = kafka_ready and model_loaded
    fallback_mode = True if not bundle else (bundle['anomaly'].degraded or bundle['escalation'].degraded)
    return (jsonify({'ready': ready, 'fallback_mode': fallback_mode, 'model_version': f"{bundle['anomaly'].version},{bundle['escalation'].version}"}), 200 if ready else 503)


@app.get('/metrics')
def metrics():
    return Response(generate_latest(), mimetype=CONTENT_TYPE_LATEST)


def run():
    global kafka_ready
    refresh_models_if_due(force=True)
    while running:
        try:
            consumer = KafkaConsumer(CONSUMER_TOPIC, bootstrap_servers=[BROKER], value_deserializer=lambda v: json.loads(v.decode()), consumer_timeout_ms=3000)
            producer = KafkaProducer(bootstrap_servers=[BROKER], value_serializer=lambda v: json.dumps(v).encode())
            kafka_ready = True
            LOOP_RUNNING.set(1)
            for msg in consumer:
                if not running:
                    break
                try:
                    current_models = refresh_models_if_due()
                    MESSAGES_CONSUMED.inc()
                    out = build_output(msg.value, current_models)
                    producer.send(PRODUCER_TOPIC, out)
                    producer.flush()
                    MESSAGES_PRODUCED.inc()
                except Exception:
                    PROCESSING_ERRORS.inc()
            consumer.close(); producer.close()
        except Exception:
            kafka_ready = False
            LOOP_RUNNING.set(0)
            PROCESSING_ERRORS.inc()
            time.sleep(2)


def signal_handler(sig, frame):
    del sig, frame
    global running
    running = False


if __name__ == '__main__':
    signal.signal(signal.SIGTERM, signal_handler)
    signal.signal(signal.SIGINT, signal_handler)
    threading.Thread(target=run, daemon=True).start()
    app.run(host='0.0.0.0', port=int(os.getenv('PORT', '8090')), debug=False)
