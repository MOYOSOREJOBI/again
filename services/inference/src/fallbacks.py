from .scoring import clip, priority_score


def fallback_anomaly(payload: dict) -> tuple[float, float]:
    lr = abs(float(payload.get('log_return_1s', payload.get('log_return', 0.0))))
    raw = min(0.99, lr * 3.0)
    return raw, clip(raw)


def fallback_escalation(payload: dict) -> float:
    z = abs(float(payload.get('z_return_30s', 0.0)))
    vol = abs(float(payload.get('volume_surprise', payload.get('volume_z', 0.0))))
    return clip((0.55 * min(1.0, z / 4.0)) + (0.45 * min(1.0, vol / 5.0)))


def fallback_recommended_action(priority: float, confidence: float, trust_penalty: float) -> str:
    if trust_penalty > 0 or confidence < 0.55:
        return 'low_confidence_check_data'
    if priority >= 0.8:
        return 'review_now'
    if priority >= 0.6:
        return 'promote_to_case'
    return 'watch'


def fallback_rank_reason(composite: float, priority: float, confidence: float, trust_penalty: float) -> str:
    return f"fallback_rank priority={priority:.3f} composite={composite:.3f} confidence={confidence:.3f} trust_penalty={trust_penalty:.3f}"


def fallback_priority(composite: float, payload: dict, trust_penalty: float) -> float:
    return priority_score(composite, payload.get('incident_pressure', 0.1), payload.get('recurrence', 0.1), trust_penalty)
