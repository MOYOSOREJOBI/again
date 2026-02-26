package replay

import "time"

type Job struct {
	ID                string
	Status            string
	RequestedBy       string
	TimeWindowStart   time.Time
	TimeWindowEnd     time.Time
	WatermarkPolicyID string
	AllowedLatenessMS int
	ModelVersion      string
	FeatureSetVersion string
}

type Tick struct {
	EventID    string
	Symbol     string
	Price      float64
	Volume     float64
	EventTime  time.Time
	SequenceID int64
}

type Result struct {
	TickCount    int `json:"tick_count"`
	CandleCount  int `json:"candle_count"`
	FeatureCount int `json:"feature_count"`
	ScoreCount   int `json:"score_count"`
}
	ModelVersion      string
	FeatureSetVersion string
}
