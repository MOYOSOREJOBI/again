package main

import (
	"context"
	"crypto/rsa"
	"log"
	"net/http"
	"os"
	"os/signal"
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

func (s *statusWriter) WriteHeader(code int) {
	s.status = code
	s.ResponseWriter.WriteHeader(code)
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
		pr.Get("/queue", func(w http.ResponseWriter, r *http.Request) {
			rows, err := pool.Query(ctx, `SELECT id,primary_symbol,status,severity_band,priority_score,composite_risk,escalation_probability,confidence,last_activity_at,coalesce(owner_name,'') FROM incidents ORDER BY priority_score DESC,last_activity_at DESC LIMIT 100`)
			if err != nil {
				http.Error(w, "internal error", http.StatusInternalServerError)
				return
			}
			defer rows.Close()
			out := []map[string]any{}
			for rows.Next() {
				var id int64
				var sym, st, sev, owner string
				var prio, risk, esc, conf float64
				var ts interface{}
				if err := rows.Scan(&id, &sym, &st, &sev, &prio, &risk, &esc, &conf, &ts, &owner); err == nil {
					out = append(out, map[string]any{"id": id, "symbol": sym, "status": st, "severity_band": sev, "priority_score": prio, "composite_risk": risk, "escalation_probability": esc, "confidence": conf, "last_activity_at": ts, "owner": owner})
				}
			}
			httpx.JSON(w, http.StatusOK, out)
		})

		pr.Get("/incident/{id}", func(w http.ResponseWriter, r *http.Request) {
			id := chi.URLParam(r, "id")
			row := pool.QueryRow(ctx, `SELECT id,primary_symbol,status,severity_band,priority_score,composite_risk,escalation_probability,confidence,trust_state,top_driver_1,top_driver_2,top_driver_3,driver_payload,last_activity_at FROM incidents WHERE id=$1`, id)
			var iid int64
			var sym, st, sev, trust, d1, d2, d3 string
			var prio, risk, esc, conf float64
			var payload any
			var last any
			if err := row.Scan(&iid, &sym, &st, &sev, &prio, &risk, &esc, &conf, &trust, &d1, &d2, &d3, &payload, &last); err != nil {
				http.Error(w, "not found", http.StatusNotFound)
				return
			}
			httpx.JSON(w, http.StatusOK, map[string]any{"id": iid, "symbol": sym, "status": st, "severity_band": sev, "priority_score": prio, "composite_risk": risk, "escalation_probability": esc, "confidence": conf, "trust_state": trust, "top_drivers": []string{d1, d2, d3}, "driver_payload": payload, "last_activity_at": last})
		})

		pr.Get("/trust", func(w http.ResponseWriter, r *http.Request) {
			rows, err := pool.Query(ctx, `SELECT symbol,count(*) FROM incidents WHERE trust_state <> 'stable' AND last_activity_at > now()- interval '24 hours' GROUP BY symbol ORDER BY count(*) DESC LIMIT 20`)
			if err != nil {
				http.Error(w, "internal error", http.StatusInternalServerError)
				return
			}
			defer rows.Close()
			out := []map[string]any{}
			for rows.Next() {
				var sym string
				var c int
				if err := rows.Scan(&sym, &c); err == nil {
					out = append(out, map[string]any{"symbol": sym, "degraded_count": c})
				}
			}
			httpx.JSON(w, http.StatusOK, map[string]any{"degraded": out})
		})

		pr.Get("/world-map", func(w http.ResponseWriter, r *http.Request) {
			rows, err := pool.Query(ctx, `SELECT m.country,m.region,m.sector,count(i.id) FROM incidents i JOIN instrument_metadata m ON m.instrument_id=i.primary_symbol GROUP BY m.country,m.region,m.sector ORDER BY count(i.id) DESC`)
			if err != nil {
				http.Error(w, "internal error", http.StatusInternalServerError)
				return
			}
			defer rows.Close()
			out := []map[string]any{}
			for rows.Next() {
				var c, reg, sec string
				var n int
				if err := rows.Scan(&c, &reg, &sec, &n); err == nil {
					out = append(out, map[string]any{"country": c, "region": reg, "sector": sec, "incident_count": n})
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
	logger.Info("shutting down", nil)
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
