package main

import (
	"sync"
	"testing"
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
