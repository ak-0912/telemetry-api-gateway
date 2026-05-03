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
		db = defaultDatabaseURL()
	}
	addr := os.Getenv("HTTP_ADDR")
	if addr == "" {
		// 0.0.0.0 so Docker-published ports (e.g. 8081:8080) reach the server; :8080 alone can be IPv6-only on some hosts.
		addr = "0.0.0.0:8080"
	}
	return Config{DatabaseURL: db, HTTPAddr: addr}
}
