package main

import (
	"net/http/httptest"
	"testing"
)

func TestParseFiltersDefault(t *testing.T) {
	r := httptest.NewRequest("GET", "/queue", nil)
	f := parseFilters(r)
	if f.TimeWindow != "24h" {
		t.Fatalf("expected default window")
	}
}
