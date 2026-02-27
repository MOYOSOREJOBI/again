package cache

import (
	"context"
	"encoding/json"
	"sync"
	"time"
)

type entry struct {
	b   []byte
	exp time.Time
}

var mem sync.Map

func GetOrLoadJSON[T any](_ context.Context, key string, ttl time.Duration, loader func() (T, error)) (T, error) {
	var zero T
	if v, ok := mem.Load(key); ok {
		e := v.(entry)
		if time.Now().Before(e.exp) {
			var out T
			if json.Unmarshal(e.b, &out) == nil {
				return out, nil
			}
		}
	}
	out, err := loader()
	if err != nil {
		return zero, err
	}
	if b, err := json.Marshal(out); err == nil {
		mem.Store(key, entry{b: b, exp: time.Now().Add(ttl)})
	}
	return out, nil
}
