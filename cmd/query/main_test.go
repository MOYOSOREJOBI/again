package main

import (
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestParseFiltersDefault(t *testing.T) {
	r := httptest.NewRequest("GET", "/queue", nil)
	f := parseFilters(r)
	if f.Window != "24h" {
		t.Fatalf("expected default window")
	}
}

func TestParseFiltersExtendedContract(t *testing.T) {
	r := httptest.NewRequest("GET", "/queue?window=custom&from=2025-01-02T03:04:05Z&to=2025-01-02T04:04:05Z&countryCode=US&region=NA&sector=Technology&industry=Software&venue=XNAS&symbol=AAPL&locale=fr", nil)
	f := parseFilters(r)
	if f.Window != "custom" || f.CountryCode != "US" || f.Region != "NA" || f.Sector != "Technology" || f.Industry != "Software" || f.Venue != "XNAS" || f.Symbol != "AAPL" || f.Locale != "fr" {
		t.Fatalf("unexpected parsed filters: %+v", f)
	}
	if !f.From.Equal(time.Date(2025, 1, 2, 3, 4, 5, 0, time.UTC)) || !f.To.Equal(time.Date(2025, 1, 2, 4, 4, 5, 0, time.UTC)) {
		t.Fatalf("expected parsed custom range")
	}
}

func TestParseFiltersNormalizesAndBackCompat(t *testing.T) {
	r := httptest.NewRequest("GET", "/queue?time_window=bogus&country=GB", nil)
	f := parseFilters(r)
	if f.Window != "24h" {
		t.Fatalf("expected fallback window, got %s", f.Window)
	}
	if f.CountryCode != "GB" {
		t.Fatalf("expected country alias propagation")
	}
}

func TestNormalizeReplayViewMode(t *testing.T) {
	if got := normalizeReplayViewMode("as_scored"); got != "as_scored" {
		t.Fatalf("expected as_scored, got %s", got)
	}
	if got := normalizeReplayViewMode("recomputed"); got != "recomputed" {
		t.Fatalf("expected recomputed, got %s", got)
	}
	if got := normalizeReplayViewMode("invalid"); got != "recomputed" {
		t.Fatalf("expected default recomputed, got %s", got)
	}
}

func TestBuildExecutiveIndustryWhere_CustomFromOnly(t *testing.T) {
	f := filters{Window: "custom", From: time.Date(2025, 2, 1, 12, 0, 0, 0, time.UTC)}
	where, args := buildExecutiveIndustryWhere(f)
	if !strings.Contains(where, "i.last_activity_at >= $1") {
		t.Fatalf("expected from-only clause, got %q", where)
	}
	if len(args) != 1 {
		t.Fatalf("expected 1 arg, got %d", len(args))
	}
}

func TestBuildExecutiveIndustryWhere_CustomSwapsInvertedRange(t *testing.T) {
	f := filters{
		Window: "custom",
		From:   time.Date(2025, 2, 1, 14, 0, 0, 0, time.UTC),
		To:     time.Date(2025, 2, 1, 11, 0, 0, 0, time.UTC),
	}
	_, args := buildExecutiveIndustryWhere(f)
	if len(args) != 2 {
		t.Fatalf("expected 2 args, got %d", len(args))
	}
	from, ok1 := args[0].(time.Time)
	to, ok2 := args[1].(time.Time)
	if !ok1 || !ok2 {
		t.Fatalf("expected time args")
	}
	if from.After(to) {
		t.Fatalf("expected normalized range: from=%s to=%s", from, to)
	}
}
