package incidents

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

const DefaultClusterWindow = 60 * time.Second

// FindOrCreateOpenIncident clusters alerts by symbol and recent activity.
func FindOrCreateOpenIncident(ctx context.Context, pool *pgxpool.Pool, symbol string, score, escalation, confidence float64, featureHash, modelVersion string, at time.Time) (int64, error) {
	var id int64
	err := pool.QueryRow(ctx, `
SELECT id
FROM incidents
WHERE primary_symbol=$1 AND status IN ('open','ack') AND last_activity_at >= $2
ORDER BY id DESC
LIMIT 1
`, symbol, at.Add(-DefaultClusterWindow)).Scan(&id)
	if err == nil {
		_, err = pool.Exec(ctx, `UPDATE incidents SET last_activity_at=$2, updated_at=now(), composite_risk=GREATEST(composite_risk,$3), escalation_probability=GREATEST(escalation_probability,$4), confidence=LEAST(confidence,$5) WHERE id=$1`, id, at, score, escalation, confidence)
		return id, err
	}

	err = pool.QueryRow(ctx, `
INSERT INTO incidents (
	primary_symbol,status,severity_band,priority_score,composite_risk,escalation_probability,confidence,trust_state,
	feature_snapshot_hash,feature_set_version,model_version,started_at,last_activity_at,created_at,updated_at
) VALUES ($1,'open','elevated',$2,$2,$3,$4,'stable',$5,'v1',$6,$7,$7,now(),now())
RETURNING id
`, symbol, score, escalation, confidence, featureHash, modelVersion, at).Scan(&id)
	return id, err
}
