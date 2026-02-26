package replay

import "time"

type Job struct {
	ID                string
	Status            string
	RequestedBy       string
	TimeWindowStart   time.Time
	TimeWindowEnd     time.Time
	ModelVersion      string
	FeatureSetVersion string
}
