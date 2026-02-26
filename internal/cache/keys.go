package cache

import "fmt"

func QueueKey(role, region, country, industry, timeWindow string) string {
	return fmt.Sprintf("queue:v1:%s:%s:%s:%s:%s", role, region, country, industry, timeWindow)
}

func CommandCenterKey(role, region, industry, timeWindow string) string {
	return fmt.Sprintf("cc:v1:%s:%s:%s:%s", role, region, industry, timeWindow)
}

func TrustKey(region, industry, timeWindow string) string {
	return fmt.Sprintf("trust:v1:%s:%s:%s", region, industry, timeWindow)
}

func WorldMapKey(region, industry, timeWindow string) string {
	return fmt.Sprintf("worldmap:v1:%s:%s:%s", region, industry, timeWindow)
}

func ExecutiveKey(region, industry, timeWindow string) string {
	return fmt.Sprintf("exec:v1:%s:%s:%s", region, industry, timeWindow)
}
