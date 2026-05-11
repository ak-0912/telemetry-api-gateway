package postgres

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed migrations/001_telemetry.sql
var schemaSQL string

// Migrate applies the embedded DDL to ensure the telemetry schema exists.
// The SQL is idempotent (CREATE TABLE IF NOT EXISTS).
func Migrate(ctx context.Context, pool *pgxpool.Pool) error {
	if _, err := pool.Exec(ctx, schemaSQL); err != nil {
		return fmt.Errorf("apply schema migration: %w", err)
	}
	return nil
}
