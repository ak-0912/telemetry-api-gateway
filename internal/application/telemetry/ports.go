package telemetry

import (
	"context"

	"telemetry-api-gateway/internal/domain"
)

// Repository is a driven port: persistence for telemetry queries.
type Repository interface {
	ListDistinctGPUIDs(ctx context.Context) ([]string, error)
	ListByGPU(ctx context.Context, gpuID string, startUnixNano, endUnixNano *int64) ([]domain.Telemetry, error)
}
