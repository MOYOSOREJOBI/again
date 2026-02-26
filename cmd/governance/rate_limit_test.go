package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRateLimitSubjectKey_PrioritizesAuthorization(t *testing.T) {
	keyFn := rateLimitSubjectKey("gov")
	req := httptest.NewRequest(http.MethodPost, "/models/deploy", nil)
	req.RemoteAddr = "10.0.0.1:1234"
	req.Header.Set("Authorization", "Bearer token123")
	req.AddCookie(&http.Cookie{Name: "sentinel_token", Value: "cookie-token"})

	if got := keyFn(req); got != "gov:Bearer token123" {
		t.Fatalf("unexpected key: %s", got)
	}
}

func TestRateLimitSubjectKey_UsesCookieThenIP(t *testing.T) {
	keyFn := rateLimitSubjectKey("workflow")
	req := httptest.NewRequest(http.MethodPost, "/replay/start", nil)
	req.RemoteAddr = "10.0.0.2:8080"
	req.AddCookie(&http.Cookie{Name: "sentinel_token", Value: "cookie-token"})

	if got := keyFn(req); got != "workflow:cookie:cookie-token" {
		t.Fatalf("unexpected cookie key: %s", got)
	}

	req2 := httptest.NewRequest(http.MethodPost, "/replay/start", nil)
	req2.RemoteAddr = "10.0.0.3:9090"
	if got := keyFn(req2); got != "workflow:ip:10.0.0.3:9090" {
		t.Fatalf("unexpected ip key: %s", got)
	}
}
