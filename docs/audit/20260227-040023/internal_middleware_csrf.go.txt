package middleware

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"sync"
	"time"
)

type csrfEntry struct {
	Token  string
	Expiry time.Time
}

var csrfStore sync.Map

func NewCSRFToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func BindCSRF(subject, token string, ttl time.Duration) {
	csrfStore.Store(subject, csrfEntry{Token: token, Expiry: time.Now().Add(ttl)})
}

func ClearCSRF(subject string) { csrfStore.Delete(subject) }

func ValidCSRF(subject, token string) bool {
	v, ok := csrfStore.Load(subject)
	if !ok {
		return false
	}
	e := v.(csrfEntry)
	if time.Now().After(e.Expiry) || token == "" {
		return false
	}
	return e.Token == token
}

func RequireCSRFFunc(subjectFromReq func(*http.Request) (string, bool)) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodGet || r.Method == http.MethodHead || r.Method == http.MethodOptions {
				next.ServeHTTP(w, r)
				return
			}
			sub, ok := subjectFromReq(r)
			if !ok {
				http.Error(w, "forbidden", http.StatusForbidden)
				return
			}
			if !ValidCSRF(sub, r.Header.Get("X-CSRF-Token")) {
				http.Error(w, "forbidden", http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
