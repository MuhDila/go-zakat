package config

import (
	"log"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type AppConfig struct {
	AppPort string

	DatabaseURL string

	JWTAccessSecret  string
	JWTRefreshSecret string

	JWTAccessTTL  time.Duration
	JWTRefreshTTL time.Duration

	GoogleClientID     string
	GoogleClientSecret string
	GoogleRedirectURL  string

	FrontendURL string

	CORSAllowedOrigins []string
}

func Load() *AppConfig {
	_ = godotenv.Load()

	cfg := &AppConfig{
		AppPort: getEnv("APP_PORT", "8080"),

		DatabaseURL: getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/auth_db?sslmode=disable"),

		JWTAccessSecret:  mustGet("JWT_ACCESS_SECRET"),
		JWTRefreshSecret: mustGet("JWT_REFRESH_SECRET"),

		GoogleClientID:     mustGet("GOOGLE_CLIENT_ID"),
		GoogleClientSecret: mustGet("GOOGLE_CLIENT_SECRET"),
		GoogleRedirectURL:  mustGet("GOOGLE_REDIRECT_URL"),

		FrontendURL: getEnv("FRONTEND_URL", "http://localhost:3000"),

		CORSAllowedOrigins: split(getEnv("CORS_ALLOWED_ORIGINS", "")),
	}

	// ambil TTL dari env
	cfg.JWTAccessTTL = parseTTL(getEnv("JWT_ACCESS_TTL", "15m"))    // default 15m
	cfg.JWTRefreshTTL = parseTTL(getEnv("JWT_REFRESH_TTL", "168h")) // default 7d

	return cfg
}

func mustGet(key string) string {
	val := os.Getenv(key)
	if val == "" {
		log.Fatalf("ENV %s wajib diisi", key)
	}
	return val
}

func getEnv(key, def string) string {
	val := os.Getenv(key)
	if val == "" {
		return def
	}
	return val
}

func split(s string) []string {
	if s == "" {
		return nil
	}
	var parts []string
	for _, p := range strings.Split(s, ",") {
		p = strings.TrimSpace(p)
		if p != "" {
			parts = append(parts, p)
		}
	}
	return parts
}

// parseTTL mendukung format time seperti:
// 15m, 24h, 7d, 30s, dll
func parseTTL(s string) time.Duration {
	if strings.HasSuffix(s, "d") {
		// convert "7d" → 7 * 24h
		days := strings.TrimSuffix(s, "d")
		n, err := time.ParseDuration(days + "h")
		if err != nil {
			log.Fatalf("TTL %s tidak valid: %v", s, err)
		}
		return n * 24
	}

	d, err := time.ParseDuration(s)
	if err != nil {
		log.Fatalf("TTL %s tidak valid: %v", s, err)
	}
	return d
}
