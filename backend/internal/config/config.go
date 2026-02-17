package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Port               string
	DatabaseURL        string
	WeatherAPIURL      string
	WeatherAPIToken    string
	WeatherCacheMin    time.Duration
	DefaultLocationLat float64
	DefaultLocationLng float64
}

func Load() (Config, error) {
	_ = godotenv.Load()
	cacheMin := getEnvInt("WEATHER_CACHE_MINUTES", 20)
	lat := getEnvFloat("DEFAULT_LOCATION_LAT", 45.092)
	lng := getEnvFloat("DEFAULT_LOCATION_LNG", 37.268)

	cfg := Config{
		Port:               getEnv("PORT", "8080"),
		DatabaseURL:        os.Getenv("DATABASE_URL"),
		WeatherAPIURL:      getEnv("WEATHER_API_URL", "https://api.open-meteo.com/v1/forecast"),
		WeatherAPIToken:    os.Getenv("WEATHER_API_TOKEN"),
		WeatherCacheMin:    time.Duration(cacheMin) * time.Minute,
		DefaultLocationLat: lat,
		DefaultLocationLng: lng,
	}

	if cfg.DatabaseURL == "" {
		return Config{}, fmt.Errorf("DATABASE_URL is required")
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
