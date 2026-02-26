from dataclasses import asdict, dataclass
from typing import Any


CANONICAL_KEYS = [
    'raw_anomaly_score', 'normalized_anomaly_score', 'escalation_probability', 'confidence',
    'expected_severity_band', 'priority_score', 'rank_reason', 'recommended_action',
    'composite_risk', 'safety_level', 'top_drivers', 'explanation_text',
    'feature_snapshot_hash', 'feature_set_version', 'model_name', 'model_version',
    'artifact_hash', 'fallback_mode', 'dq_penalty', 'incident_pressure', 'business_weight',
]


@dataclass
class ScoreOutput:
    symbol: str
    raw_anomaly_score: float = 0.0
    normalized_anomaly_score: float = 0.0
    escalation_probability: float = 0.0
    confidence: float = 0.0
    expected_severity_band: str = 'stable'
    priority_score: float = 0.0
    rank_reason: str = 'Insufficient signal'
    recommended_action: str = 'watch'
    composite_risk: float = 0.0
    safety_level: str = 'Stable'
    top_drivers: list[dict[str, Any]] | None = None
    explanation_text: str = 'No explanation available'
    feature_snapshot_hash: str = ''
    feature_set_version: str = 'v1'
    model_name: str = 'deterministic_fallback'
    model_version: str = 'fallback-v1'
    artifact_hash: str | None = None
    fallback_mode: bool = True
    dq_penalty: float = 0.0
    incident_pressure: float = 0.0
    business_weight: float = 0.0

    def to_dict(self) -> dict[str, Any]:
        out = asdict(self)
        if out['top_drivers'] is None:
            out['top_drivers'] = []
        return out


def normalize_output(payload: dict[str, Any]) -> dict[str, Any]:
    data = ScoreOutput(symbol=str(payload.get('symbol', ''))).to_dict()
    data.update({k: payload.get(k, data.get(k)) for k in data.keys()})
    if not data['symbol']:
        raise ValueError('symbol is required')
    for key in CANONICAL_KEYS:
        data.setdefault(key, None)
    return data
