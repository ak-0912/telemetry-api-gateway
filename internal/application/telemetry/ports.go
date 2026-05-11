// Package telemetry implements the application (use-case) layer for GPU telemetry queries.
// It defines the Repository port (driven adapter interface) and the QueryService that
// orchestrates reads.
package telemetry

import (
	"context"

	"telemetry-api-gateway/internal/domain"
)

// Repository is the driven port for telemetry persistence.
// Implementations live in adapter packages (e.g. adapter/postgres).
type Repository interface {
	// ListDistinctGPUIDs returns every gpu_id that has at least one telemetry row.
	ListDistinctGPUIDs(ctx context.Context) ([]string, error)

	// ListByGPU returns telemetry for a GPU, ordered by processed_at_unix_nano ASC.
	// Both time bounds are optional (nil = unbounded) and inclusive.
	ListByGPU(ctx context.Context, gpuID string, startUnixNano, endUnixNano *int64) ([]domain.Telemetry, error)
}
