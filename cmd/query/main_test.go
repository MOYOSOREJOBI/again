package main

import (
	"net/http/httptest"
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
