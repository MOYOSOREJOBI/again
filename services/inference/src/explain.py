def fallback_explanation(payload: dict, trust_penalty: float) -> dict:
    keys = ['z_return_30s', 'volume_surprise', 'ewma_vol_30', 'log_return_1s']
    ranked = sorted(((k, float(payload.get(k, 0.0))) for k in keys), key=lambda x: abs(x[1]), reverse=True)
    top = [{'feature': k, 'value': v} for k, v in ranked[:3]]
    caveats = []
    if trust_penalty > 0:
        caveats.append('data quality penalty reduced confidence')
    summary = '; '.join(f"{x['feature']}={x['value']:.3f}" for x in top) if top else 'fallback explanation'
    return {
        'top_features': top,
        'summary': summary,
        'caveats': caveats,
        'trust_penalty_note': 'trust penalty applied' if trust_penalty > 0 else 'none',
        'mode': 'deterministic_fallback',
    }
