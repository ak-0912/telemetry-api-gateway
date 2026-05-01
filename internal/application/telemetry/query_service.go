package telemetry

import (
	"context"
	"time"

	"telemetry-api-gateway/internal/domain"
)

// QueryService implements read use cases for telemetry (application layer).
type QueryService struct {
	repo Repository
}

// NewQueryService wires the query service with its repository port.
func NewQueryService(repo Repository) *QueryService {
	return &QueryService{repo: repo}
}

// ListGPUsWithTelemetry returns distinct GPU IDs that have at least one telemetry row.
func (s *QueryService) ListGPUsWithTelemetry(ctx context.Context) ([]string, error) {
	return s.repo.ListDistinctGPUIDs(ctx)
}

// ListTelemetryForGPU returns telemetry for a GPU ordered by processed time, optionally filtered in time (inclusive).
func (s *QueryService) ListTelemetryForGPU(ctx context.Context, gpuID string, start, end *time.Time) ([]domain.Telemetry, error) {
	if start != nil && end != nil && start.After(*end) {
		return nil, domain.ErrInvalidTimeWindow
	}
	var startNano, endNano *int64
	if start != nil {
		n := start.UnixNano()
		startNano = &n
	}
	if end != nil {
		n := end.UnixNano()
		endNano = &n
	}
	return s.repo.ListByGPU(ctx, gpuID, startNano, endNano)
}
