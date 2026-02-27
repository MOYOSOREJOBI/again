package metrics

import (
	"fmt"
	"net/http"
)

func Handler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; version=0.0.4")
	_, _ = fmt.Fprintln(w, "# HELP sentinel_up service process is up")
	_, _ = fmt.Fprintln(w, "# TYPE sentinel_up gauge")
	_, _ = fmt.Fprintln(w, "sentinel_up 1")
}
