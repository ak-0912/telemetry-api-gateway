package telemetry

import (
	"context"
	"testing"
	"time"

	"telemetry-api-gateway/internal/domain"
)

type fakeRepo struct {
	ids   []string
	rows  []domain.Telemetry
	err   error
	gpuID string
	start *int64
	end   *int64
}

func (f *fakeRepo) ListDistinctGPUIDs(context.Context) ([]string, error) {
	return f.ids, f.err
}

func (f *fakeRepo) ListByGPU(_ context.Context, gpuID string, startUnixNano, endUnixNano *int64) ([]domain.Telemetry, error) {
	f.gpuID = gpuID
	f.start = startUnixNano
	f.end = endUnixNano
	return f.rows, f.err
}

func TestQueryService_ListTelemetryForGPU_InvalidWindow(t *testing.T) {
	s := NewQueryService(&fakeRepo{})
	start := time.Unix(10, 0)
	end := time.Unix(5, 0)
	_, err := s.ListTelemetryForGPU(context.Background(), "gpu-1", &start, &end)
	if err == nil {
		t.Fatal("expected error")
	}
	if err != domain.ErrInvalidTimeWindow {
		t.Fatalf("got %v want %v", err, domain.ErrInvalidTimeWindow)
	}
}

func TestQueryService_ListGPUsWithTelemetry_OK(t *testing.T) {
	repo := &fakeRepo{ids: []string{"x", "y"}}
	s := NewQueryService(repo)
	got, err := s.ListGPUsWithTelemetry(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0] != "x" || got[1] != "y" {
		t.Fatalf("got %#v", got)
	}
}

func TestQueryService_ListTelemetryForGPU_PassesUnixNanoToRepo(t *testing.T) {
	repo := &fakeRepo{rows: []domain.Telemetry{{ID: 7, GPUID: "g"}}}
	s := NewQueryService(repo)
	start := time.Unix(0, 1000).UTC()
	end := time.Unix(0, 2000).UTC()
	got, err := s.ListTelemetryForGPU(context.Background(), "gpu-9", &start, &end)
	if err != nil {
		t.Fatal(err)
	}
	if repo.gpuID != "gpu-9" {
		t.Fatalf("gpu id %q", repo.gpuID)
	}
	if repo.start == nil || *repo.start != start.UnixNano() || repo.end == nil || *repo.end != end.UnixNano() {
		t.Fatalf("start/end nano %v %v", repo.start, repo.end)
	}
	if len(got) != 1 || got[0].ID != 7 {
		t.Fatalf("got %#v", got)
	}
}

func TestQueryService_ListTelemetryForGPU_NoTimeFilter(t *testing.T) {
	repo := &fakeRepo{}
	s := NewQueryService(repo)
	_, err := s.ListTelemetryForGPU(context.Background(), "gpu-1", nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if repo.start != nil || repo.end != nil {
		t.Fatalf("expected nil bounds, got %v %v", repo.start, repo.end)
	}
}
