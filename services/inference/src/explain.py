def _driver(name: str, value: float) -> dict:
    return {'driver': name, 'impact': round(value, 4)}


def explain_record(payload: dict, trust_penalty: float, model_drivers: list[dict] | None = None) -> dict:
    caveats = []
    if model_drivers:
        top = model_drivers[:3]
        text = 'Model-backed explanation from top contributing features.'
    else:
        z = abs(float(payload.get('z_return_30s', 0.0)))
        vol = abs(float(payload.get('volume_surprise', 0.0)))
        vol_spike = abs(float(payload.get('volatility_deviation', payload.get('ewma_vol_30', 0.0))))
        top = [_driver('return anomaly', z), _driver('volume surprise', vol), _driver('volatility spike', vol_spike)]
        text = 'Abnormal short-window return with elevated volume surprise'
    if trust_penalty > 0:
        caveats.append('Priority reduced because trust was degraded by missing data')
    return {'top_drivers': top, 'explanation_text': text, 'caveats': caveats}
