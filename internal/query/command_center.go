package query

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

func LoadCommandCenter(ctx context.Context, db *pgxpool.Pool) (map[string]any, error) {
	var open, high int
	_ = db.QueryRow(ctx, `SELECT count(*) FROM incidents WHERE status in ('open','ack')`).Scan(&open)
	_ = db.QueryRow(ctx, `SELECT count(*) FROM incidents WHERE severity_band in ('high','critical') AND status in ('open','ack')`).Scan(&high)
	return map[string]any{"openIncidents": open, "highRiskCount": high, "trust": map[string]any{"state": "stable"}}, nil
}
