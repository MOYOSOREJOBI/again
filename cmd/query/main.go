package main

import (
	"context"
	"crypto/rsa"
	"encoding/json"
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
	"sentinel/internal/httpx"
	"sentinel/internal/logging"
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

type filters struct{ TimeWindow, Country, Region, Industry string }

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
	r.Get("/readyz", func(w http.ResponseWriter, r *http.Request) {
		if err := pool.Ping(r.Context()); err != nil {
			http.Error(w, "not ready", 503)
			return
		}
		_, _ = w.Write([]byte("ok"))
	})

	r.Group(func(pr chi.Router) {
		pr.Use(requireRole(pub, "read"))
		pr.Get("/queue", func(w http.ResponseWriter, r *http.Request) {
			f := parseFilters(r)
			claims, _ := authn(r, pub)
			key := cache.QueueKey(claims.Role, f.Region, f.Country, f.Industry, f.TimeWindow)
			out, err := cache.GetOrLoadJSON(r.Context(), key, ttlQueue, func() ([]qrm.QueueRow, error) {
				return qrm.LoadQueue(r.Context(), pool, qrm.QueueFilters{TimeWindow: f.TimeWindow, Country: f.Country, Region: f.Region, Industry: f.Industry})
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
			key := cache.CommandCenterKey(claims.Role, f.Region, f.Industry, f.TimeWindow)
			out, err := cache.GetOrLoadJSON(r.Context(), key, ttlSummary, func() (map[string]any, error) { return qrm.LoadCommandCenter(r.Context(), pool, f.TimeWindow) })
			if err != nil {
				http.Error(w, "internal", 500)
				return
			}
			httpx.JSON(w, 200, out)
		})
		pr.Get("/trust", func(w http.ResponseWriter, r *http.Request) {
			f := parseFilters(r)
			key := cache.TrustKey(f.Region, f.Industry, f.TimeWindow)
			out, err := cache.GetOrLoadJSON(r.Context(), key, ttlSummary, func() (map[string]any, error) { return qrm.LoadTrust(r.Context(), pool) })
			if err != nil {
				http.Error(w, "internal", 500)
				return
			}
			httpx.JSON(w, 200, out)
		})
		pr.Get("/world-map", func(w http.ResponseWriter, r *http.Request) {
			f := parseFilters(r)
			key := cache.WorldMapKey(f.Region, f.Industry, f.TimeWindow)
			countries, err := cache.GetOrLoadJSON(r.Context(), key, ttlWorldMap, func() ([]qrm.CountryAgg, error) { return qrm.LoadWorldMap(r.Context(), pool, f.TimeWindow) })
			if err != nil {
				http.Error(w, "internal", 500)
				return
			}
			httpx.JSON(w, 200, map[string]any{"countries": countries, "timeWindow": f.TimeWindow})
		})
		pr.Get("/executive-summary", func(w http.ResponseWriter, r *http.Request) {
			f := parseFilters(r)
			key := cache.ExecutiveKey(f.Region, f.Industry, f.TimeWindow)
			out, err := cache.GetOrLoadJSON(r.Context(), key, ttlExecutive, func() (map[string]any, error) {
				return map[string]any{"topRisks": []any{}, "countryConcentration": []any{}, "industryConcentration": []any{}, "trustSummary": map[string]any{"state": "stable"}, "riskMemo": "No material change."}, nil
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
			rp, err := pool.Query(r.Context(), `SELECT id::text,status,requested_at,started_at,completed_at,model_version,feature_set_version,watermark_policy_id,allowed_lateness_ms FROM replay_jobs ORDER BY requested_at DESC LIMIT 25`)
			if err == nil {
				defer rp.Close()
				for rp.Next() {
					var id, st, mv, fv, wp string
					var req, stt, ct any
					var late int
					if rp.Scan(&id, &st, &req, &stt, &ct, &mv, &fv, &wp, &late) == nil {
						replays = append(replays, map[string]any{"id": id, "status": st, "requestedAt": req, "startedAt": stt, "completedAt": ct, "modelVersion": mv, "featureSetVersion": fv, "watermarkPolicy": wp, "allowedLatenessMs": late})
					}
				}
			}
			httpx.JSON(w, 200, map[string]any{"modelLineage": lineage, "replayJobs": replays, "thresholdChanges": []any{}})
		})
		pr.Get("/replay/{job}", func(w http.ResponseWriter, r *http.Request) {
			id := chi.URLParam(r, "job")
			var status, mv, fv, wp string
			var late int
			var s, e, started, completed any
			if err := pool.QueryRow(r.Context(), `SELECT status,time_window_start,time_window_end,started_at,completed_at,model_version,feature_set_version,watermark_policy_id,allowed_lateness_ms FROM replay_jobs WHERE id=$1`, id).Scan(&status, &s, &e, &started, &completed, &mv, &fv, &wp, &late); err != nil {
				http.Error(w, "not found", 404)
				return
			}
			var result any = map[string]any{}
			_ = pool.QueryRow(r.Context(), `SELECT diff_summary FROM replay_runs WHERE id=$1`, id).Scan(&result)
			partial := status != "completed"
			httpx.JSON(w, 200, map[string]any{"id": id, "status": status, "timeWindowStart": s, "timeWindowEnd": e, "startedAt": started, "completedAt": completed, "modelVersion": mv, "featureSetVersion": fv, "watermarkPolicy": wp, "allowedLatenessMs": late, "result": result, "partial": partial, "limitations": []string{"incident projection replay remains simplified"}})
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
			rows, _ := qrm.LoadQueue(ctx, pool, qrm.QueueFilters{TimeWindow: "24h"})
			inc := map[string]any{"id": 0}
			if len(rows) > 0 {
				inc = map[string]any{"id": rows[0].ID, "symbol": rows[0].Symbol, "priorityScore": rows[0].PriorityScore, "severityBand": rows[0].SeverityBand, "recommendedAction": rows[0].RecommendedAction}
			}
			return map[string]any{"type": "upsert", "incident": inc}
		}, "queue_patch"))
		pr.Get("/stream/command-center", sseStream(func(ctx context.Context) any {
			cc, _ := qrm.LoadCommandCenter(ctx, pool, "24h")
			return cc
		}, "command_center_patch"))
		pr.Get("/stream/trust", sseStream(func(ctx context.Context) any {
			t, _ := qrm.LoadTrust(ctx, pool)
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
	tw := q.Get("time_window")
	if tw == "" {
		tw = "24h"
	}
	return filters{TimeWindow: tw, Country: q.Get("country"), Region: q.Get("region"), Industry: q.Get("industry")}
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
