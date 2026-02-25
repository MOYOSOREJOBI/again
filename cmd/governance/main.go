package main

import (
	"context"
	"crypto/rsa"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/segmentio/ksuid"
	"sentinel/internal/audit"
	"sentinel/internal/auth"
	"sentinel/internal/config"
	"sentinel/internal/db"
	"sentinel/internal/httpx"
	"sentinel/internal/logging"
	"sentinel/internal/rbac"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	logger := logging.New("governance")
	cfg := config.Load("governance")
	if err := config.Validate(cfg); err != nil {
		log.Fatalf("governance: invalid config: %v", err)
	}

	pool, err := db.Connect(ctx, cfg.PostgresURL)
	if err != nil {
		log.Fatalf("governance: failed to connect to database: %v", err)
	}
	defer pool.Close()

	pub, err := auth.ReadPublic(cfg.JWTPublicKey)
	if err != nil {
		log.Fatalf("governance: failed to read JWT public key: %v", err)
	}

	r := chi.NewRouter()
	r.Use(corsMiddleware)
	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("ok")) })
	r.Get("/readyz", readinessHandler(pool))

	r.Post("/models/deploy", func(w http.ResponseWriter, r *http.Request) {
		claims, ok := authn(r, pub)
		if !ok || !(rbac.Allowed(claims.Role, "*") || rbac.Allowed(claims.Role, "model:deploy")) {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		var in map[string]string
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&in); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		modelName := in["model_name"]
		version := in["version"]
		if modelName == "" || version == "" {
			http.Error(w, "model_name and version required", http.StatusBadRequest)
			return
		}
		if _, err := pool.Exec(ctx, `INSERT INTO model_registry(model_name,version,state) VALUES($1,$2,'deployed') ON CONFLICT (model_name) DO UPDATE SET version=$2,state='deployed'`, modelName, version); err != nil {
			logger.Error("deploy model failed", map[string]any{"error": err.Error()})
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		if err := audit.Append(ctx, pool, claims.Subject, "model.deploy", modelName+":"+version); err != nil {
			logger.Error("audit append failed", map[string]any{"error": err.Error()})
		}
		w.WriteHeader(http.StatusNoContent)
	})

	r.Post("/replay/start", func(w http.ResponseWriter, r *http.Request) {
		claims, ok := authn(r, pub)
		if !ok || !(rbac.Allowed(claims.Role, "replay:write") || rbac.Allowed(claims.Role, "*")) {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		id := ksuid.New().String()
		if _, err := pool.Exec(ctx, `INSERT INTO replay_runs(id,incident_id,status,diff_summary) VALUES($1,'demo','completed',$2)`, id, `{"delta":0}`); err != nil {
			logger.Error("start replay failed", map[string]any{"error": err.Error()})
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		if err := audit.Append(ctx, pool, claims.Subject, "replay.start", id); err != nil {
			logger.Error("audit append failed", map[string]any{"error": err.Error()})
		}
		httpx.JSON(w, http.StatusOK, map[string]string{"id": id})
	})

	r.Get("/models", func(w http.ResponseWriter, r *http.Request) {
		claims, ok := authn(r, pub)
		if !ok || !rbac.Allowed(claims.Role, "read") {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		rows, err := pool.Query(ctx, `SELECT id,model_name,version,state,created_at FROM model_registry ORDER BY id DESC`)
		if err != nil {
			logger.Error("query models failed", map[string]any{"error": err.Error()})
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		defer rows.Close()
		out := []map[string]any{}
		for rows.Next() {
			var id int
			var name, version, state string
			var ts interface{}
			if err := rows.Scan(&id, &name, &version, &state, &ts); err != nil {
				continue
			}
			out = append(out, map[string]any{"id": id, "model_name": name, "version": version, "state": state, "created_at": ts})
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
			log.Fatalf("governance: server error: %v", err)
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

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

type pinger interface {
	Ping(context.Context) error
}

func readinessHandler(p pinger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		probeCtx, probeCancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer probeCancel()
		if err := p.Ping(probeCtx); err != nil {
			http.Error(w, "database not ready", http.StatusServiceUnavailable)
			return
		}
		w.Write([]byte("ok"))
	}
}

var errNotReady = errors.New("not ready")
