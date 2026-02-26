package query

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

func LoadTrust(ctx context.Context, db *pgxpool.Pool) (map[string]any, error) {
	var dq, fallback, total int
	_ = db.QueryRow(ctx, `SELECT count(*) FROM scores WHERE coalesce(priority_score,0)<1`).Scan(&dq)
	_ = db.QueryRow(ctx, `SELECT count(*) FROM incidents`).Scan(&total)
	_ = db.QueryRow(ctx, `SELECT count(*) FROM incidents WHERE model_version LIKE '%fallback%'`).Scan(&fallback)
	cacheHealth := "ok"
	if total == 0 {
		cacheHealth = "warming"
	}
	return map[string]any{
		"dqTotals":            map[string]any{"warnings": dq},
		"modelDegradedCount":  fallback,
		"cacheHealth":         cacheHealth,
		"circuitBreakerCount": 0,
		"trustSummary":        map[string]any{"state": "stable", "fallbackRatio": ratioInt(fallback, total)},
	}, nil
}

func ratioInt(a, b int) float64 {
	if b == 0 {
		return 0
	}
	return float64(a) / float64(b)
}
