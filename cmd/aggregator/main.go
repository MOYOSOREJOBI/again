package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"log"
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

type Tick struct {
	EventID   string    `json:"event_id"`
	Symbol    string    `json:"symbol"`
	Price     float64   `json:"price"`
	Volume    float64   `json:"volume"`
	EventTime time.Time `json:"event_time"`
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	logger := logging.New("aggregator")
	cfg := config.Load("aggregator")
	if err := config.Validate(cfg); err != nil {
		log.Fatalf("aggregator: invalid config: %v", err)
	}

	pool, err := db.Connect(ctx, cfg.PostgresURL)
	if err != nil {
		log.Fatalf("aggregator: failed to connect to database: %v", err)
	}
	defer pool.Close()

	client, err := kafka.New(cfg.KafkaBrokers, "aggregator", "raw.ticks")
	if err != nil {
		log.Fatalf("aggregator: failed to create kafka consumer: %v", err)
	}
	defer client.Close()

	prod, err := kafka.New(cfg.KafkaBrokers, "")
	if err != nil {
		log.Fatalf("aggregator: failed to create kafka producer: %v", err)
	}
	defer prod.Close()

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("ok")) })
	mux.HandleFunc("/readyz", func(w http.ResponseWriter, r *http.Request) {
		probeCtx, probeCancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer probeCancel()
		if err := pool.Ping(probeCtx); err != nil {
			http.Error(w, "database not ready", http.StatusServiceUnavailable)
			return
		}
		if err := client.Ping(probeCtx); err != nil {
			http.Error(w, "kafka consumer not ready", http.StatusServiceUnavailable)
			return
		}
		if err := prod.Ping(probeCtx); err != nil {
			http.Error(w, "kafka producer not ready", http.StatusServiceUnavailable)
			return
		}
		w.Write([]byte("ok"))
	})
	srv := &http.Server{Addr: cfg.HTTPAddr, Handler: mux, ReadTimeout: 10 * time.Second, WriteTimeout: 10 * time.Second}
	go func() {
		logger.Info("starting health server", map[string]any{"addr": cfg.HTTPAddr})
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("aggregator: health server error: %v", err)
		}
	}()

	go func() {
		for {
			fetches := client.PollFetches(ctx)
			if ctx.Err() != nil {
				return
			}
			fetches.EachRecord(func(rec *kgo.Record) {
				var t Tick
				if err := json.Unmarshal(rec.Value, &t); err != nil {
					logger.Error("unmarshal tick failed", map[string]any{"error": err.Error()})
					return
				}
				if t.Symbol == "" {
					logger.Error("tick missing symbol", nil)
					return
				}
				if t.EventTime.IsZero() {
					t.EventTime = time.Now().UTC()
				}
				tx, err := pool.Begin(ctx)
				if err != nil {
					logger.Error("begin tx failed", map[string]any{"error": err.Error()})
					return
				}
				defer tx.Rollback(ctx)

				if _, err := tx.Exec(ctx, `INSERT INTO raw_ticks(event_id,symbol,price,volume,event_time) VALUES($1,$2,$3,$4,$5) ON CONFLICT DO NOTHING`, t.EventID, t.Symbol, t.Price, t.Volume, t.EventTime); err != nil {
					logger.Error("insert raw_tick failed", map[string]any{"error": err.Error()})
					return
				}
				bucket := t.EventTime.Truncate(time.Second)
				h := sha256.Sum256([]byte(t.Symbol + bucket.String() + "1s"))
				key := hex.EncodeToString(h[:])
				if _, err := tx.Exec(ctx, `INSERT INTO candles(idempotency_key,symbol,bucket,interval,open,high,low,close,volume) VALUES($1,$2,$3,'1s',$4,$4,$4,$4,$5) ON CONFLICT DO NOTHING`, key, t.Symbol, bucket, t.Price, t.Volume); err != nil {
					logger.Error("insert candle 1s failed", map[string]any{"error": err.Error()})
					return
				}
				if _, err := tx.Exec(ctx, `INSERT INTO candles(idempotency_key,symbol,bucket,interval,open,high,low,close,volume) VALUES($1,$2,$3,'1m',$4,$4,$4,$4,$5) ON CONFLICT DO NOTHING`, key+"m", t.Symbol, t.EventTime.Truncate(time.Minute), t.Price, t.Volume); err != nil {
					logger.Error("insert candle 1m failed", map[string]any{"error": err.Error()})
					return
				}
				if err := tx.Commit(ctx); err != nil {
					logger.Error("commit failed", map[string]any{"error": err.Error()})
					return
				}
				if err := kafka.ProduceJSON(ctx, prod, "derived.candles", t.Symbol, map[string]any{"symbol": t.Symbol, "bucket": bucket, "price": t.Price, "volume": t.Volume}); err != nil {
					logger.Error("produce candle msg failed", map[string]any{"error": err.Error()})
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
