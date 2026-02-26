import json
import logging
import os
import signal
import sys
import threading
import time
from pathlib import Path

from flask import Flask, Response
from kafka import KafkaConsumer, KafkaProducer
from prometheus_client import CONTENT_TYPE_LATEST, Counter, Gauge, generate_latest

SERVICE_DIR = Path(__file__).resolve().parent
if str(SERVICE_DIR) not in sys.path:
    sys.path.insert(0, str(SERVICE_DIR))

from src.artifacts import snapshot_hash
from src.explain import fallback_explanation
from src.fallbacks import (
    fallback_anomaly,
    fallback_escalation,
    fallback_priority,
    fallback_rank_reason,
    fallback_recommended_action,
)
from src.models import load_model
from src.scoring import clip, composite_risk, confidence_bound, expected_severity_band

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
model_loaded = False


@app.get('/healthz')
def healthz():
    return 'ok'


@app.get('/readyz')
def readyz():
    if not model_loaded:
        return 'model not loaded', 503
    if not kafka_ready:
        return 'kafka not ready', 503
    return 'ok'


@app.get('/metrics')
def metrics():
    return Response(generate_latest(), mimetype=CONTENT_TYPE_LATEST)


def build_output(m: dict, model: dict | None = None) -> dict:
    model = model or load_model()
    payload = m.get('payload', {})
    symbol = m.get('symbol', '')
    if not symbol:
        raise ValueError('symbol is required')

    raw, norm = fallback_anomaly(payload)
    escalation = fallback_escalation(payload)
    dq = clip(float(payload.get('missingness_rate', 0.0)) + float(payload.get('out_of_order_rate', 0.0)))
    trust_penalty = clip(dq)
    composite = composite_risk(norm, escalation, clip(abs(float(payload.get('ewma_vol_30', 0.2)))), clip(float(payload.get('incident_pressure', 0.1))), clip(float(payload.get('business_weight', 0.1))), trust_penalty)
    conf = confidence_bound(0.9 if not model.get('degraded_mode', True) else 0.75, bool(model.get('degraded_mode', True)))
    priority = fallback_priority(composite, payload, trust_penalty)
    rank_reason = fallback_rank_reason(composite, priority, conf, trust_penalty)
    recommended_action = fallback_recommended_action(priority, conf, trust_penalty)
    explanation_payload = fallback_explanation(payload, trust_penalty)

    return {
        'symbol': symbol,
        'score': composite,
        'raw_anomaly_score': raw,
        'normalized_anomaly_score': norm,
        'escalation_probability': escalation,
        'confidence': conf,
        'expected_severity_band': expected_severity_band(composite, trust_penalty),
        'priority_score': priority,
        'rank_reason': rank_reason,
        'recommended_action': recommended_action,
        'composite_risk': composite,
        'feature_snapshot_hash': snapshot_hash(payload),
        'feature_set_version': model.get('feature_set_version', 'v1'),
        'model_version': model.get('version', 'baseline-v1'),
        'model_name': model.get('name', 'baseline'),
        'deployment_status': 'fallback' if model.get('degraded_mode', True) else 'deployed',
        'model_unavailable': bool(model.get('degraded_mode', True)),
        'artifact_hash': model.get('artifact_hash', ''),
        'explanation': explanation_payload['summary'],
        'explanation_payload': explanation_payload,
        'severity': expected_severity_band(composite, trust_penalty),
        'ts': time.time(),
    }


def run():
    global running, kafka_ready, model_loaded
    logger.info('Starting inference consumer on topic %s', CONSUMER_TOPIC)
    try:
        model = load_model()
        model_loaded = True
    except Exception as e:
        logger.error('Model load failed: %s', e)
        model_loaded = False
        return

    LOOP_RUNNING.set(0)
    while running:
        try:
            c = KafkaConsumer(CONSUMER_TOPIC, bootstrap_servers=[BROKER], value_deserializer=lambda v: json.loads(v.decode()), consumer_timeout_ms=5000)
            p = KafkaProducer(bootstrap_servers=[BROKER], value_serializer=lambda v: json.dumps(v).encode())
            LOOP_RUNNING.set(1)
            kafka_ready = True
            break
        except Exception as e:
            logger.warning('Kafka not ready, retrying in 3s: %s', e)
            time.sleep(3)

    if not running:
        return

    try:
        while running:
            for msg in c:
                if not running:
                    break
                try:
                    MESSAGES_CONSUMED.inc()
                    m = msg.value
                    if not isinstance(m, dict):
                        continue
                    out = build_output(m, model)
                    p.send(PRODUCER_TOPIC, out)
                    p.flush()
                    MESSAGES_PRODUCED.inc()
                except Exception as e:
                    PROCESSING_ERRORS.inc()
                    logger.error('Error processing message: %s', e)
    except Exception as e:
        PROCESSING_ERRORS.inc()
        logger.error('Consumer loop error: %s', e)
    finally:
        LOOP_RUNNING.set(0)
        kafka_ready = False
        model_loaded = False
        try:
            c.close()
        except Exception:
            pass
        try:
            p.close()
        except Exception:
            pass


def signal_handler(sig, frame):
    global running
    logger.info('Received shutdown signal')
    running = False


if __name__ == '__main__':
    signal.signal(signal.SIGTERM, signal_handler)
    signal.signal(signal.SIGINT, signal_handler)
    t = threading.Thread(target=run, daemon=True)
    t.start()
    port = int(os.getenv('PORT', '8090'))
    app.run(host='0.0.0.0', port=port, debug=False)
