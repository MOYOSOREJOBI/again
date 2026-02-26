package query

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

func LoadTrust(ctx context.Context, db *pgxpool.Pool, f QueueFilters) (map[string]any, error) {
	args := []any{}
	where := windowSQL(f.TimeWindow)
	add := func(col, val string) {
		if val == "" {
			return
		}
		args = append(args, val)
		where += fmt.Sprintf(" AND %s=$%d", col, len(args))
	}
	add("m.region", f.Region)
	add("m.country_code", f.Country)
	add("m.industry", f.Industry)

	var dq, fallback, total int
	_ = db.QueryRow(ctx, `SELECT count(s.id) FROM scores s LEFT JOIN incidents i ON i.primary_symbol=s.symbol LEFT JOIN instrument_metadata m ON m.instrument_id=i.primary_symbol WHERE coalesce(s.priority_score,0)<1 AND `+where, args...).Scan(&dq)
	_ = db.QueryRow(ctx, `SELECT count(i.id) FROM incidents i LEFT JOIN instrument_metadata m ON m.instrument_id=i.primary_symbol WHERE `+where, args...).Scan(&total)
	_ = db.QueryRow(ctx, `SELECT count(i.id) FROM incidents i LEFT JOIN instrument_metadata m ON m.instrument_id=i.primary_symbol WHERE i.model_version LIKE '%fallback%' AND `+where, args...).Scan(&fallback)
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
