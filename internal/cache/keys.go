package cache

import "fmt"

func QueueKey(role, region, country, sector, industry, venue, symbol, window string) string {
	return fmt.Sprintf("queue:v2:%s:%s:%s:%s:%s:%s:%s:%s", role, region, country, sector, industry, venue, symbol, window)
}

func CommandCenterKey(role, region, country, sector, industry, venue, symbol, window string) string {
	return fmt.Sprintf("cc:v2:%s:%s:%s:%s:%s:%s:%s:%s", role, region, country, sector, industry, venue, symbol, window)
}

func TrustKey(region, country, sector, industry, venue, symbol, window string) string {
	return fmt.Sprintf("trust:v2:%s:%s:%s:%s:%s:%s:%s", region, country, sector, industry, venue, symbol, window)
}

func WorldMapKey(region, country, sector, industry, venue, symbol, window string) string {
	return fmt.Sprintf("worldmap:v2:%s:%s:%s:%s:%s:%s:%s", region, country, sector, industry, venue, symbol, window)
}

func ExecutiveKey(region, country, sector, industry, venue, symbol, window string) string {
	return fmt.Sprintf("exec:v2:%s:%s:%s:%s:%s:%s:%s", region, country, sector, industry, venue, symbol, window)
}
