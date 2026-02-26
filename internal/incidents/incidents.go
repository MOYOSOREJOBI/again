package incidents

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

const DefaultClusterWindow = 60 * time.Second

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
		recommended := ComputeRecommendedAction(score, confidence, "stable")
		reason := ComputeRankReason(score, confidence, 0, 0, "stable")
		_, err = pool.Exec(ctx, `UPDATE incidents SET last_activity_at=$2, updated_at=now(), priority_score=GREATEST(priority_score,$3), composite_risk=GREATEST(composite_risk,$3), escalation_probability=GREATEST(escalation_probability,$4), confidence=LEAST(confidence,$5), top_driver_1=$6, top_driver_2=$7 WHERE id=$1`, id, at, score, escalation, confidence, recommended, reason)
		return id, err
	}

	recommended := ComputeRecommendedAction(score, confidence, "stable")
	reason := ComputeRankReason(score, confidence, 0, 0, "stable")
	err = pool.QueryRow(ctx, `
INSERT INTO incidents (
	primary_symbol,status,severity_band,priority_score,composite_risk,escalation_probability,confidence,trust_state,
	feature_snapshot_hash,feature_set_version,model_version,started_at,last_activity_at,created_at,updated_at,top_driver_1,top_driver_2
) VALUES ($1,'open','elevated',$2,$2,$3,$4,'stable',$5,'v1',$6,$7,$7,now(),now(),$8,$9)
RETURNING id
`, symbol, score, escalation, confidence, featureHash, modelVersion, at, recommended, reason).Scan(&id)
	return id, err
}

func CanPromoteToCase(status string) bool {
	return status == "open" || status == "ack"
}

func NextIncidentStatus(current, command string) (string, error) {
	switch command {
	case "ack":
		if current == "open" {
			return "ack", nil
		}
	case "resolve":
		if current == "open" || current == "ack" {
			return "resolved", nil
		}
	case "suppress":
		if current == "open" || current == "ack" {
			return "suppressed", nil
		}
	}
	return "", fmt.Errorf("invalid transition %s -> %s", current, command)
}

func ComputeRecommendedAction(compositeRisk, confidence float64, trustState string) string {
	if trustState != "stable" || confidence < 0.5 {
		return "low_confidence_check_data"
	}
	if compositeRisk >= 0.85 {
		return "review_now"
	}
	if compositeRisk >= 0.65 {
		return "promote_to_case"
	}
	if compositeRisk >= 0.35 {
		return "watch"
	}
	return "watch"
}

func ComputeRankReason(compositeRisk, confidence, ageMinutes, recurrence float64, trustState string) string {
	trustPenalty := 0.0
	if trustState != "stable" {
		trustPenalty = 0.2
	}
	sla := math.Min(0.2, ageMinutes/600)
	recur := math.Min(0.2, recurrence/10)
	rank := math.Max(0, math.Min(1, (0.6*compositeRisk)+(0.2*(1-confidence))+sla+recur-trustPenalty))
	return fmt.Sprintf("fallback_rank=%.3f;risk=%.3f;confidence=%.3f;trust=%s", rank, compositeRisk, confidence, trustState)
}
