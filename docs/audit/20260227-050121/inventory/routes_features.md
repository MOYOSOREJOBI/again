# Routes for features
Source: cmd/features/main.go

## Route registrations
101:	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write([]byte("ok")) })
102:	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
106:	mux.HandleFunc("/readyz", func(w http.ResponseWriter, r *http.Request) {
123:	mux.HandleFunc("/metrics", metrics.Handler)
