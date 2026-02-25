package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"log"
	"math"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"sentinel/internal/config"
	"sentinel/internal/db"
	"sentinel/internal/kafka"
	"sentinel/internal/logging"

	"github.com/twmb/franz-go/pkg/kgo"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	logger := logging.New("features")
	cfg := config.Load("features")
	if err := config.Validate(cfg); err != nil {
		log.Fatalf("features: invalid config: %v", err)
	}

	pool, err := db.Connect(ctx, cfg.PostgresURL)
	if err != nil {
		log.Fatalf("features: failed to connect to database: %v", err)
	}
	defer pool.Close()

	c, err := kafka.New(cfg.KafkaBrokers, "features", "derived.candles")
	if err != nil {
		log.Fatalf("features: failed to create kafka consumer: %v", err)
	}
	defer c.Close()

	p, err := kafka.New(cfg.KafkaBrokers, "")
	if err != nil {
		log.Fatalf("features: failed to create kafka producer: %v", err)
	}
	defer p.Close()

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("ok")) })
	mux.HandleFunc("/readyz", func(w http.ResponseWriter, r *http.Request) {
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
	srv := &http.Server{Addr: cfg.HTTPAddr, Handler: mux, ReadTimeout: 10 * time.Second, WriteTimeout: 10 * time.Second}
	go func() {
		logger.Info("starting health server", map[string]any{"addr": cfg.HTTPAddr})
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("features: health server error: %v", err)
		}
	}()

	go func() {
		for {
			f := c.PollFetches(ctx)
			if ctx.Err() != nil {
				return
			}
			f.EachRecord(func(rec *kgo.Record) {
				var m map[string]any
				if err := json.Unmarshal(rec.Value, &m); err != nil {
					logger.Error("unmarshal failed", map[string]any{"error": err.Error()})
					return
				}
				symbol, _ := m["symbol"].(string)
				price, _ := m["price"].(float64)
				vol, _ := m["volume"].(float64)
				if symbol == "" {
					logger.Error("missing symbol in message", nil)
					return
				}
				if price <= 0 {
					logger.Error("invalid price", map[string]any{"price": price, "symbol": symbol})
					return
				}
				ret := math.Log(price / 100.0)
				payload := map[string]any{"log_return": ret, "volume_z": vol / 100.0}
				pb, err := json.Marshal(payload)
				if err != nil {
					logger.Error("marshal payload failed", map[string]any{"error": err.Error()})
					return
				}
				h := sha256.Sum256(pb)
				hash := hex.EncodeToString(h[:])
				id := symbol + hash[:16]
				if _, err := pool.Exec(ctx, `INSERT INTO features(idempotency_key,symbol,ts,feature_hash,payload) VALUES($1,$2,$3,$4,$5) ON CONFLICT DO NOTHING`, id, symbol, time.Now().UTC(), hash, pb); err != nil {
					logger.Error("insert feature failed", map[string]any{"error": err.Error()})
				}
				if err := kafka.ProduceJSON(ctx, p, "derived.features", symbol, map[string]any{"symbol": symbol, "feature_hash": hash, "payload": payload}); err != nil {
					logger.Error("produce feature msg failed", map[string]any{"error": err.Error()})
				}
				if err := kafka.ProduceJSON(ctx, p, "dq.metrics", symbol, map[string]any{"symbol": symbol, "missing": 0, "stale": 0}); err != nil {
					logger.Error("produce dq metric failed", map[string]any{"error": err.Error()})
				}
			})
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
