package query

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

func LoadTrust(ctx context.Context, db *pgxpool.Pool) (map[string]any, error) {
	var dq int
	_ = db.QueryRow(ctx, `SELECT count(*) FROM scores WHERE coalesce(priority_score,0)<1`).Scan(&dq)
	return map[string]any{"dqTotals": map[string]any{"warnings": dq}, "modelDegradedCount": 0, "cacheHealth": "ok", "circuitBreakerCount": 0}, nil
}
