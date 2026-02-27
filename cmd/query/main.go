package main

import (
	"context"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"sentinel/internal/auth"
	"sentinel/internal/cache"
	"sentinel/internal/config"
	"sentinel/internal/db"
	"sentinel/internal/healthcheck"
	"sentinel/internal/httpx"
	"sentinel/internal/logging"
	"sentinel/internal/metrics"
	"sentinel/internal/middleware"
	qrm "sentinel/internal/query"
	"sentinel/internal/rbac"
)

const (
	ttlQueue     = 3 * time.Second
	ttlSummary   = 5 * time.Second
	ttlWorldMap  = 10 * time.Second
	ttlExecutive = 10 * time.Second
)

type filters struct {
	Window      string
	From        time.Time
	To          time.Time
	CountryCode string
	Region      string
	Sector      string
	Industry    string
	Venue       string
	Symbol      string
	Locale      string
}

func main() {
	if len(os.Args) > 1 && os.Args[1] == "healthcheck" {
		port := healthcheck.MustPort("PORT", 8085)
		os.Exit(healthcheck.Run(port, "/readyz"))
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	logger := logging.New("query")
	cfg := config.Load("query")
	if err := config.Validate(cfg); err != nil {
		log.Fatalf("query: invalid config: %v", err)
	}
	pool, err := db.Connect(ctx, cfg.PostgresURL)
	if err != nil {
		log.Fatalf("query: failed db: %v", err)
	}
	defer pool.Close()
	pub, err := auth.ReadPublic(cfg.JWTPublicKey)
	if err != nil {
		log.Fatalf("query: failed read key: %v", err)
	}

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.CORS)
	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write([]byte("ok")) })
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		httpx.JSON(w, http.StatusOK, map[string]any{"service": "query", "version": "dev", "links": []string{"/healthz", "/readyz", "/metrics", "/docs"}})
	})
	r.Get("/readyz", func(w http.ResponseWriter, r *http.Request) {
		if err := pool.Ping(r.Context()); err != nil {
			http.Error(w, "not ready", 503)
			return
		}
		_, _ = w.Write([]byte("ok"))
	})
	r.Get("/metrics", metrics.Handler)

	r.Group(func(pr chi.Router) {
		pr.Use(requireRole(pub, "read"))
		pr.Get("/queue", func(w http.ResponseWriter, r *http.Request) {
			f := parseFilters(r)
			claims, _ := authn(r, pub)
			key := cache.QueueKey(claims.Role, f.Region, f.CountryCode, f.Sector, f.Industry, f.Venue, f.Symbol, f.Locale, f.Window)
			out, err := cache.GetOrLoadJSON(r.Context(), key, ttlQueue, func() ([]qrm.QueueRow, error) {
				return qrm.LoadQueue(r.Context(), pool, qrm.QueueFilters{Window: f.Window, From: f.From, To: f.To, CountryCode: f.CountryCode, Region: f.Region, Sector: f.Sector, Industry: f.Industry, Venue: f.Venue, Symbol: f.Symbol, Locale: f.Locale})
			})
			if err != nil {
				http.Error(w, "internal", 500)
				return
			}
			httpx.JSON(w, 200, out)
		})
		pr.Get("/command-center", func(w http.ResponseWriter, r *http.Request) {
			f := parseFilters(r)
			claims, _ := authn(r, pub)
			key := cache.CommandCenterKey(claims.Role, f.Region, f.CountryCode, f.Sector, f.Industry, f.Venue, f.Symbol, f.Locale, f.Window)
			out, err := cache.GetOrLoadJSON(r.Context(), key, ttlSummary, func() (map[string]any, error) {
				return qrm.LoadCommandCenter(r.Context(), pool, qrm.QueueFilters{Window: f.Window, From: f.From, To: f.To, CountryCode: f.CountryCode, Region: f.Region, Sector: f.Sector, Industry: f.Industry, Venue: f.Venue, Symbol: f.Symbol, Locale: f.Locale})
			})
			if err != nil {
				http.Error(w, "internal", 500)
				return
			}
			httpx.JSON(w, 200, out)
		})
		pr.Get("/trust", func(w http.ResponseWriter, r *http.Request) {
			f := parseFilters(r)
			key := cache.TrustKey(f.Region, f.CountryCode, f.Sector, f.Industry, f.Venue, f.Symbol, f.Locale, f.Window)
			out, err := cache.GetOrLoadJSON(r.Context(), key, ttlSummary, func() (map[string]any, error) {
				return qrm.LoadTrust(r.Context(), pool, qrm.QueueFilters{Window: f.Window, From: f.From, To: f.To, CountryCode: f.CountryCode, Region: f.Region, Sector: f.Sector, Industry: f.Industry, Venue: f.Venue, Symbol: f.Symbol, Locale: f.Locale})
			})
			if err != nil {
				http.Error(w, "internal", 500)
				return
			}
			httpx.JSON(w, 200, out)
		})
		pr.Get("/world-map", func(w http.ResponseWriter, r *http.Request) {
			f := parseFilters(r)
			key := cache.WorldMapKey(f.Region, f.CountryCode, f.Sector, f.Industry, f.Venue, f.Symbol, f.Locale, f.Window)
			countries, err := cache.GetOrLoadJSON(r.Context(), key, ttlWorldMap, func() ([]qrm.CountryAgg, error) {
				return qrm.LoadWorldMap(r.Context(), pool, qrm.QueueFilters{Window: f.Window, From: f.From, To: f.To, CountryCode: f.CountryCode, Region: f.Region, Sector: f.Sector, Industry: f.Industry, Venue: f.Venue, Symbol: f.Symbol, Locale: f.Locale})
			})
			if err != nil {
				http.Error(w, "internal", 500)
				return
			}
			httpx.JSON(w, 200, map[string]any{"countries": countries, "timeWindow": f.Window})
		})
		pr.Get("/executive-summary", func(w http.ResponseWriter, r *http.Request) {
			f := parseFilters(r)
			key := cache.ExecutiveKey(f.Region, f.CountryCode, f.Sector, f.Industry, f.Venue, f.Symbol, f.Locale, f.Window)
			out, err := cache.GetOrLoadJSON(r.Context(), key, ttlExecutive, func() (map[string]any, error) {
				topRisks, err := qrm.LoadQueue(r.Context(), pool, qrm.QueueFilters{Window: f.Window, From: f.From, To: f.To, CountryCode: f.CountryCode, Region: f.Region, Sector: f.Sector, Industry: f.Industry, Venue: f.Venue, Symbol: f.Symbol, Locale: f.Locale})
				if err != nil {
					return nil, err
				}
				countryConcentration, err := qrm.LoadWorldMap(r.Context(), pool, qrm.QueueFilters{Window: f.Window, From: f.From, To: f.To, CountryCode: f.CountryCode, Region: f.Region, Sector: f.Sector, Industry: f.Industry, Venue: f.Venue, Symbol: f.Symbol, Locale: f.Locale})
				if err != nil {
					return nil, err
				}
				trust, _ := qrm.LoadTrust(r.Context(), pool, qrm.QueueFilters{Window: f.Window, From: f.From, To: f.To, CountryCode: f.CountryCode, Region: f.Region, Sector: f.Sector, Industry: f.Industry, Venue: f.Venue, Symbol: f.Symbol, Locale: f.Locale})
				top := make([]map[string]any, 0, 5)
				for i, row := range topRisks {
					if i >= 5 {
						break
					}
					top = append(top, map[string]any{"id": row.ID, "symbol": row.Symbol, "priorityScore": row.PriorityScore, "compositeRisk": row.CompositeRisk, "severityBand": row.SeverityBand})
				}
				industryConcentration := []map[string]any{}
				industryWhere, industryArgs := buildExecutiveIndustryWhere(f)
				if f.Window == "custom" && !f.From.IsZero() && !f.To.IsZero() {
					industryArgs = append(industryArgs, f.From, f.To)
					industryWhere = fmt.Sprintf("i.last_activity_at BETWEEN $%d AND $%d", len(industryArgs)-1, len(industryArgs))
				}
				addIndustry := func(col, val string) {
					if val == "" {
						return
					}
					industryArgs = append(industryArgs, val)
					industryWhere += fmt.Sprintf(" AND %s=$%d", col, len(industryArgs))
				}
				addIndustry("m.region", f.Region)
				addIndustry("m.country_code", f.CountryCode)
				addIndustry("m.industry", f.Industry)
				addIndustry("m.sector", f.Sector)
				addIndustry("m.venue", f.Venue)
				addIndustry("i.primary_symbol", f.Symbol)
				ir, err := pool.Query(r.Context(), `SELECT coalesce(m.industry,'Unknown'),count(i.id) FROM incidents i LEFT JOIN instrument_metadata m ON m.instrument_id=i.primary_symbol WHERE `+industryWhere+` GROUP BY 1 ORDER BY 2 DESC LIMIT 8`, industryArgs...)
				if err == nil {
					defer ir.Close()
					for ir.Next() {
						var industry string
						var count int
						if ir.Scan(&industry, &count) == nil {
							industryConcentration = append(industryConcentration, map[string]any{"industry": industry, "incidentCount": count})
						}
					}
				}
				riskMemo := "No material change."
				if len(top) > 0 {
					riskMemo = "Top risk concentration is increasing in the leading instruments for the selected time window."
				}
				return map[string]any{"topRisks": top, "countryConcentration": countryConcentration, "industryConcentration": industryConcentration, "trustSummary": trust["trustSummary"], "riskMemo": riskMemo}, nil
			})
			if err != nil {
				http.Error(w, "internal", 500)
				return
			}
			httpx.JSON(w, 200, out)
		})
		pr.Get("/governance/summary", func(w http.ResponseWriter, r *http.Request) {
			lineage := []map[string]any{}
			rows, err := pool.Query(r.Context(), `SELECT model_name,version,state,created_at FROM model_registry ORDER BY created_at DESC LIMIT 20`)
			if err == nil {
				defer rows.Close()
				for rows.Next() {
					var n, v, s string
					var t any
					if rows.Scan(&n, &v, &s, &t) == nil {
						lineage = append(lineage, map[string]any{"model": n, "version": v, "state": s, "createdAt": t})
					}
				}
			}
			replays := []map[string]any{}
			rp, err := pool.Query(r.Context(), `SELECT id::text,status,requested_at,started_at,completed_at,replay_mode,model_version,feature_set_version,watermark_policy_id,allowed_lateness_ms FROM replay_jobs ORDER BY requested_at DESC LIMIT 25`)
			if err == nil {
				defer rp.Close()
				for rp.Next() {
					var id, st, rm, mv, fv, wp string
					var req, stt, ct any
					var late int
					if rp.Scan(&id, &st, &req, &stt, &ct, &rm, &mv, &fv, &wp, &late) == nil {
						replays = append(replays, map[string]any{"id": id, "status": st, "requestedAt": req, "startedAt": stt, "completedAt": ct, "replayMode": rm, "modelVersion": mv, "featureSetVersion": fv, "watermarkPolicy": wp, "allowedLatenessMs": late})
					}
				}
			}
			httpx.JSON(w, 200, map[string]any{"modelLineage": lineage, "replayJobs": replays, "thresholdChanges": []any{}})
		})
		pr.Get("/replay/{job}", func(w http.ResponseWriter, r *http.Request) {
			id := chi.URLParam(r, "job")
			var status, replayMode, mv, fv, wp string
			var late int
			var s, e, started, completed any
			if err := pool.QueryRow(r.Context(), `SELECT status,replay_mode,time_window_start,time_window_end,started_at,completed_at,model_version,feature_set_version,watermark_policy_id,allowed_lateness_ms FROM replay_jobs WHERE id=$1`, id).Scan(&status, &replayMode, &s, &e, &started, &completed, &mv, &fv, &wp, &late); err != nil {
				http.Error(w, "not found", 404)
				return
			}
			var result any = map[string]any{}
			_ = pool.QueryRow(r.Context(), `SELECT diff_summary FROM replay_runs WHERE id=$1`, id).Scan(&result)
			selectedMode := normalizeReplayViewMode(r.URL.Query().Get("mode"))
			partial := status != "completed"
			resp := map[string]any{"id": id, "status": status, "replayMode": replayMode, "selectedMode": selectedMode, "timeWindowStart": s, "timeWindowEnd": e, "startedAt": started, "completedAt": completed, "modelVersion": mv, "featureSetVersion": fv, "watermarkPolicy": wp, "allowedLatenessMs": late, "result": result, "partial": partial, "limitations": []string{"deterministic replay scoring uses return-based approximation"}}
			if m, ok := result.(map[string]any); ok {
				if selectedMode == "as_scored" {
					resp["selectedStats"] = m["as_scored"]
				} else {
					resp["selectedStats"] = m["recomputed"]
				}
			}
			httpx.JSON(w, 200, resp)
		})
		pr.Get("/incident/{id}", func(w http.ResponseWriter, r *http.Request) {
			id := chi.URLParam(r, "id")
			var out = map[string]any{"id": id}
			var status, sym, sev, mv string
			var p, c float64
			if err := pool.QueryRow(r.Context(), `SELECT status,primary_symbol,severity_band,coalesce(priority_score,0),coalesce(composite_risk,0),coalesce(model_version,'') FROM incidents WHERE id=$1`, id).Scan(&status, &sym, &sev, &p, &c, &mv); err == nil {
				out["status"] = status
				out["symbol"] = sym
				out["severity"] = sev
				out["score_header"] = map[string]any{"priority": p, "composite": c}
				out["model_version"] = mv
			}
			httpx.JSON(w, 200, out)
		})
		pr.Get("/cases", func(w http.ResponseWriter, r *http.Request) {
			rows, err := pool.Query(r.Context(), `SELECT id,incident_id,status,reason,coalesce(owner_name,''),created_at,updated_at FROM cases ORDER BY updated_at DESC LIMIT 200`)
			if err != nil {
				httpx.JSON(w, 200, []any{})
				return
			}
			defer rows.Close()
			out := []map[string]any{}
			for rows.Next() {
				var id, incidentID int64
				var status, reason, owner string
				var createdAt, updatedAt any
				if rows.Scan(&id, &incidentID, &status, &reason, &owner, &createdAt, &updatedAt) == nil {
					out = append(out, map[string]any{"id": id, "incident_id": incidentID, "status": status, "reason": reason, "owner": owner, "created_at": createdAt, "updated_at": updatedAt})
				}
			}
			httpx.JSON(w, 200, out)
		})
		pr.Get("/case/{id}", func(w http.ResponseWriter, r *http.Request) {
			id := chi.URLParam(r, "id")
			var incidentID int64
			var cid int64
			var status, reason, owner string
			var createdAt, updatedAt any
			if err := pool.QueryRow(r.Context(), `SELECT id,incident_id,status,reason,coalesce(owner_name,''),created_at,updated_at FROM cases WHERE id=$1`, id).Scan(&cid, &incidentID, &status, &reason, &owner, &createdAt, &updatedAt); err != nil {
				http.Error(w, "not found", http.StatusNotFound)
				return
			}
			httpx.JSON(w, 200, map[string]any{"id": cid, "status": status, "reason": reason, "owner": owner, "created_at": createdAt, "updated_at": updatedAt, "incident_id": incidentID})
		})
		pr.Get("/stream/queue", sseStream(func(ctx context.Context) any {
			rows, _ := qrm.LoadQueue(ctx, pool, qrm.QueueFilters{Window: "24h"})
			inc := map[string]any{"id": 0}
			if len(rows) > 0 {
				inc = map[string]any{"id": rows[0].ID, "symbol": rows[0].Symbol, "priorityScore": rows[0].PriorityScore, "severityBand": rows[0].SeverityBand, "recommendedAction": rows[0].RecommendedAction}
			}
			return map[string]any{"type": "upsert", "incident": inc}
		}, "queue_patch"))
		pr.Get("/stream/command-center", sseStream(func(ctx context.Context) any {
			cc, _ := qrm.LoadCommandCenter(ctx, pool, qrm.QueueFilters{Window: "24h"})
			return cc
		}, "command_center_patch"))
		pr.Get("/stream/trust", sseStream(func(ctx context.Context) any {
			t, _ := qrm.LoadTrust(ctx, pool, qrm.QueueFilters{Window: "24h"})
			return t
		}, "trust_patch"))
	})

	srv := &http.Server{Addr: cfg.HTTPAddr, Handler: r, ReadTimeout: 15 * time.Second, WriteTimeout: 15 * time.Second, IdleTimeout: 60 * time.Second}
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("query server: %v", err)
		}
	}()
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	cancel()
	_ = srv.Shutdown(context.Background())
	logger.Info("query shutdown", nil)
}

func sseStream(payload func(context.Context) any, event string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "unsupported", 500)
			return
		}
		tk := time.NewTicker(15 * time.Second)
		defer tk.Stop()
		for {
			select {
			case <-r.Context().Done():
				return
			case <-tk.C:
				b, _ := json.Marshal(payload(r.Context()))
				_, _ = w.Write([]byte(": heartbeat\n"))
				_, _ = w.Write([]byte("event: " + event + "\n"))
				_, _ = w.Write([]byte("data: " + string(b) + "\n\n"))
				flusher.Flush()
			}
		}
	}
}

func parseFilters(r *http.Request) filters {
	q := r.URL.Query()
	tw := normalizeWindow(q.Get("window"))
	if tw == "" {
		tw = normalizeWindow(q.Get("time_window"))
	}
	if tw == "" {
		tw = "24h"
	}
	from, _ := time.Parse(time.RFC3339, q.Get("from"))
	to, _ := time.Parse(time.RFC3339, q.Get("to"))
	country := q.Get("countryCode")
	if country == "" {
		country = q.Get("country")
	}
	return filters{Window: tw, From: from, To: to, CountryCode: country, Region: q.Get("region"), Sector: q.Get("sector"), Industry: q.Get("industry"), Venue: q.Get("venue"), Symbol: q.Get("symbol"), Locale: q.Get("locale")}
}

func normalizeReplayViewMode(in string) string {
	switch in {
	case "as_scored", "recomputed":
		return in
	default:
		return "recomputed"
	}
}

func normalizeWindow(in string) string {
	switch in {
	case "1h", "24h", "7d", "custom":
		return in
	default:
		return ""
	}
}

func authn(r *http.Request, pub *rsa.PublicKey) (*auth.Claims, bool) {
	c, err := r.Cookie("sentinel_token")
	if err != nil {
		return nil, false
	}
	claims, err := auth.Parse(c.Value, pub)
	if err != nil {
		return nil, false
	}
	return claims, true
}

func requireRole(pub *rsa.PublicKey, perm string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := authn(r, pub)
			if !ok {
				http.Error(w, "unauthorized", 401)
				return
			}
			if !rbac.Allowed(claims.Role, perm) {
				http.Error(w, "forbidden", 403)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func windowSQLForExecutive(tw string) string {
	switch tw {
	case "1h":
		return "i.last_activity_at > now()-interval '1 hour'"
	case "7d":
		return "i.last_activity_at > now()-interval '7 days'"
	default:
		return "i.last_activity_at > now()-interval '24 hours'"
	}
}

func buildExecutiveIndustryWhere(f filters) (string, []any) {
	args := []any{}
	where := windowSQLForExecutive(f.Window)
	if f.Window == "custom" {
		from, to := f.From, f.To
		if !from.IsZero() && !to.IsZero() && to.Before(from) {
			from, to = to, from
		}
		switch {
		case !from.IsZero() && !to.IsZero():
			args = append(args, from, to)
			where = fmt.Sprintf("i.last_activity_at BETWEEN $%d AND $%d", len(args)-1, len(args))
		case !from.IsZero():
			args = append(args, from)
			where = fmt.Sprintf("i.last_activity_at >= $%d", len(args))
		case !to.IsZero():
			args = append(args, to)
			where = fmt.Sprintf("i.last_activity_at <= $%d", len(args))
		}
	}
	return where, args
}
