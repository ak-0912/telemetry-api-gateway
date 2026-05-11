// Package config loads process-level configuration from environment variables.
// When DATABASE_URL is unset, it probes candidate hosts to auto-discover Postgres.
package config

import "os"

// Config holds all process-level settings.
type Config struct {
	DatabaseURL string // Postgres connection string.
	HTTPAddr    string // TCP address the HTTP server binds to (host:port).
}

// Load reads configuration from the environment. Unset variables fall back to
// sensible defaults (0.0.0.0:8080 for HTTP, auto-discovered Postgres URL).
func Load() Config {
	db := os.Getenv("DATABASE_URL")
	if db == "" {
		db = defaultDatabaseURL()
	}

	addr := os.Getenv("HTTP_ADDR")
	if addr == "" {
		addr = "0.0.0.0:8080"
	}
	return Config{DatabaseURL: db, HTTPAddr: addr}
}
