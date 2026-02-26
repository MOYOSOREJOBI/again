from math import isfinite

def clip(v: float) -> float:
    return max(0.0, min(1.0, v))

def composite(a: float, e: float, v: float = 0.2, p: float = 0.1, b: float = 0.1, dq: float = 0.0) -> float:
    return clip(0.35*a + 0.30*e + 0.15*v + 0.10*p + 0.10*b - 0.20*dq)

def severity(score: float, trust_penalty: float = 0.0) -> str:
    s = score - trust_penalty
    if s > 0.85: return 'critical'
    if s > 0.65: return 'high'
    if s > 0.45: return 'elevated'
    return 'stable'

def safe(v: float) -> float:
    return v if isfinite(v) else 0.0
