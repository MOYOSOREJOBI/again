from dataclasses import dataclass


def clip(v: float) -> float:
    return max(0.0, min(1.0, float(v)))


def composite_risk(a: float, e: float, v: float, p: float, b: float, dq: float) -> float:
    return clip((0.35 * a) + (0.30 * e) + (0.15 * v) + (0.10 * p) + (0.10 * b) - (0.20 * dq))


def expected_severity_band(score: float, trust_penalty: float) -> str:
    s = clip(score - trust_penalty)
    if s >= 0.85:
        return 'critical'
    if s >= 0.65:
        return 'high risk'
    if s >= 0.45:
        return 'elevated'
    return 'stable'


def confidence_bound(c: float, degraded: bool) -> float:
    base = clip(c)
    return clip(base - (0.2 if degraded else 0.0))


def priority_score(composite: float, age_pressure: float, recurrence: float, trust_penalty: float) -> float:
    return clip((0.65 * composite) + (0.2 * clip(age_pressure)) + (0.15 * clip(recurrence)) - trust_penalty)
