package query

import (
	"context"

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

func LoadQueue(ctx context.Context, db *pgxpool.Pool, f QueueFilters) ([]QueueRow, error) {
	rows, err := db.Query(ctx, `SELECT i.id,i.primary_symbol,i.status,i.severity_band,i.priority_score,i.composite_risk,i.confidence,coalesce(m.region,''),coalesce(m.country_code,''),coalesce(m.industry,''),coalesce(i.top_driver_1,'watch') FROM incidents i LEFT JOIN instrument_metadata m ON m.instrument_id=i.primary_symbol WHERE i.last_activity_at > now()-interval '24 hour' ORDER BY i.priority_score DESC,i.last_activity_at DESC LIMIT 250`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []QueueRow{}
	for rows.Next() {
		var r QueueRow
		if rows.Scan(&r.ID, &r.Symbol, &r.Status, &r.SeverityBand, &r.PriorityScore, &r.CompositeRisk, &r.Confidence, &r.Region, &r.CountryCode, &r.Industry, &r.RecommendedAction) == nil {
			if f.Region != "" && r.Region != f.Region {
				continue
			}
			if f.Country != "" && r.CountryCode != f.Country {
				continue
			}
			if f.Industry != "" && r.Industry != f.Industry {
				continue
			}
			out = append(out, r)
		}
	}
	return out, nil
}
