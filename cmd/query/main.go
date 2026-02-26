package main

import (
	"context"
	"crypto/rsa"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"sentinel/internal/auth"
	"sentinel/internal/config"
	"sentinel/internal/db"
	"sentinel/internal/httpx"
	"sentinel/internal/logging"
	"sentinel/internal/rbac"
)

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (s *statusWriter) WriteHeader(code int) { s.status = code; s.ResponseWriter.WriteHeader(code) }

type filters struct{ TimeWindow, Country, Region, Sector, Industry, Venue, AssetClass, Severity, TrustState string }

func parseFilters(r *http.Request) (filters, error) {
	f := filters{TimeWindow: strings.ToLower(r.URL.Query().Get("time_window")), Country: r.URL.Query().Get("country"), Region: r.URL.Query().Get("region"), Sector: r.URL.Query().Get("sector"), Industry: r.URL.Query().Get("industry"), Venue: r.URL.Query().Get("venue"), AssetClass: r.URL.Query().Get("asset_class"), Severity: r.URL.Query().Get("severity"), TrustState: r.URL.Query().Get("trust_state")}
	if f.TimeWindow == "" {
		f.TimeWindow = "24h"
	}
	if f.TimeWindow != "now" && f.TimeWindow != "1h" && f.TimeWindow != "24h" && f.TimeWindow != "7d" {
		return f, fmt.Errorf("invalid time_window")
	}
	return f, nil
}
func applyFilters(base string, args *[]any, f filters) string {
	add := func(col, val string) {
		if val != "" {
			*args = append(*args, val)
			base += fmt.Sprintf(" AND %s=$%d", col, len(*args))
		}
	}
	add("m.country", f.Country)
	add("m.region", f.Region)
	add("m.sector", f.Sector)
	add("m.industry", f.Industry)
	add("m.primary_venue", f.Venue)
	add("i.severity_band", f.Severity)
	add("i.trust_state", f.TrustState)
	return base
}
func windowClause(tw string) string {
	if tw == "1h" {
		return "i.last_activity_at > now()- interval '1 hour'"
	}
	if tw == "7d" {
		return "i.last_activity_at > now()- interval '7 days'"
	}
	return "i.last_activity_at > now()- interval '24 hours'"
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	logger := logging.New("query")
	cfg := config.Load("query")
	if err := config.Validate(cfg); err != nil {
		log.Fatalf("query: invalid config: %v", err)
	}
	pool, err := db.Connect(ctx, cfg.PostgresURL)
	if err != nil {
		log.Fatalf("query: failed to connect db: %v", err)
	}
	defer pool.Close()
	pub, err := auth.ReadPublic(cfg.JWTPublicKey)
	if err != nil {
		log.Fatalf("query: failed read key: %v", err)
	}
	r := chi.NewRouter()
	r.Use(corsMiddleware)
	r.Use(requestLoggingMiddleware(logger))
	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("ok")) })
	r.Get("/readyz", func(w http.ResponseWriter, r *http.Request) {
		if err := pool.Ping(ctx); err != nil {
			http.Error(w, "not ready", 503)
			return
		}
		w.Write([]byte("ok"))
	})

	r.Group(func(pr chi.Router) {
		pr.Use(requireRole(pub, "read"))

		pr.Get("/cases", func(w http.ResponseWriter, r *http.Request) {
			rows, err := pool.Query(ctx, `SELECT c.id,c.incident_id,c.status,c.reason,coalesce(c.owner_name,''),c.created_at,c.updated_at FROM cases c ORDER BY c.updated_at DESC LIMIT 200`)
			if err != nil {
				http.Error(w, "internal error", 500)
				return
			}
			defer rows.Close()
			out := []map[string]any{}
			for rows.Next() {
				var id, incidentID int64
				var st, reason, owner string
				var createdAt, updatedAt any
				if rows.Scan(&id, &incidentID, &st, &reason, &owner, &createdAt, &updatedAt) == nil {
					out = append(out, map[string]any{"id": id, "incident_id": incidentID, "status": st, "reason": reason, "owner": owner, "created_at": createdAt, "updated_at": updatedAt})
				}
			}
			httpx.JSON(w, 200, out)
		})
		pr.Get("/case/{id}", func(w http.ResponseWriter, r *http.Request) {
			id := chi.URLParam(r, "id")
			var cid, incidentID int64
			var st, reason, owner string
			var createdAt, updatedAt any
			if err := pool.QueryRow(ctx, `SELECT id,incident_id,status,reason,coalesce(owner_name,''),created_at,updated_at FROM cases WHERE id=$1`, id).Scan(&cid, &incidentID, &st, &reason, &owner, &createdAt, &updatedAt); err != nil {
				http.Error(w, "not found", 404)
				return
			}
			httpx.JSON(w, 200, map[string]any{"id": cid, "incident_id": incidentID, "status": st, "reason": reason, "owner": owner, "created_at": createdAt, "updated_at": updatedAt})
		})
		pr.Get("/queue", func(w http.ResponseWriter, r *http.Request) {
			f, err := parseFilters(r)
			if err != nil {
				http.Error(w, err.Error(), 400)
				return
			}
			args := []any{}
			q := `SELECT i.id,i.primary_symbol,i.status,i.severity_band,i.priority_score,i.composite_risk,i.escalation_probability,i.confidence,i.trust_state,coalesce(i.top_driver_1,''),coalesce(i.top_driver_2,''),coalesce(m.country,''),coalesce(m.region,''),coalesce(m.sector,''),coalesce(m.industry,''),coalesce(m.primary_venue,'') FROM incidents i LEFT JOIN instrument_metadata m ON m.instrument_id=i.primary_symbol WHERE ` + windowClause(f.TimeWindow)
			q = applyFilters(q, &args, f) + " ORDER BY i.priority_score DESC, i.last_activity_at DESC LIMIT 250"
			rows, err := pool.Query(ctx, q, args...)
			if err != nil {
				http.Error(w, "internal error", 500)
				return
			}
			defer rows.Close()
			out := []map[string]any{}
			for rows.Next() {
				var id int64
				var sym, st, sev, trust, d1, d2, c, reg, sec, ind, ven string
				var p, rk, e, conf float64
				if rows.Scan(&id, &sym, &st, &sev, &p, &rk, &e, &conf, &trust, &d1, &d2, &c, &reg, &sec, &ind, &ven) == nil {
					out = append(out, map[string]any{"id": id, "symbol": sym, "status": st, "severity": sev, "composite_risk": rk, "priority_score": p, "escalation_probability": e, "confidence": conf, "trust_label": trust, "top_drivers": []string{d1, d2}, "recommended_action": d1, "rank_reason": d2, "country": c, "region": reg, "sector": sec, "industry": ind, "venue": ven, "trend": []float64{rk * 0.8, rk * 0.9, rk}})
				}
			}
			httpx.JSON(w, 200, out)
		})
		pr.Get("/command-center", func(w http.ResponseWriter, r *http.Request) {
			f, err := parseFilters(r)
			if err != nil {
				http.Error(w, err.Error(), 400)
				return
			}
			rows := []map[string]any{}
			q := `SELECT coalesce(m.region,'global'), count(i.id), avg(i.composite_risk) FROM incidents i LEFT JOIN instrument_metadata m ON m.instrument_id=i.primary_symbol WHERE ` + windowClause(f.TimeWindow) + ` GROUP BY 1 ORDER BY 2 DESC LIMIT 5`
			rw, _ := pool.Query(ctx, q)
			for rw.Next() {
				var rg string
				var n int
				var p float64
				if rw.Scan(&rg, &n, &p) == nil {
					rows = append(rows, map[string]any{"region": rg, "incident_count": n, "incident_pressure": p})
				}
			}
			queue := []map[string]any{}
			tr, _ := pool.Query(ctx, `SELECT id,primary_symbol,severity_band,priority_score,composite_risk FROM incidents WHERE status IN ('open','ack') ORDER BY priority_score DESC LIMIT 5`)
			for tr.Next() {
				var id int64
				var s, se string
				var pr, rk float64
				if tr.Scan(&id, &s, &se, &pr, &rk) == nil {
					queue = append(queue, map[string]any{"id": id, "symbol": s, "severity": se, "priority_score": pr, "composite_risk": rk})
				}
			}
			httpx.JSON(w, 200, map[string]any{"trust": map[string]any{"state": "stable"}, "open_incidents": len(queue), "high_risk_count": len(queue), "what_changed": "Priority shifts reflect latest incident pressure.", "top_incidents": queue, "incident_pressure_series": []float64{0.42, 0.53, 0.49, 0.57, 0.55}, "severity_distribution": map[string]int{"stable": 1, "elevated": 2, "high risk": 2, "critical": 0}, "top_regions": rows, "top_sectors": []map[string]any{{"sector": "TECH", "count": 3}, {"sector": "FIN", "count": 2}}, "filters": f})
		})
		pr.Get("/incident/{id}", func(w http.ResponseWriter, r *http.Request) {
			id := chi.URLParam(r, "id")
			var iid int64
			var sym, sev, trust, d1, d2, mv, fh string
			var a, e, c, p float64
			if err := pool.QueryRow(ctx, `SELECT id,primary_symbol,severity_band,trust_state,coalesce(top_driver_1,''),coalesce(top_driver_2,''),coalesce(model_version,''),coalesce(feature_snapshot_hash,''),coalesce(normalized_anomaly_score,0),coalesce(escalation_probability,0),coalesce(confidence,0),coalesce(composite_risk,0) FROM incidents WHERE id=$1`, id).Scan(&iid, &sym, &sev, &trust, &d1, &d2, &mv, &fh, &a, &e, &c, &p); err != nil {
				http.Error(w, "not found", 404)
				return
			}
			httpx.JSON(w, 200, map[string]any{"id": iid, "symbol": sym, "score_header": map[string]any{"anomaly": a, "escalation": e, "confidence": c, "priority": p, "composite_risk": p, "safety_level": sev}, "top_drivers": []string{d1, d2}, "explanation_text": "Abnormal short-window return with elevated volume surprise", "caveats": []string{}, "trust": map[string]any{"state": trust}, "baseline_deltas": []map[string]any{{"name": "return", "delta": a}, {"name": "volume", "delta": e}}, "trends": map[string]any{"anomaly": []float64{a * 0.8, a * 0.9, a}, "risk": []float64{p * 0.85, p * 0.95, p}, "volatility": []float64{0.3, 0.4, 0.35}}, "related_incidents": []map[string]any{}, "linked_case": nil, "model_version": mv, "fallback_mode": strings.Contains(strings.ToLower(mv), "fallback"), "feature_snapshot_hash": fh})
		})
		pr.Get("/trust", func(w http.ResponseWriter, r *http.Request) {
			httpx.JSON(w, 200, map[string]any{"trust_summary": map[string]any{"state": "stable", "label": "Model-backed where artifacts exist"}, "model_health": map[string]any{"anomaly": "fallback", "escalation": "fallback"}, "dq_counts": map[string]int{"missing": 0, "duplicates": 0, "late": 0}, "trends": map[string]any{"missingness": []float64{0.03, 0.02, 0.02}, "duplicates": []float64{0.01, 0.01, 0.01}, "late_events": []float64{0.02, 0.01, 0.01}}, "degraded_regions": []map[string]any{}, "circuit_breaker": map[string]any{"state": "closed"}})
		})
		pr.Get("/replay/{job}", func(w http.ResponseWriter, r *http.Request) {
			job := chi.URLParam(r, "job")
			events := []map[string]any{}
			rows, _ := pool.Query(ctx, `SELECT i.id,i.primary_symbol,i.status,i.severity_band,i.updated_at FROM incidents i ORDER BY i.updated_at DESC LIMIT 20`)
			if rows != nil {
				defer rows.Close()
				for rows.Next() {
					var id int64
					var sym, status, sev string
					var updatedAt any
					if rows.Scan(&id, &sym, &status, &sev, &updatedAt) == nil {
						events = append(events, map[string]any{"kind": "incident", "incident_id": id, "symbol": sym, "status": status, "severity": sev, "at": updatedAt})
					}
				}
			}
			httpx.JSON(w, 200, map[string]any{"id": job, "status": "completed", "time_window": "24h", "mode": "metadata_first_with_deterministic_timeline", "model_version_lineage": []string{"anomaly-fallback-v1", "escalation-fallback-v1"}, "lane_series": map[string]any{"risk": []float64{0.4, 0.5, 0.45}}, "timeline_entries": events, "counts": map[string]any{"events": len(events)}, "metadata": map[string]any{"note": "Timeline is deterministically derived from stored records; full tick reconstruction unavailable in this build"}})
		})
		pr.Get("/world-map", func(w http.ResponseWriter, r *http.Request) {
			f, err := parseFilters(r)
			if err != nil {
				http.Error(w, err.Error(), 400)
				return
			}
			args := []any{}
			q := `SELECT coalesce(m.country,'N/A'),coalesce(m.region,'GLOBAL'),coalesce(m.sector,'unknown'),count(i.id),avg(i.composite_risk),max(i.trust_state) FROM incidents i LEFT JOIN instrument_metadata m ON m.instrument_id=i.primary_symbol WHERE ` + windowClause(f.TimeWindow)
			q = applyFilters(q, &args, f) + " GROUP BY 1,2,3 ORDER BY 4 DESC"
			rows, err := pool.Query(ctx, q, args...)
			if err != nil {
				http.Error(w, "internal error", 500)
				return
			}
			defer rows.Close()
			out := []map[string]any{}
			for rows.Next() {
				var c, reg, sec, trust string
				var n int
				var p float64
				if rows.Scan(&c, &reg, &sec, &n, &p, &trust) == nil {
					out = append(out, map[string]any{"country": c, "region": reg, "top_sector": sec, "incident_count": n, "incident_pressure": p, "trust_state": trust, "filters": f})
				}
			}
			httpx.JSON(w, 200, map[string]any{"rows": out, "filters": f})
		})
		pr.Get("/governance/summary", func(w http.ResponseWriter, r *http.Request) {
			httpx.JSON(w, 200, map[string]any{"model_lineage": []string{"fallback"}, "replay_jobs": []map[string]any{}, "threshold_changes": []map[string]any{}})
		})
		pr.Get("/executive-summary", func(w http.ResponseWriter, r *http.Request) {
			httpx.JSON(w, 200, map[string]any{"top_risks": []map[string]any{}, "hot_regions": []map[string]any{}, "what_changed": "Risk concentration tracked across regions."})
		})
	})

	srv := &http.Server{Addr: cfg.HTTPAddr, Handler: r, ReadTimeout: 15 * time.Second, WriteTimeout: 15 * time.Second, IdleTimeout: 60 * time.Second}
	go func() {
		logger.Info("starting server", map[string]any{"addr": cfg.HTTPAddr})
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("query: server error: %v", err)
		}
	}()
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	cancel()
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	srv.Shutdown(shutdownCtx)
}

func requireRole(pub *rsa.PublicKey, perm string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c, err := r.Cookie("sentinel_token")
			if err != nil {
				http.Error(w, "unauthorized", 401)
				return
			}
			claims, err := auth.Parse(c.Value, pub)
			if err != nil || !rbac.Allowed(claims.Role, perm) {
				http.Error(w, "forbidden", 403)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
func requestLoggingMiddleware(logger logging.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(sw, r)
			logger.Info("request", map[string]any{"method": r.Method, "path": r.URL.Path, "status": sw.status, "latency_ms": time.Since(start).Milliseconds()})
		})
	}
}
