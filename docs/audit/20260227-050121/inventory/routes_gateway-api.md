# Routes for gateway-api
Source: cmd/gateway-api/main.go

## Route registrations
55:	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write([]byte("ok")) })
56:	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
59:	r.Get("/readyz", func(w http.ResponseWriter, r *http.Request) {
66:	r.Get("/metrics", metrics.Handler)
68:	r.Post("/auth/login", func(w http.ResponseWriter, r *http.Request) {
102:	r.Post("/auth/logout", func(w http.ResponseWriter, r *http.Request) {
113:	r.Get("/me", func(w http.ResponseWriter, r *http.Request) {
135:	if xff := strings.TrimSpace(r.Header.Get("X-Forwarded-For")); xff != "" {
