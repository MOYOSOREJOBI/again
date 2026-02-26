def explain(payload: dict) -> str:
    keys = ['z_return_30s', 'volume_surprise', 'ewma_vol_30', 'log_return_1s']
    ranked = sorted(((k, abs(float(payload.get(k, 0.0)))) for k in keys), key=lambda x: x[1], reverse=True)
    top = [f"{k}={v:.3f}" for k, v in ranked[:3]]
    return '; '.join(top) if top else 'baseline fallback explanation'
