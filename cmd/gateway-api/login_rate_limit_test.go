package main

import (
	"os"
	"sync"
	"testing"

	"sentinel/internal/rediskv"
	"sentinel/internal/testredis"
)

func TestLoginAllowed_LimitsByIPAndEmail(t *testing.T) {
	loginAttempts = sync.Map{}
	ip := "10.0.0.1"
	email := "user@example.com"
	for i := 0; i < 5; i++ {
		if !loginAllowed(ip, email) {
			t.Fatalf("attempt %d unexpectedly blocked", i+1)
		}
	}
	if loginAllowed(ip, email) {
		t.Fatalf("expected 6th attempt to be blocked")
	}
}

func TestLoginAllowed_SeparatesDifferentEmails(t *testing.T) {
	loginAttempts = sync.Map{}
	ip := "10.0.0.1"
	for i := 0; i < 5; i++ {
		_ = loginAllowed(ip, "first@example.com")
	}
	if !loginAllowed(ip, "second@example.com") {
		t.Fatalf("different email should have independent quota")
	}
}

func TestLoginAllowed_UsesRedisWhenEnabled(t *testing.T) {
	srv, err := testredis.Start()
	if err != nil {
		t.Fatalf("start redis stub: %v", err)
	}
	t.Cleanup(func() { _ = srv.Close() })

	oldAddr := os.Getenv("REDIS_ADDR")
	oldEnabled := os.Getenv("REDIS_ENABLED")
	t.Cleanup(func() {
		_ = os.Setenv("REDIS_ADDR", oldAddr)
		_ = os.Setenv("REDIS_ENABLED", oldEnabled)
		loginRedis = rediskv.NewFromEnv()
		loginAttempts = sync.Map{}
	})

	_ = os.Setenv("REDIS_ADDR", srv.Addr())
	_ = os.Setenv("REDIS_ENABLED", "true")
	loginRedis = rediskv.NewFromEnv()
	loginAttempts = sync.Map{}

	for i := 0; i < 5; i++ {
		if !loginAllowed("127.0.0.1", "redis@example.com") {
			t.Fatalf("attempt %d unexpectedly blocked", i+1)
		}
	}
	if loginAllowed("127.0.0.1", "redis@example.com") {
		t.Fatalf("expected redis-backed limiter to block 6th attempt")
	}
}
