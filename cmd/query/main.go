package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"sentinel/internal/config"
	"sentinel/internal/db"
	"sentinel/internal/httpx"
	"sentinel/internal/logging"
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

	r.Get("/alerts", func(w http.ResponseWriter, r *http.Request) {
		rows, err := pool.Query(ctx, `SELECT id,symbol,status,created_at FROM alerts ORDER BY id DESC LIMIT 50`)
		if err != nil {
			logger.Error("query alerts failed", map[string]any{"error": err.Error(), "path": r.URL.Path})
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		defer rows.Close()
		out := []map[string]any{}
		for rows.Next() {
			var id int
			var sym, st string
			var ts interface{}
			if err := rows.Scan(&id, &sym, &st, &ts); err != nil {
				logger.Error("scan alert failed", map[string]any{"error": err.Error()})
				continue
			}
			out = append(out, map[string]any{"id": id, "symbol": sym, "status": st, "ts": ts, "created_at": ts})
		}
		if err := rows.Err(); err != nil {
			logger.Error("rows iteration error", map[string]any{"error": err.Error(), "path": r.URL.Path})
		}
		httpx.JSON(w, http.StatusOK, out)
	})

	r.Get("/scores", func(w http.ResponseWriter, r *http.Request) {
		rows, err := pool.Query(ctx, `SELECT id,symbol,score,severity,ts FROM scores ORDER BY id DESC LIMIT 50`)
		if err != nil {
			logger.Error("query scores failed", map[string]any{"error": err.Error(), "path": r.URL.Path})
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		defer rows.Close()
		out := []map[string]any{}
		for rows.Next() {
			var id int
			var sym, sev string
			var score float64
			var ts interface{}
			if err := rows.Scan(&id, &sym, &score, &sev, &ts); err != nil {
				logger.Error("scan score failed", map[string]any{"error": err.Error()})
				continue
			}
			out = append(out, map[string]any{"id": id, "symbol": sym, "score": score, "severity": sev, "ts": ts})
		}
		if err := rows.Err(); err != nil {
			logger.Error("rows iteration error", map[string]any{"error": err.Error(), "path": r.URL.Path})
		}
		httpx.JSON(w, http.StatusOK, out)
	})

	srv := &http.Server{
		Addr:         cfg.HTTPAddr,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

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
