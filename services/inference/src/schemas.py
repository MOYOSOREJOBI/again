from dataclasses import dataclass, field


@dataclass
class ScoreOutput:
    symbol: str
    raw_anomaly_score: float
    normalized_anomaly_score: float
    escalation_probability: float
    confidence: float
    expected_severity_band: str
    priority_score: float
    rank_reason: str
    recommended_action: str
    composite_risk: float
    feature_snapshot_hash: str
    model_version: str
    explanation: str
    explanation_payload: dict = field(default_factory=dict)
    model_unavailable: bool = False
    deployment_status: str = 'fallback'
