package main

import (
	"context"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/twmb/franz-go/pkg/kgo"
	"sentinel/internal/audit"
	"sentinel/internal/auth"
	"sentinel/internal/config"
	"sentinel/internal/db"
	"sentinel/internal/httpx"
	"sentinel/internal/incidents"
	"sentinel/internal/kafka"
	"sentinel/internal/logging"
	"sentinel/internal/rbac"
)

var subs = struct {
	sync.Mutex
	conns []chan string
}{}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	logger := logging.New("alerts")
	cfg := config.Load("alerts")
	if err := config.Validate(cfg); err != nil {
		log.Fatalf("alerts: invalid config: %v", err)
	}

	pool, err := db.Connect(ctx, cfg.PostgresURL)
	if err != nil {
		log.Fatalf("alerts: failed to connect to database: %v", err)
	}
	defer pool.Close()

	pub, err := auth.ReadPublic(cfg.JWTPublicKey)
	if err != nil {
		log.Fatalf("alerts: failed to read JWT public key: %v", err)
	}

	c, err := kafka.New(cfg.KafkaBrokers, "alerts", "derived.scores")
	if err != nil {
		log.Fatalf("alerts: failed to create kafka consumer: %v", err)
	}
	defer c.Close()

	p, err := kafka.New(cfg.KafkaBrokers, "")
	if err != nil {
		log.Fatalf("alerts: failed to create kafka producer: %v", err)
	}
	defer p.Close()

	go func() {
		for {
			f := c.PollFetches(ctx)
			if ctx.Err() != nil {
				return
			}
			f.EachRecord(func(rec *kgo.Record) {
				var m map[string]any
				if err := json.Unmarshal(rec.Value, &m); err != nil {
					logger.Error("unmarshal score failed", map[string]any{"error": err.Error()})
					return
				}
				symbol, _ := m["symbol"].(string)
				sev, _ := m["severity"].(string)
				explanation, _ := m["explanation"].(string)
				score, _ := m["score"].(float64)
				escalation, _ := m["escalation_probability"].(float64)
				confidence, _ := m["confidence"].(float64)
				featureHash, _ := m["feature_snapshot_hash"].(string)
				modelVersion, _ := m["model_version"].(string)
				if symbol == "" {
					return
				}
				if confidence == 0 {
					confidence = 0.8
				}
				if modelVersion == "" {
					modelVersion = "baseline-v1"
				}
				ts := scoreTS(m)
				idempotency := scoreIdempotencyKey(symbol, ts, score, sev, explanation)
				var scoreID int64
				if err := pool.QueryRow(ctx, `INSERT INTO scores(idempotency_key,symbol,ts,score,severity,explanation,raw_anomaly_score,normalized_anomaly_score,escalation_probability,priority_score,composite_risk,feature_snapshot_hash,model_version,explanation_payload) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$10,$11,$12,$13) ON CONFLICT (idempotency_key,ts) DO UPDATE SET idempotency_key=EXCLUDED.idempotency_key RETURNING id`, idempotency, symbol, ts, score, sev, explanation, score, score, escalation, score, featureHash, modelVersion, map[string]any{"explanation": explanation}).Scan(&scoreID); err != nil {
					logger.Error("insert score failed", map[string]any{"error": err.Error()})
					return
				}
				if sev == "" || sev == "low" {
					return
				}
				incidentID, err := incidents.FindOrCreateOpenIncident(ctx, pool, symbol, score, escalation, confidence, featureHash, modelVersion, ts)
				if err != nil {
					logger.Error("incident upsert failed", map[string]any{"error": err.Error()})
					return
				}
				var alertID int64
				if err := pool.QueryRow(ctx, `INSERT INTO alerts(score_id,incident_id,symbol,status,justification) VALUES($1,$2,$3,'open','auto') RETURNING id`, scoreID, incidentID, symbol).Scan(&alertID); err != nil {
					logger.Error("insert alert failed", map[string]any{"error": err.Error()})
					return
				}
				if _, err := pool.Exec(ctx, `INSERT INTO incident_alert_links(incident_id,alert_id) VALUES($1,$2) ON CONFLICT DO NOTHING`, incidentID, alertID); err != nil {
					logger.Error("link incident alert failed", map[string]any{"error": err.Error()})
				}
				if err := kafka.ProduceJSON(ctx, p, "alerts.created", symbol, m); err != nil {
					logger.Error("produce alert msg failed", map[string]any{"error": err.Error()})
				}
				if err := audit.Append(ctx, pool, "system", "alert.create", symbol); err != nil {
					logger.Error("audit append failed", map[string]any{"error": err.Error()})
				}
				broadcast(fmt.Sprintf("event: alert\ndata: %s\n\n", string(rec.Value)))
			})
		}
	}()

	r := chi.NewRouter()
	r.Use(corsMiddleware)
	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("ok")) })
	r.Get("/readyz", func(w http.ResponseWriter, r *http.Request) {
		probeCtx, probeCancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer probeCancel()
		if err := pool.Ping(probeCtx); err != nil {
			http.Error(w, "database not ready", http.StatusServiceUnavailable)
			return
		}
		if err := c.Ping(probeCtx); err != nil {
			http.Error(w, "kafka consumer not ready", http.StatusServiceUnavailable)
			return
		}
		if err := p.Ping(probeCtx); err != nil {
			http.Error(w, "kafka producer not ready", http.StatusServiceUnavailable)
			return
		}
		w.Write([]byte("ok"))
	})
	r.Get("/sse/alerts", sse)
	r.Post("/alerts/{id}/ack", func(w http.ResponseWriter, r *http.Request) {
		claims, ok := authn(r, pub)
		if !ok || !(rbac.Allowed(claims.Role, "alerts:write") || rbac.Allowed(claims.Role, "*")) {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		id := chi.URLParam(r, "id")
		if _, err := pool.Exec(ctx, `UPDATE alerts SET status='ack',updated_at=now() WHERE id=$1`, id); err != nil {
			logger.Error("ack alert failed", map[string]any{"error": err.Error(), "id": id})
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		if err := audit.Append(ctx, pool, claims.Subject, "alert.ack", id); err != nil {
			logger.Error("audit append failed", map[string]any{"error": err.Error(), "id": id})
		}
		w.WriteHeader(http.StatusNoContent)
	})

	r.Post("/incidents/{id}/promote-case", func(w http.ResponseWriter, r *http.Request) {
		claims, ok := authn(r, pub)
		if !ok || !(rbac.Allowed(claims.Role, "alerts:write") || rbac.Allowed(claims.Role, "*")) {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		id := chi.URLParam(r, "id")
		var in struct {
			Reason string `json:"reason"`
		}
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&in); err != nil || in.Reason == "" {
			http.Error(w, "reason required", http.StatusBadRequest)
			return
		}
		var caseID int64
		if err := pool.QueryRow(ctx, `INSERT INTO cases(incident_id,reason,owner_name) VALUES($1,$2,$3) RETURNING id`, id, in.Reason, claims.Subject).Scan(&caseID); err != nil {
			http.Error(w, "failed to create case", http.StatusInternalServerError)
			return
		}
		_ = audit.Append(ctx, pool, claims.Subject, "case.create", fmt.Sprintf("%d", caseID))
		httpx.JSON(w, http.StatusCreated, map[string]any{"case_id": caseID})
	})

	r.Post("/cases/{id}/notes", func(w http.ResponseWriter, r *http.Request) {
		claims, ok := authn(r, pub)
		if !ok || !(rbac.Allowed(claims.Role, "alerts:write") || rbac.Allowed(claims.Role, "*")) {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		id := chi.URLParam(r, "id")
		var in struct {
			Note string `json:"note"`
		}
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&in); err != nil || in.Note == "" {
			http.Error(w, "note required", http.StatusBadRequest)
			return
		}
		if _, err := pool.Exec(ctx, `INSERT INTO case_notes(case_id,actor,note) VALUES($1,$2,$3)`, id, claims.Subject, in.Note); err != nil {
			http.Error(w, "failed to add note", http.StatusInternalServerError)
			return
		}
		_ = audit.Append(ctx, pool, claims.Subject, "case.note", id)
		w.WriteHeader(http.StatusNoContent)
	})

	srv := &http.Server{
		Addr:         cfg.HTTPAddr,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 0,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		logger.Info("starting server", map[string]any{"addr": cfg.HTTPAddr})
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("alerts: server error: %v", err)
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

func scoreTS(m map[string]any) time.Time {
	if f, ok := m["ts"].(float64); ok && f > 0 {
		sec := int64(f)
		nsec := int64((f - float64(sec)) * 1e9)
		return time.Unix(sec, nsec).UTC()
	}
	if s, ok := m["ts"].(string); ok {
		if t, err := time.Parse(time.RFC3339Nano, s); err == nil {
			return t.UTC()
		}
	}
	return time.Now().UTC()
}

func scoreIdempotencyKey(symbol string, ts time.Time, score float64, severity, explanation string) string {
	raw := fmt.Sprintf("%s|%s|%.6f|%s|%s", symbol, ts.Format(time.RFC3339Nano), score, severity, explanation)
	h := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(h[:])
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

func sse(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	ch := make(chan string, 16)
	subs.Lock()
	subs.conns = append(subs.conns, ch)
	subs.Unlock()
	ctx := r.Context()
	defer func() {
		subs.Lock()
		for i, c := range subs.conns {
			if c == ch {
				subs.conns = append(subs.conns[:i], subs.conns[i+1:]...)
				break
			}
		}
		subs.Unlock()
		close(ch)
	}()
	heartbeat := time.NewTicker(15 * time.Second)
	defer heartbeat.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-heartbeat.C:
			if _, err := w.Write([]byte(": keepalive\n\n")); err != nil {
				return
			}
			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}
		case m := <-ch:
			if _, err := w.Write([]byte(m)); err != nil {
				return
			}
			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}
		}
	}
}

func broadcast(msg string) {
	subs.Lock()
	defer subs.Unlock()
	for _, c := range subs.conns {
		select {
		case c <- msg:
		default:
		}
	}
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
