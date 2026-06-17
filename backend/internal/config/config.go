package config

import "os"

type Config struct {
	DatabaseURL      string
	JWTSecret        string
	JWTRefreshSecret string
	Port             string
	Env              string
	CORSOrigin       string
}

func Load() *Config {
	return &Config{
		DatabaseURL:      getEnv("DATABASE_URL", "postgres://kei:kei@localhost:5432/kei?sslmode=disable"),
		JWTSecret:        getEnv("JWT_SECRET", "change-me-in-production"),
		JWTRefreshSecret: getEnv("JWT_REFRESH_SECRET", "change-me-in-production-too"),
		Port:             getEnv("PORT", "8080"),
		Env:              getEnv("ENV", "development"),
		CORSOrigin:       getEnv("CORS_ORIGIN", "http://localhost:4200"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
