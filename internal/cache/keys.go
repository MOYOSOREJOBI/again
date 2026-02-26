package cache

import "fmt"

func QueueKey(role, region, country, sector, industry, venue, symbol, locale, window string) string {
	return fmt.Sprintf("queue:v3:%s:%s:%s:%s:%s:%s:%s:%s:%s", role, region, country, sector, industry, venue, symbol, locale, window)
}

func CommandCenterKey(role, region, country, sector, industry, venue, symbol, locale, window string) string {
	return fmt.Sprintf("cc:v3:%s:%s:%s:%s:%s:%s:%s:%s:%s", role, region, country, sector, industry, venue, symbol, locale, window)
}

func TrustKey(region, country, sector, industry, venue, symbol, locale, window string) string {
	return fmt.Sprintf("trust:v3:%s:%s:%s:%s:%s:%s:%s:%s", region, country, sector, industry, venue, symbol, locale, window)
}

func WorldMapKey(region, country, sector, industry, venue, symbol, locale, window string) string {
	return fmt.Sprintf("worldmap:v3:%s:%s:%s:%s:%s:%s:%s:%s", region, country, sector, industry, venue, symbol, locale, window)
}

func ExecutiveKey(region, country, sector, industry, venue, symbol, locale, window string) string {
	return fmt.Sprintf("exec:v3:%s:%s:%s:%s:%s:%s:%s:%s", region, country, sector, industry, venue, symbol, locale, window)
}
