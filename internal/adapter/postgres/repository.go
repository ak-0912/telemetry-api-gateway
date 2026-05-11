// Package postgres implements the telemetry.Repository port using PostgreSQL.
package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	apptelemetry "telemetry-api-gateway/internal/application/telemetry"
	"telemetry-api-gateway/internal/domain"
)

// Compile-time check: *TelemetryRepository satisfies apptelemetry.Repository.
var _ apptelemetry.Repository = (*TelemetryRepository)(nil)

// TelemetryRepository is a Postgres-backed implementation of apptelemetry.Repository.
type TelemetryRepository struct {
	pool *pgxpool.Pool
}

// NewTelemetryRepository creates a TelemetryRepository backed by the given connection pool.
func NewTelemetryRepository(pool *pgxpool.Pool) *TelemetryRepository {
	return &TelemetryRepository{pool: pool}
}

// ListDistinctGPUIDs returns every gpu_id present in the telemetry table, sorted ascending.
func (r *TelemetryRepository) ListDistinctGPUIDs(ctx context.Context) ([]string, error) {
	const q = `SELECT DISTINCT gpu_id FROM telemetry ORDER BY gpu_id`
	rows, err := r.pool.Query(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("list distinct gpu ids: %w", err)
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scan gpu id: %w", err)
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list distinct gpu ids: rows: %w", err)
	}
	return ids, nil
}

// ListByGPU returns telemetry rows for the given GPU, ordered by processed_at_unix_nano ASC.
// Both time bounds are optional (nil = unbounded) and inclusive.
func (r *TelemetryRepository) ListByGPU(ctx context.Context, gpuID string, startUnixNano, endUnixNano *int64) ([]domain.Telemetry, error) {
	const q = `
SELECT id, metric_name, gpu_id, device, uuid, model_name, host_name, value, labels_raw, processed_at_unix_nano, created_at
FROM telemetry
WHERE gpu_id = $1
  AND ($2::bigint IS NULL OR processed_at_unix_nano >= $2)
  AND ($3::bigint IS NULL OR processed_at_unix_nano <= $3)
ORDER BY processed_at_unix_nano ASC`
	rows, err := r.pool.Query(ctx, q, gpuID, startUnixNano, endUnixNano)
	if err != nil {
		return nil, fmt.Errorf("list telemetry by gpu: %w", err)
	}
	defer rows.Close()

	var out []domain.Telemetry
	for rows.Next() {
		var t domain.Telemetry
		if err := rows.Scan(
			&t.ID,
			&t.MetricName,
			&t.GPUID,
			&t.Device,
			&t.UUID,
			&t.ModelName,
			&t.HostName,
			&t.Value,
			&t.LabelsRaw,
			&t.ProcessedAtUnixNano,
			&t.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan telemetry: %w", err)
		}
		out = append(out, t)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list telemetry by gpu: rows: %w", err)
	}
	return out, nil
}
