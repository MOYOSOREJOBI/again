package replay

import "testing"

func TestResultShape(t *testing.T) {
	r := Result{TickCount: 1, FeatureCount: 1, ScoreCount: 1}
	if r.TickCount != 1 || r.FeatureCount != 1 || r.ScoreCount != 1 {
		t.Fatal("invalid result")
	}
}
