package config

import (
	"os"
	"strconv"
	"time"
)

// Config holds all runtime configuration derived from environment variables.
type Config struct {
	Port               string
	ReadTimeout        time.Duration
	WriteTimeout       time.Duration
	FetchTimeout       time.Duration
	MaxConcurrency     int
	MaxURLsPerRequest  int
}

// Load reads configuration from environment variables, applying defaults where needed.
func Load() Config {
	return Config{
		Port:              getEnv("PORT", "8080"),
		ReadTimeout:       getDuration("READ_TIMEOUT", 5*time.Second),
		WriteTimeout:      getDuration("WRITE_TIMEOUT", 60*time.Second),
		FetchTimeout:      getDuration("FETCH_TIMEOUT", 15*time.Second),
		MaxConcurrency:    getInt("MAX_CONCURRENCY", 5),
		MaxURLsPerRequest: getInt("MAX_URLS_PER_REQUEST", 10),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getDuration(key string, fallback time.Duration) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return fallback
	}
	return d
}

func getInt(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return n
}
