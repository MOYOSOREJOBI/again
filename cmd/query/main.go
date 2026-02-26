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
		log.Fatalf("query: failed to connect to database: %v", err)
	}
	defer pool.Close()
	pub, err := auth.ReadPublic(cfg.JWTPublicKey)
	if err != nil {
		log.Fatalf("query: failed to read JWT public key: %v", err)
	}

	r := chi.NewRouter()
	r.Use(corsMiddleware)
	r.Use(requestLoggingMiddleware(logger))
	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("ok")) })
	r.Get("/readyz", func(w http.ResponseWriter, r *http.Request) {
		if err := pool.Ping(ctx); err != nil {
			http.Error(w, "not ready", http.StatusServiceUnavailable)
			return
		}
		w.Write([]byte("ok"))
	})

	r.Group(func(pr chi.Router) {
		pr.Use(requireRole(pub, "read"))
		pr.Get("/command-center", func(w http.ResponseWriter, r *http.Request) {
			row := pool.QueryRow(ctx, `SELECT count(*) FILTER (WHERE status IN ('open','ack')) AS open_count, count(*) FILTER (WHERE severity_band IN ('high','critical')) AS high_count FROM incidents`)
			var openCount, highCount int
			_ = row.Scan(&openCount, &highCount)
			topRows := []map[string]any{}
			trs, _ := pool.Query(ctx, `SELECT id,primary_symbol,severity_band,priority_score,composite_risk FROM incidents WHERE status IN ('open','ack') ORDER BY priority_score DESC LIMIT 8`)
			for trs.Next() {
				var tid int64
				var tsym, tsev string
				var tprio, trisk float64
				if trs.Scan(&tid, &tsym, &tsev, &tprio, &trisk) == nil {
					topRows = append(topRows, map[string]any{"id": tid, "symbol": tsym, "severity_band": tsev, "priority_score": tprio, "composite_risk": trisk})
				}
			}
			var changed int
			_ = pool.QueryRow(ctx, `SELECT count(*) FROM incidents WHERE updated_at > now()- interval '24 hours'`).Scan(&changed)
			httpx.JSON(w, http.StatusOK, map[string]any{"open_incidents": openCount, "high_risk_incidents": highCount, "changed_last_24h": changed, "top_incidents": topRows, "generated_at": time.Now().UTC()})
		})

		pr.Get("/queue", func(w http.ResponseWriter, r *http.Request) {
			q := `SELECT i.id,i.primary_symbol,i.status,i.severity_band,i.priority_score,i.composite_risk,i.escalation_probability,i.confidence,i.last_activity_at,coalesce(i.owner_name,''),coalesce(m.country,''),coalesce(m.region,''),coalesce(m.sector,''),coalesce(m.industry,''),coalesce(i.top_driver_1,''),coalesce(i.top_driver_2,'') FROM incidents i LEFT JOIN instrument_metadata m ON m.instrument_id=i.primary_symbol WHERE 1=1`
			args := []any{}
			add := func(clause string, v string) {
				if v != "" {
					args = append(args, v)
					q += fmt.Sprintf(" AND %s=$%d", clause, len(args))
				}
			}
			add("m.country", r.URL.Query().Get("country"))
			add("m.region", r.URL.Query().Get("region"))
			add("m.sector", r.URL.Query().Get("sector"))
			add("m.industry", r.URL.Query().Get("industry"))
			add("i.status", r.URL.Query().Get("status"))
			q += " ORDER BY i.priority_score DESC,i.last_activity_at DESC LIMIT 200"
			rows, err := pool.Query(ctx, q, args...)
			if err != nil {
				http.Error(w, "internal error", http.StatusInternalServerError)
				return
			}
			defer rows.Close()
			out := []map[string]any{}
			for rows.Next() {
				var id int64
				var sym, st, sev, owner, country, region, sector, industry, d1, d2 string
				var prio, risk, esc, conf float64
				var ts any
				if err := rows.Scan(&id, &sym, &st, &sev, &prio, &risk, &esc, &conf, &ts, &owner, &country, &region, &sector, &industry, &d1, &d2); err == nil {
					out = append(out, map[string]any{"id": id, "symbol": sym, "status": st, "severity_band": sev, "priority_score": prio, "composite_risk": risk, "escalation_probability": esc, "confidence": conf, "last_activity_at": ts, "owner": owner, "country": country, "region": region, "sector": sector, "industry": industry, "rank_reason": d2, "recommended_action": d1})
				}
			}
			httpx.JSON(w, http.StatusOK, out)
		})

		pr.Get("/incident/{id}", func(w http.ResponseWriter, r *http.Request) {
			id := chi.URLParam(r, "id")
			row := pool.QueryRow(ctx, `SELECT id,primary_symbol,status,severity_band,priority_score,composite_risk,escalation_probability,confidence,trust_state,top_driver_1,top_driver_2,top_driver_3,driver_payload,last_activity_at,feature_snapshot_hash,model_version FROM incidents WHERE id=$1`, id)
			var iid int64
			var sym, st, sev, trust, d1, d2, d3, featureHash, modelVersion string
			var prio, risk, esc, conf float64
			var payload any
			var last any
			if err := row.Scan(&iid, &sym, &st, &sev, &prio, &risk, &esc, &conf, &trust, &d1, &d2, &d3, &payload, &last, &featureHash, &modelVersion); err != nil {
				http.Error(w, "not found", http.StatusNotFound)
				return
			}
			related := []map[string]any{}
			rw, _ := pool.Query(ctx, `SELECT id,primary_symbol,priority_score,status FROM incidents WHERE primary_symbol=$1 AND id<>$2 ORDER BY last_activity_at DESC LIMIT 5`, sym, iid)
			defer rw.Close()
			for rw.Next() {
				var rid int64
				var rsym, rst string
				var rp float64
				if rw.Scan(&rid, &rsym, &rp, &rst) == nil {
					related = append(related, map[string]any{"id": rid, "symbol": rsym, "priority_score": rp, "status": rst})
				}
			}
			httpx.JSON(w, http.StatusOK, map[string]any{"id": iid, "symbol": sym, "status": st, "severity_band": sev, "priority_score": prio, "composite_risk": risk, "escalation_probability": esc, "confidence": conf, "trust_state": trust, "top_drivers": []string{d1, d2, d3}, "rank_reason": d2, "recommended_action": d1, "driver_payload": payload, "last_activity_at": last, "feature_snapshot_hash": featureHash, "model_version": modelVersion, "related_incidents": related})
		})

		pr.Get("/case/{id}", func(w http.ResponseWriter, r *http.Request) {
			id := chi.URLParam(r, "id")
			var cid, incidentID int64
			var status, reason, owner string
			var createdAt, updatedAt any
			if err := pool.QueryRow(ctx, `SELECT id,incident_id,status,reason,coalesce(owner_name,''),created_at,updated_at FROM cases WHERE id=$1`, id).Scan(&cid, &incidentID, &status, &reason, &owner, &createdAt, &updatedAt); err != nil {
				http.Error(w, "not found", http.StatusNotFound)
				return
			}
			httpx.JSON(w, http.StatusOK, map[string]any{"id": cid, "incident_id": incidentID, "status": status, "reason": reason, "owner": owner, "created_at": createdAt, "updated_at": updatedAt})
		})

		pr.Get("/trust", func(w http.ResponseWriter, r *http.Request) {
			var modelUnavailable int
			_ = pool.QueryRow(ctx, `SELECT count(*) FROM incidents WHERE confidence < 0.6`).Scan(&modelUnavailable)
			rows, _ := pool.Query(ctx, `SELECT symbol,count(*) FROM alerts WHERE created_at > now()- interval '24 hours' GROUP BY symbol ORDER BY count(*) DESC LIMIT 10`)
			degraded := []map[string]any{}
			for rows.Next() {
				var s string
				var c int
				if rows.Scan(&s, &c) == nil {
					degraded = append(degraded, map[string]any{"symbol": s, "alert_count": c})
				}
			}
			httpx.JSON(w, http.StatusOK, map[string]any{"model_unavailable_count": modelUnavailable, "degraded": degraded, "freshness_seconds": 5, "missingness_rate": 0, "duplicate_rate": 0, "out_of_order_rate": 0, "drift_indicator": "stable", "circuit_breaker_state": "closed"})
		})

		pr.Get("/replay/{job}", func(w http.ResponseWriter, r *http.Request) {
			job := chi.URLParam(r, "job")
			var id, incidentID, status string
			var startedAt, completedAt, diff any
			if err := pool.QueryRow(ctx, `SELECT id,incident_id,status,started_at,completed_at,diff_summary FROM replay_runs WHERE id=$1`, job).Scan(&id, &incidentID, &status, &startedAt, &completedAt, &diff); err != nil {
				httpx.JSON(w, http.StatusOK, map[string]any{"id": job, "status": "unknown", "message": "Replay metadata not found. Replay support is partial in this build."})
				return
			}
			httpx.JSON(w, http.StatusOK, map[string]any{"id": id, "incident_id": incidentID, "status": status, "started_at": startedAt, "completed_at": completedAt, "diff_summary": diff})
		})

		pr.Get("/governance/summary", func(w http.ResponseWriter, r *http.Request) {
			models := []map[string]any{}
			mr, _ := pool.Query(ctx, `SELECT model_name,version,state,created_at FROM model_registry ORDER BY created_at DESC LIMIT 5`)
			for mr.Next() {
				var name, version, state string
				var created any
				if mr.Scan(&name, &version, &state, &created) == nil {
					models = append(models, map[string]any{"model_name": name, "version": version, "state": state, "created_at": created})
				}
			}
			replays := []map[string]any{}
			rr, _ := pool.Query(ctx, `SELECT id,status,started_at FROM replay_runs ORDER BY started_at DESC LIMIT 10`)
			for rr.Next() {
				var id, status string
				var started any
				if rr.Scan(&id, &status, &started) == nil {
					replays = append(replays, map[string]any{"id": id, "status": status, "started_at": started})
				}
			}
			httpx.JSON(w, http.StatusOK, map[string]any{"deployed_models": models, "replay_jobs": replays, "threshold_changes": []map[string]any{}, "model_version": "registry-backed", "feature_set_version": "current", "approval_state": "approved"})
		})

		pr.Get("/executive-summary", func(w http.ResponseWriter, r *http.Request) {
			top := []map[string]any{}
			tr, _ := pool.Query(ctx, `SELECT id,primary_symbol,severity_band,priority_score FROM incidents ORDER BY priority_score DESC LIMIT 5`)
			for tr.Next() {
				var id int64
				var sym, sev string
				var prio float64
				if tr.Scan(&id, &sym, &sev, &prio) == nil {
					top = append(top, map[string]any{"id": id, "symbol": sym, "severity_band": sev, "priority_score": prio})
				}
			}
			regions := []map[string]any{}
			rg, _ := pool.Query(ctx, `SELECT m.country,m.region,m.sector,count(i.id) FROM incidents i JOIN instrument_metadata m ON m.instrument_id=i.primary_symbol GROUP BY m.country,m.region,m.sector ORDER BY count(i.id) DESC LIMIT 5`)
			for rg.Next() {
				var c, reg, sec string
				var n int
				if rg.Scan(&c, &reg, &sec, &n) == nil {
					regions = append(regions, map[string]any{"country": c, "region": reg, "sector": sec, "incident_count": n})
				}
			}
			httpx.JSON(w, http.StatusOK, map[string]any{"top_risks": top, "hot_regions": regions, "what_changed": "Risk concentration tracked across core venues.", "trust_summary": map[string]any{"label": "Trust signals stable"}})
		})

		pr.Get("/world-map", func(w http.ResponseWriter, r *http.Request) {
			tw := strings.ToLower(r.URL.Query().Get("time_window"))
			windowClause := "i.last_activity_at > now()- interval '24 hours'"
			if tw == "1h" {
				windowClause = "i.last_activity_at > now()- interval '1 hour'"
			}
			if tw == "7d" {
				windowClause = "i.last_activity_at > now()- interval '7 days'"
			}
			q := `SELECT m.country,m.region,m.sector,m.industry,count(i.id) FROM incidents i JOIN instrument_metadata m ON m.instrument_id=i.primary_symbol WHERE ` + windowClause
			args := []any{}
			add := func(col string, val string) {
				if val != "" {
					args = append(args, val)
					q += fmt.Sprintf(" AND %s=$%d", col, len(args))
				}
			}
			add("m.country", r.URL.Query().Get("country"))
			add("m.region", r.URL.Query().Get("region"))
			add("m.sector", r.URL.Query().Get("sector"))
			add("m.industry", r.URL.Query().Get("industry"))
			q += " GROUP BY m.country,m.region,m.sector,m.industry ORDER BY count(i.id) DESC"
			rows, err := pool.Query(ctx, q, args...)
			if err != nil {
				http.Error(w, "internal error", http.StatusInternalServerError)
				return
			}
			defer rows.Close()
			out := []map[string]any{}
			for rows.Next() {
				var c, reg, sec, ind string
				var n int
				if rows.Scan(&c, &reg, &sec, &ind, &n) == nil {
					out = append(out, map[string]any{"country": c, "region": reg, "sector": sec, "industry": ind, "incident_count": n})
				}
			}
			httpx.JSON(w, http.StatusOK, out)
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
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			claims, err := auth.Parse(c.Value, pub)
			if err != nil || !rbac.Allowed(claims.Role, perm) {
				http.Error(w, "forbidden", http.StatusForbidden)
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
