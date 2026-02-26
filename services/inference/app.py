import json
import logging
import os
import signal
import threading
import time

from flask import Flask, Response
from kafka import KafkaConsumer, KafkaProducer
from prometheus_client import Counter, Gauge, generate_latest, CONTENT_TYPE_LATEST

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


from src.models import load_model
from src.fallbacks import escalation_probability
from src.scoring import composite, severity, clip
from src.explain import explain
from src.artifacts import snapshot_hash


def build_output(m: dict, model: dict) -> dict:
    payload = m.get('payload', {})
    lr = abs(float(payload.get('log_return_1s', payload.get('log_return', 0.0))))
    symbol = m.get('symbol', '')
    if not symbol:
        raise ValueError('symbol is required')
    raw = min(0.99, lr * 3)
    norm = clip(raw)
    esc = escalation_probability(payload)
    risk = composite(norm, esc, v=min(1.0, abs(float(payload.get('ewma_vol_30', 0.2)))), dq=float(payload.get('missingness_rate', 0.0)))
    return {
        'symbol': symbol,
        'score': risk,
        'raw_anomaly_score': raw,
        'normalized_anomaly_score': norm,
        'escalation_probability': esc,
        'confidence': 0.8 if model.get('name') == 'iforest-fallback' else 0.9,
        'severity': severity(risk),
        'explanation': explain(payload),
        'feature_snapshot_hash': snapshot_hash(payload),
        'model_version': model.get('version', 'baseline-v1'),
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
            c = KafkaConsumer(
                CONSUMER_TOPIC,
                bootstrap_servers=[BROKER],
                value_deserializer=lambda v: json.loads(v.decode()),
                consumer_timeout_ms=5000,
            )
            p = KafkaProducer(
                bootstrap_servers=[BROKER],
                value_serializer=lambda v: json.dumps(v).encode(),
            )
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
                        logger.warning('Skipping non-dict message')
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
        logger.info('Inference consumer stopped')


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
    logger.info('Starting Flask on port %d', port)
    app.run(host='0.0.0.0', port=port, debug=False)
