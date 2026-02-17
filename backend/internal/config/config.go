package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Port               string
	WeatherAPIURL      string
	WeatherCacheMin    time.Duration
	DefaultLocationLat float64
	DefaultLocationLng float64
}

func Load() (Config, error) {
	cacheMin := getEnvInt("WEATHER_CACHE_MINUTES", 20)
	cfg := Config{
		Port:               getEnv("PORT", "8080"),
		WeatherAPIURL:      getEnv("WEATHER_API_URL", "https://api.open-meteo.com/v1/forecast"),
		WeatherCacheMin:    time.Duration(cacheMin) * time.Minute,
		DefaultLocationLat: getEnvFloat("DEFAULT_LOCATION_LAT", 45.092),
		DefaultLocationLng: getEnvFloat("DEFAULT_LOCATION_LNG", 37.268),
	}
	if cfg.Port == "" {
		return Config{}, fmt.Errorf("PORT is required")
	}
	return cfg, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}
func getEnvFloat(key string, fallback float64) float64 {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.ParseFloat(v, 64); err == nil {
			return n
		}
	}
	return fallback
}
