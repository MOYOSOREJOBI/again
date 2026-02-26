package cache

import (
	"context"
	"strings"
)

func InvalidateByPrefixes(_ context.Context, prefixes ...string) {
	mem.Range(func(key, _ any) bool {
		ks, ok := key.(string)
		if !ok {
			return true
		}
		for _, p := range prefixes {
			if strings.HasPrefix(ks, p) {
				mem.Delete(ks)
				break
			}
		}
		return true
	})
}
