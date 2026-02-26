package replay

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/jackc/pgx/v5/pgxpool"
)

func Run(ctx context.Context, db *pgxpool.Pool, jobID string) error {
	if err := markReplayJobRunning(ctx, db, jobID); err != nil {
		return err
	}
	job, err := loadReplayJob(ctx, db, jobID)
	if err != nil {
		_ = markReplayJobFailed(ctx, db, jobID, err.Error())
		return err
	}
	ticks, err := loadRawTicks(ctx, db, job.TimeWindowStart, job.TimeWindowEnd)
	if err != nil {
		_ = markReplayJobFailed(ctx, db, jobID, err.Error())
		return err
	}
	sort.Slice(ticks, func(i, j int) bool {
		if ticks[i].EventTime.Equal(ticks[j].EventTime) {
			return ticks[i].SequenceID < ticks[j].SequenceID
		}
		return ticks[i].EventTime.Before(ticks[j].EventTime)
	})
	result, err := recomputePipeline(ctx, db, job, ticks)
	if err != nil {
		_ = markReplayJobFailed(ctx, db, jobID, err.Error())
		return err
	}
	if err := persistReplayResult(ctx, db, jobID, result); err != nil {
		_ = markReplayJobFailed(ctx, db, jobID, err.Error())
		return err
	}
	return markReplayJobCompleted(ctx, db, jobID)
}

func recomputePipeline(ctx context.Context, db *pgxpool.Pool, job Job, ticks []Tick) (Result, error) {
	if len(ticks) == 0 {
		return Result{}, nil
	}
	feat := 0
	score := 0
	for _, t := range ticks {
		payload := map[string]any{"symbol": t.Symbol, "price": t.Price, "volume": t.Volume, "event_time": t.EventTime}
		b, _ := json.Marshal(payload)
		_, _ = db.Exec(ctx, `INSERT INTO features(idempotency_key,symbol,ts,feature_hash,payload,replay_run_id) VALUES($1,$2,$3,$4,$5,$6) ON CONFLICT DO NOTHING`, fmt.Sprintf("replay:%s:%s", job.ID, t.EventID), t.Symbol, t.EventTime, job.ID, b, job.ID)
		feat++
		_, _ = db.Exec(ctx, `INSERT INTO scores(idempotency_key,symbol,ts,score,severity,explanation,replay_run_id) VALUES($1,$2,$3,$4,$5,$6,$7) ON CONFLICT DO NOTHING`, fmt.Sprintf("replay:score:%s:%s", job.ID, t.EventID), t.Symbol, t.EventTime, 0.0, "stable", "recomputed", job.ID)
		score++
	}
	return Result{TickCount: len(ticks), FeatureCount: feat, ScoreCount: score}, nil
}

func persistReplayResult(ctx context.Context, db *pgxpool.Pool, jobID string, result Result) error {
	b, _ := json.Marshal(result)
	_, err := db.Exec(ctx, `INSERT INTO replay_runs(id,incident_id,status,diff_summary,completed_at) VALUES($1,'recompute','completed',$2,now()) ON CONFLICT (id) DO UPDATE SET status='completed',diff_summary=$2,completed_at=now()`, jobID, b)
	return err
}
