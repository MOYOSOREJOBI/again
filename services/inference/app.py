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
from src.explain import explain_record
from src.fallbacks import fallback_anomaly, fallback_escalation, fallback_rank_reason, fallback_ranking, fallback_recommended_action, fallback_summary
from src.models import load_model_bundle
from src.scoring import clip, composite_risk, confidence_bound
from src.schemas import normalize_output

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
model_bundle: dict | None = None


def load_model() -> dict:
    path = os.getenv('MODEL_CONFIG_PATH', '')
    if path and not Path(path).exists():
        raise FileNotFoundError(path)
    if path:
        with open(path, 'r', encoding='utf-8') as f:
            model = json.load(f)
        model.setdefault('name', 'configured-model')
        model.setdefault('version', 'custom-v1')
        model['degraded_mode'] = False
        return model
    return {'name': 'baseline', 'version': 'baseline-v1', 'degraded_mode': True, 'feature_set_version': 'v1'}


def build_output(msg: dict, models: dict | None = None) -> dict:
    models = models or load_model_bundle()
    payload = msg.get('payload')
    symbol = msg.get('symbol', '')
    if not isinstance(payload, dict) or not symbol:
        raise ValueError('malformed feature payload')

    raw_anomaly, normalized_anomaly = fallback_anomaly(payload)
    dq_penalty = clip(float(payload.get('missingness_rate', 0.0)) + float(payload.get('duplicate_rate', 0.0)) + float(payload.get('out_of_order_rate', 0.0)))
    incident_pressure = clip(float(payload.get('incident_pressure', payload.get('recurrence', 0.0))))
    business_weight = clip(float(payload.get('business_weight', 0.2)))
    volatility_context = clip(abs(float(payload.get('volatility_deviation', payload.get('ewma_vol_30', 0.2)))))

    escalation = fallback_escalation(normalized_anomaly, payload, dq_penalty)
    composite = composite_risk(normalized_anomaly, escalation, volatility_context, incident_pressure, business_weight, dq_penalty)
    fallback_mode = any(models[k]['fallback_mode'] for k in ('anomaly', 'escalation', 'ranking'))
    confidence = confidence_bound(0.92, fallback_mode, dq_penalty)
    priority_score = fallback_ranking(composite, confidence, dq_penalty, payload.get('open_sla_pressure', incident_pressure))
    rank_reason = fallback_rank_reason(composite, escalation, dq_penalty, incident_pressure)
    recommended_action = fallback_recommended_action(priority_score, confidence, dq_penalty)
    expected_severity_band, safety_level = fallback_summary(composite, dq_penalty)
    explain = explain_record(payload, dq_penalty)

    merged_model_name = ','.join([models[k]['model_name'] for k in ('anomaly', 'escalation', 'ranking')])
    merged_model_version = ','.join([models[k]['model_version'] for k in ('anomaly', 'escalation', 'ranking')])
    merged_hash = ','.join([models[k]['artifact_hash'] or '' for k in ('anomaly', 'escalation', 'ranking')]).strip(',')

    out = normalize_output({
        'symbol': symbol,
        'raw_anomaly_score': raw_anomaly,
        'normalized_anomaly_score': normalized_anomaly,
        'escalation_probability': escalation,
        'confidence': confidence,
        'expected_severity_band': expected_severity_band,
        'priority_score': priority_score,
        'rank_reason': rank_reason,
        'recommended_action': recommended_action,
        'composite_risk': composite,
        'safety_level': safety_level,
        'top_drivers': explain['top_drivers'],
        'explanation_text': explain['explanation_text'],
        'caveats': explain['caveats'],
        'feature_snapshot_hash': snapshot_hash(payload),
        'feature_set_version': str(payload.get('feature_set_version', 'v1')),
        'model_name': merged_model_name,
        'model_version': merged_model_version,
        'artifact_hash': merged_hash or None,
        'fallback_mode': fallback_mode,
        'dq_penalty': dq_penalty,
        'incident_pressure': incident_pressure,
        'business_weight': business_weight,
        'ts': time.time(),
    })
    out['score'] = out['composite_risk']
    out['severity'] = out['expected_severity_band']
    out['explanation'] = out['explanation_text']
    return out


@app.get('/healthz')
def healthz():
    return 'ok'


@app.get('/readyz')
def readyz():
    ready = kafka_ready and model_loaded
    details = {
        'ready': ready,
        'kafka_ready': kafka_ready,
        'models_loaded': bool(model_bundle),
        'fallback_mode': True if not model_bundle else any(model_bundle[k]['fallback_mode'] for k in ('anomaly', 'escalation', 'ranking')),
    }
    return (jsonify(details), 200) if ready else (jsonify(details), 503)


@app.get('/metrics')
def metrics():
    return Response(generate_latest(), mimetype=CONTENT_TYPE_LATEST)


def run():
    global kafka_ready, model_bundle, model_loaded
    logger.info('starting inference service')
    model_bundle = load_model_bundle()
    model_loaded = True
    logger.info('model metadata', extra={'models': model_bundle})
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
                    MESSAGES_CONSUMED.inc()
                    out = build_output(msg.value, model_bundle)
                    producer.send(PRODUCER_TOPIC, out)
                    producer.flush()
                    MESSAGES_PRODUCED.inc()
                except Exception as err:
                    PROCESSING_ERRORS.inc()
                    logger.error('scoring failure', extra={'error': str(err)})
            consumer.close()
            producer.close()
        except Exception as err:
            kafka_ready = False
            LOOP_RUNNING.set(0)
            PROCESSING_ERRORS.inc()
            logger.error('consumer failure', extra={'error': str(err)})
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
