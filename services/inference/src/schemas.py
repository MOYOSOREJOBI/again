from dataclasses import dataclass

@dataclass
class ScoreOutput:
    symbol: str
    raw_anomaly_score: float
    normalized_anomaly_score: float
    escalation_probability: float
    confidence: float
    composite_risk: float
    severity: str
    explanation: str
    model_version: str
    feature_snapshot_hash: str
