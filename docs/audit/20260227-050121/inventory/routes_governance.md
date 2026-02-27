# Routes for governance
Source: cmd/governance/main.go

## Route registrations
67:	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("ok")) })
68:	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
71:	r.Get("/readyz", readinessHandler(pool))
72:	r.Get("/active-models", func(w http.ResponseWriter, r *http.Request) {
107:	r.Get("/metrics", metrics.Handler)
108:	r.With(middleware.RateLimit(rateLimitSubjectKey("gov", pub), 10, 1*time.Minute)).Post("/models/deploy", func(w http.ResponseWriter, r *http.Request) {
138:	r.With(middleware.RateLimit(rateLimitSubjectKey("workflow", pub), 30, 1*time.Minute)).Post("/replay/start", func(w http.ResponseWriter, r *http.Request) {
181:	r.Get("/models", func(w http.ResponseWriter, r *http.Request) {
279:		if authz := r.Header.Get("Authorization"); len(authz) > 7 {
