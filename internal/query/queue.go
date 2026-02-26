package query

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type QueueFilters struct{ TimeWindow, Country, Region, Industry string }

type QueueRow struct {
	ID                int64   `json:"id"`
	Symbol            string  `json:"symbol"`
	Status            string  `json:"status"`
	SeverityBand      string  `json:"severityBand"`
	PriorityScore     float64 `json:"priorityScore"`
	CompositeRisk     float64 `json:"compositeRisk"`
	Confidence        float64 `json:"confidence"`
	Region            string  `json:"region"`
	CountryCode       string  `json:"countryCode"`
	Industry          string  `json:"industry"`
	RecommendedAction string  `json:"recommendedAction"`
}

func windowSQL(tw string) string {
	switch tw {
	case "1h":
		return "i.last_activity_at > now()-interval '1 hour'"
	case "7d":
		return "i.last_activity_at > now()-interval '7 days'"
	default:
		return "i.last_activity_at > now()-interval '24 hours'"
	}
}

func LoadQueue(ctx context.Context, db *pgxpool.Pool, f QueueFilters) ([]QueueRow, error) {
	args := []any{}
	q := `SELECT i.id,i.primary_symbol,i.status,i.severity_band,i.priority_score,i.composite_risk,i.confidence,coalesce(m.region,''),coalesce(m.country_code,''),coalesce(m.industry,''),coalesce(i.top_driver_1,'watch') FROM incidents i LEFT JOIN instrument_metadata m ON m.instrument_id=i.primary_symbol WHERE ` + windowSQL(f.TimeWindow)
	add := func(col, val string) {
		if val == "" {
			return
		}
		args = append(args, val)
		q += fmt.Sprintf(" AND %s=$%d", col, len(args))
	}
	add("m.region", f.Region)
	add("m.country_code", f.Country)
	add("m.industry", f.Industry)
	q += " ORDER BY i.priority_score DESC,i.last_activity_at DESC LIMIT 250"

	rows, err := db.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []QueueRow{}
	for rows.Next() {
		var r QueueRow
		if rows.Scan(&r.ID, &r.Symbol, &r.Status, &r.SeverityBand, &r.PriorityScore, &r.CompositeRisk, &r.Confidence, &r.Region, &r.CountryCode, &r.Industry, &r.RecommendedAction) == nil {
			out = append(out, r)
		}
	}
	return out, nil
}
