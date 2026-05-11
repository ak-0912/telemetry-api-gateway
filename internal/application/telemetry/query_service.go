package telemetry

import (
	"context"
	"time"

	"telemetry-api-gateway/internal/domain"
)

// QueryService implements read-only use cases for GPU telemetry.
type QueryService struct {
	repo Repository
}

// NewQueryService creates a QueryService backed by the given Repository.
func NewQueryService(repo Repository) *QueryService {
	return &QueryService{repo: repo}
}

// ListGPUsWithTelemetry returns distinct GPU IDs that have at least one telemetry row.
func (s *QueryService) ListGPUsWithTelemetry(ctx context.Context) ([]string, error) {
	return s.repo.ListDistinctGPUIDs(ctx)
}

// ListTelemetryForGPU returns telemetry for a GPU, optionally filtered by an
// inclusive [start, end] time window. Returns domain.ErrInvalidTimeWindow when
// start is strictly after end.
func (s *QueryService) ListTelemetryForGPU(ctx context.Context, gpuID string, start, end *time.Time) ([]domain.Telemetry, error) {
	if start != nil && end != nil && start.After(*end) {
		return nil, domain.ErrInvalidTimeWindow
	}
	return s.repo.ListByGPU(ctx, gpuID, unixNanoPtr(start), unixNanoPtr(end))
}

// unixNanoPtr converts a *time.Time to a *int64 (nanoseconds since epoch).
// Returns nil when t is nil, which the repository interprets as "unbounded".
func unixNanoPtr(t *time.Time) *int64 {
	if t == nil {
		return nil
	}
	n := t.UnixNano()
	return &n
}
