package config

import "os"

// Config holds process-level settings (environment).
type Config struct {
	DatabaseURL string
	HTTPAddr    string
}

// Load reads configuration from the environment.
func Load() Config {
	db := os.Getenv("DATABASE_URL")
	if db == "" {
		db = "postgres://postgres:postgres@localhost:5433/telemetry?sslmode=disable"
	}
	addr := os.Getenv("HTTP_ADDR")
	if addr == "" {
		addr = ":8080"
	}
	return Config{DatabaseURL: db, HTTPAddr: addr}
}
