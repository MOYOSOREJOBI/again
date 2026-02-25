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

	"sentinel/internal/audit"
	"sentinel/internal/auth"
	"sentinel/internal/config"
	"sentinel/internal/db"
	"sentinel/internal/httpx"
	"sentinel/internal/logging"

	"github.com/go-chi/chi/v5"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	logger := logging.New("gateway-api")
	cfg := config.Load("gateway-api")
	if err := config.Validate(cfg); err != nil {
		log.Fatalf("gateway-api: invalid config: %v", err)
	}

	dbpool, err := db.Connect(ctx, cfg.PostgresURL)
	if err != nil {
		log.Fatalf("gateway-api: failed to connect to database: %v", err)
	}
	defer dbpool.Close()

	priv, err := auth.ReadPrivate(cfg.JWTPrivateKey)
	if err != nil {
		log.Fatalf("gateway-api: failed to read JWT private key: %v", err)
	}
	pub, err := auth.ReadPublic(cfg.JWTPublicKey)
	if err != nil {
		log.Fatalf("gateway-api: failed to read JWT public key: %v", err)
	}

	r := chi.NewRouter()
	r.Use(corsMiddleware)
	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("ok")) })
	r.Get("/readyz", func(w http.ResponseWriter, r *http.Request) {
		if err := dbpool.Ping(ctx); err != nil {
			http.Error(w, "not ready", http.StatusServiceUnavailable)
			return
		}
		w.Write([]byte("ok"))
	})

	r.Post("/auth/login", func(w http.ResponseWriter, r *http.Request) {
		var in struct{ Email, Password string }
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&in); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		if in.Email == "" || in.Password == "" {
			http.Error(w, "email and password required", http.StatusBadRequest)
			return
		}
		var hash, role string
		err := dbpool.QueryRow(ctx, `SELECT password_hash,role FROM users WHERE email=$1`, in.Email).Scan(&hash, &role)
		if err != nil || bcrypt.CompareHashAndPassword([]byte(hash), []byte(in.Password)) != nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		tok, err := auth.Sign(in.Email, role, priv)
		if err != nil {
			logger.Error("failed to sign token", map[string]any{"error": err.Error()})
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		http.SetCookie(w, &http.Cookie{
			Name:     "sentinel_token",
			Value:    tok,
			HttpOnly: true,
			Secure:   false,
			SameSite: http.SameSiteLaxMode,
			Path:     "/",
			Expires:  time.Now().Add(24 * time.Hour),
		})
		if err := audit.Append(ctx, dbpool, in.Email, "login", "success"); err != nil {
			logger.Error("audit append failed", map[string]any{"error": err.Error()})
		}
		httpx.JSON(w, http.StatusOK, map[string]any{"token": tok, "role": role})
	})

	r.Post("/auth/logout", func(w http.ResponseWriter, r *http.Request) {
		http.SetCookie(w, &http.Cookie{
			Name:     "sentinel_token",
			Value:    "",
			HttpOnly: true,
			Secure:   false,
			SameSite: http.SameSiteLaxMode,
			Path:     "/",
			MaxAge:   -1,
		})
		w.WriteHeader(http.StatusNoContent)
	})

	r.Get("/me", func(w http.ResponseWriter, r *http.Request) {
		claims, ok := authn(r, pub)
		if !ok {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		httpx.JSON(w, http.StatusOK, map[string]string{"user": claims.Subject, "role": claims.Role})
	})

	r.Get("/audit/verify", func(w http.ResponseWriter, r *http.Request) {
		claims, ok := authn(r, pub)
		if !ok || claims.Role != "admin" {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		okv, msg, err := audit.Verify(ctx, dbpool)
		if err != nil {
			logger.Error("audit verify failed", map[string]any{"error": err.Error()})
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		httpx.JSON(w, http.StatusOK, map[string]any{"ok": okv, "message": msg})
	})

	srv := &http.Server{
		Addr:         cfg.HTTPAddr,
		Handler:      withCorrelation(r),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		logger.Info("starting server", map[string]any{"addr": cfg.HTTPAddr})
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("gateway-api: server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("shutting down", nil)

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("gateway-api: shutdown error: %v", err)
	}
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
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Correlation-ID")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func withCorrelation(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cid := r.Header.Get("X-Correlation-ID")
		if cid == "" {
			cid = time.Now().UTC().Format("20060102150405.000")
		}
		w.Header().Set("X-Correlation-ID", cid)
		next.ServeHTTP(w, r)
	})
}
