from .scoring import clip, expected_severity_band, safety_level_for_score


def fallback_anomaly(payload: dict) -> tuple[float, float]:
    z = abs(float(payload.get('z_return_30s', payload.get('return_zscore', 0.0))))
    vol = abs(float(payload.get('volume_surprise', payload.get('volume_z', 0.0))))
    vol_dev = abs(float(payload.get('volatility_deviation', payload.get('ewma_vol_30', 0.0))))
    raw = (0.5 * min(6.0, z)) + (0.3 * min(6.0, vol)) + (0.2 * min(6.0, vol_dev))
    return raw, clip(raw / 6.0)


def fallback_escalation(normalized_anomaly: float, payload: dict, trust_penalty: float) -> float:
    pressure = clip(float(payload.get('incident_pressure', payload.get('recurrence', 0.0))))
    return clip((0.65 * normalized_anomaly) + (0.25 * pressure) - (0.20 * trust_penalty))


def fallback_ranking(composite: float, confidence: float, trust_penalty: float, sla_pressure: float) -> float:
    return clip((0.60 * composite) + (0.20 * confidence) + (0.20 * clip(sla_pressure)) - (0.25 * trust_penalty))


def fallback_rank_reason(composite: float, escalation: float, trust_penalty: float, pressure: float) -> str:
    if trust_penalty >= 0.25:
        return 'Low trust reduced priority'
    if composite >= 0.8 and escalation >= 0.7:
        return 'High anomaly + elevated escalation'
    if pressure >= 0.6:
        return 'Recurring incident with open SLA'
    return 'Watchlist signal with moderate risk'


def fallback_recommended_action(priority: float, confidence: float, trust_penalty: float) -> str:
    if trust_penalty >= 0.35:
        return 'check_data_quality'
    if priority >= 0.82 and confidence >= 0.55:
        return 'review_now'
    if priority >= 0.70 and confidence >= 0.6:
        return 'promote_to_case'
    if confidence < 0.45:
        return 'suppressed_by_policy'
    return 'watch'


def fallback_summary(composite: float, dq_penalty: float) -> tuple[str, str]:
    severity = expected_severity_band(composite, dq_penalty)
    return severity, safety_level_for_score(composite, dq_penalty)
