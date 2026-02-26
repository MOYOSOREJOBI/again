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
	"sentinel/internal/audit"
	"sentinel/internal/auth"
	"sentinel/internal/cache"
	"sentinel/internal/config"
	"sentinel/internal/db"
	"sentinel/internal/httpx"
	"sentinel/internal/logging"
	"sentinel/internal/middleware"
	"sentinel/internal/rbac"
	"sentinel/internal/replay"
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
	r.Use(middleware.RequestID)
	r.Use(middleware.CORS)
	r.Use(middleware.RequireCSRFFunc(func(r *http.Request) (string, bool) {
		claims, ok := authn(r, pub)
		if !ok {
			return "", false
		}
		return claims.Subject, true
	}))
	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("ok")) })
	r.Get("/readyz", readinessHandler(pool))
	r.Get("/active-models", func(w http.ResponseWriter, r *http.Request) {
		claims, ok := authn(r, pub)
		if !ok || !rbac.Allowed(claims.Role, "read") {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		rows, err := pool.Query(ctx, `SELECT model_name,model_version,artifact_hash,artifact_path,feature_set_version,coalesce(calibration_version,''),deployed_at FROM model_deployments WHERE status='deployed' ORDER BY deployed_at DESC NULLS LAST`)
		if err != nil {
			logger.Error("load active models failed", map[string]any{"error": err.Error()})
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		defer rows.Close()
		out := map[string]any{"models": []map[string]any{}}
		models := []map[string]any{}
		for rows.Next() {
			var name, version, artifactHash, artifactPath, featureSetVersion, calibrationVersion string
			var deployedAt any
			if err := rows.Scan(&name, &version, &artifactHash, &artifactPath, &featureSetVersion, &calibrationVersion, &deployedAt); err != nil {
				continue
			}
			models = append(models, map[string]any{
				"model_name":          name,
				"model_version":       version,
				"artifact_hash":       artifactHash,
				"artifact_path":       artifactPath,
				"feature_set_version": featureSetVersion,
				"calibration_version": calibrationVersion,
				"deployed_at":         deployedAt,
			})
		}
		out["models"] = models
		httpx.JSON(w, http.StatusOK, out)
	})

	r.With(middleware.RateLimit(rateLimitSubjectKey("gov"), 10, 1*time.Minute)).Post("/models/deploy", func(w http.ResponseWriter, r *http.Request) {
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
		_, _ = pool.Exec(ctx, `INSERT INTO model_deployments(model_name,model_version,artifact_hash,artifact_path,feature_set_version,calibration_version,status,created_by,approved_by,approved_at,deployed_at,change_reason) VALUES($1,$2,$3,$4,$5,$6,'deployed',$7,$7,now(),now(),$8)`, modelName, version, "inline", "/registry/"+modelName+":"+version, "v2", "", claims.Subject, "api deploy")
		if err := audit.Append(ctx, pool, claims.Subject, "model.deploy", modelName+":"+version); err != nil {
			logger.Error("audit append failed", map[string]any{"error": err.Error()})
		}
		cache.InvalidateByPrefixes(ctx, cache.ReadModelPrefixes()...)
		w.WriteHeader(http.StatusNoContent)
	})

	r.With(middleware.RateLimit(rateLimitSubjectKey("workflow"), 30, 1*time.Minute)).Post("/replay/start", func(w http.ResponseWriter, r *http.Request) {
		claims, ok := authn(r, pub)
		if !ok || !(rbac.Allowed(claims.Role, "replay:write") || rbac.Allowed(claims.Role, "*")) {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		var req struct {
			Start             time.Time `json:"start"`
			End               time.Time `json:"end"`
			ModelVersion      string    `json:"model_version"`
			FeatureSetVersion string    `json:"feature_set_version"`
			ReplayMode        string    `json:"replay_mode"`
		}
		_ = json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&req)
		if req.Start.IsZero() {
			req.Start = time.Now().UTC().Add(-1 * time.Hour)
		}
		if req.End.IsZero() {
			req.End = time.Now().UTC()
		}
		if req.ModelVersion == "" {
			req.ModelVersion = "baseline-v1"
		}
		if req.FeatureSetVersion == "" {
			req.FeatureSetVersion = "v2"
		}
		req.ReplayMode = normalizeReplayMode(req.ReplayMode)
		var id string
		if err := pool.QueryRow(ctx, `INSERT INTO replay_jobs(requested_by,status,time_window_start,time_window_end,replay_mode,watermark_policy_id,allowed_lateness_ms,model_version,feature_set_version) VALUES($1,'queued',$2,$3,$4,'wm_v1',5000,$5,$6) RETURNING id::text`, claims.Subject, req.Start, req.End, req.ReplayMode, req.ModelVersion, req.FeatureSetVersion).Scan(&id); err != nil {
			logger.Error("start replay failed", map[string]any{"error": err.Error()})
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		if err := audit.Append(ctx, pool, claims.Subject, "replay.start", id); err != nil {
			logger.Error("audit append failed", map[string]any{"error": err.Error()})
		}
		go func(id string) {
			_ = replay.Run(context.Background(), pool, id)
			cache.InvalidateByPrefixes(context.Background(), cache.ReadModelPrefixes()...)
		}(id)
		httpx.JSON(w, http.StatusAccepted, map[string]any{"id": id, "status": "queued", "replay_mode": req.ReplayMode})
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

func rateLimitSubjectKey(prefix string) func(*http.Request) string {
	return func(r *http.Request) string {
		if authz := r.Header.Get("Authorization"); len(authz) > 7 {
			return prefix + ":" + authz
		}
		if c, err := r.Cookie("sentinel_token"); err == nil && c.Value != "" {
			return prefix + ":cookie:" + c.Value
		}
		return prefix + ":ip:" + r.RemoteAddr
	}
}

func normalizeReplayMode(in string) string {
	switch in {
	case "as_scored", "recompute":
		return in
	default:
		return "recompute"
	}
}
