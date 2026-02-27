package metrics

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandler(t *testing.T) {
	rr := httptest.NewRecorder()
	Handler(rr, httptest.NewRequest(http.MethodGet, "/metrics", nil))
	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d want %d", rr.Code, http.StatusOK)
	}
	if got := rr.Header().Get("Content-Type"); !strings.Contains(got, "text/plain") {
		t.Fatalf("content-type=%q want text/plain", got)
	}
	body := rr.Body.String()
	if !strings.Contains(body, "sentinel_up") {
		t.Fatalf("body missing sentinel_up metric: %q", body)
	}
}
