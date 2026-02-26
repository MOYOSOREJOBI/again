package cache

import "testing"

func TestKeysIncludeLocaleDimension(t *testing.T) {
	a := QueueKey("viewer", "NA", "US", "Tech", "Software", "XNAS", "AAPL", "en", "24h")
	b := QueueKey("viewer", "NA", "US", "Tech", "Software", "XNAS", "AAPL", "fr", "24h")
	if a == b {
		t.Fatal("expected different keys by locale")
	}
}
