package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	Port             string
	RedisAddr        string
	AuthUser         string
	AuthPass         string
	CacheTTL         time.Duration
}

func Load() *Config {
	ttlSec := 60
	if v := os.Getenv("CACHE_TTL_SECONDS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			ttlSec = n
		}
	}
	return &Config{
		Port:      getEnv("PORT", "8080"),
		RedisAddr: getEnv("REDIS_ADDR", "localhost:6379"),
		AuthUser:  getEnv("AUTH_USER", "admin"),
		AuthPass:  getEnv("AUTH_PASS", "secret"),
		CacheTTL:  time.Duration(ttlSec) * time.Second,
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
