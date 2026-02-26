package main

import "testing"

func TestRatio(t *testing.T) {
	if ratio(1, 4) != 0.25 {
		t.Fatalf("unexpected ratio")
	}
	if ratio(2, 0) != 0 {
		t.Fatalf("ratio division guard failed")
	}
}

func TestReplayLanes(t *testing.T) {
	events := []map[string]any{{"risk": 0.7, "priority": 0.9, "region": "NA"}, {"risk": 0.5, "priority": 0.6, "region": "NA"}, {"risk": 0.4, "priority": 0.5, "region": "EU"}}
	lanes := replayLanes(events)
	risk := lanes["risk"].([]float64)
	if len(risk) != 3 || risk[0] != 0.7 || risk[2] != 0.4 {
		t.Fatalf("risk lane mismatch: %#v", risk)
	}
	byRegion := lanes["region_activity"].(map[string]int)
	if byRegion["NA"] != 2 || byRegion["EU"] != 1 {
		t.Fatalf("region activity mismatch: %#v", byRegion)
	}
}
