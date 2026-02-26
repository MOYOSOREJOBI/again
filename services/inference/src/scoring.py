def clip(v: float) -> float:
    return max(0.0, min(1.0, float(v)))


def composite_risk(a: float, e: float, v: float, p: float, b: float, dq: float) -> float:
    return clip((0.35 * clip(a)) + (0.30 * clip(e)) + (0.15 * clip(v)) + (0.10 * clip(p)) + (0.10 * clip(b)) - (0.20 * clip(dq)))


def expected_severity_band(score: float, trust_penalty: float) -> str:
    s = clip(score - (0.2 * clip(trust_penalty)))
    if s >= 0.88:
        return 'critical'
    if s >= 0.72:
        return 'high risk'
    if s >= 0.48:
        return 'elevated'
    return 'stable'


def safety_level_for_score(score: float, dq_penalty: float) -> str:
    if dq_penalty >= 0.45:
        return 'Data Unreliable'
    if score >= 0.88:
        return 'Critical'
    if score >= 0.72:
        return 'High Risk'
    if score >= 0.48:
        return 'Elevated'
    return 'Stable'


def confidence_bound(base: float, fallback_mode: bool, dq_penalty: float) -> float:
    adjusted = clip(base) - (0.15 if fallback_mode else 0.0) - (0.25 * clip(dq_penalty))
    return clip(adjusted)
