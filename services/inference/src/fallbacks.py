def escalation_probability(payload: dict) -> float:
    lr = abs(float(payload.get('log_return_1s', payload.get('log_return', 0.0))))
    z = abs(float(payload.get('z_return_30s', 0.0)))
    vol = abs(float(payload.get('volume_surprise', payload.get('volume_z', 0.0))))
    return max(0.0, min(1.0, 0.5*min(1.0, lr*5) + 0.3*min(1.0, z/4) + 0.2*min(1.0, vol/5)))
