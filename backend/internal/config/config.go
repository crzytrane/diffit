package config

import (
	"os"
)

type Config struct {
	DatabaseURL  string
	Port         string
	StoragePath  string
	AllowOrigins []string
}

func Load() *Config {
	return &Config{
		DatabaseURL: getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/diffit?sslmode=disable"),
		Port:        getEnv("PORT", "8080"),
		StoragePath: getEnv("STORAGE_PATH", "./storage"),
		AllowOrigins: []string{
			"http://localhost:5173",
			"https://diffit.markhamilton.dev",
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
