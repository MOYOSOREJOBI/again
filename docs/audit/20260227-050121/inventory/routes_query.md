# Routes for query
Source: cmd/query/main.go

## Route registrations
75:	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write([]byte("ok")) })
76:	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
79:	r.Get("/readyz", func(w http.ResponseWriter, r *http.Request) {
86:	r.Get("/metrics", metrics.Handler)
88:	r.Get("/debug/seed-status", func(w http.ResponseWriter, r *http.Request) {
116:		pr.Get("/queue", func(w http.ResponseWriter, r *http.Request) {
129:		pr.Get("/command-center", func(w http.ResponseWriter, r *http.Request) {
142:		pr.Get("/trust", func(w http.ResponseWriter, r *http.Request) {
154:		pr.Get("/world-map", func(w http.ResponseWriter, r *http.Request) {
166:		pr.Get("/executive-summary", func(w http.ResponseWriter, r *http.Request) {
228:		pr.Get("/governance/summary", func(w http.ResponseWriter, r *http.Request) {
256:		pr.Get("/replay/{job}", func(w http.ResponseWriter, r *http.Request) {
267:			selectedMode := normalizeReplayViewMode(r.URL.Query().Get("mode"))
279:		pr.Get("/incident/{id}", func(w http.ResponseWriter, r *http.Request) {
293:		pr.Get("/cases", func(w http.ResponseWriter, r *http.Request) {
311:		pr.Get("/case/{id}", func(w http.ResponseWriter, r *http.Request) {
323:		pr.Get("/stream/queue", sseStream(func(ctx context.Context) any {
331:		pr.Get("/stream/command-center", sseStream(func(ctx context.Context) any {
335:		pr.Get("/stream/trust", sseStream(func(ctx context.Context) any {
384:	tw := normalizeWindow(q.Get("window"))
386:		tw = normalizeWindow(q.Get("time_window"))
391:	from, _ := time.Parse(time.RFC3339, q.Get("from"))
392:	to, _ := time.Parse(time.RFC3339, q.Get("to"))
393:	country := q.Get("countryCode")
395:		country = q.Get("country")
397:	return filters{Window: tw, From: from, To: to, CountryCode: country, Region: q.Get("region"), Sector: q.Get("sector"), Industry: q.Get("industry"), Venue: q.Get("venue"), Symbol: q.Get("symbol"), Locale: q.Get("locale")}
