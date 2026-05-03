package ginhttp

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"telemetry-api-gateway/internal/domain"
)

type stubTelemetryQuery struct {
	gpuIDs    []string
	listErr   error
	telemetry []domain.Telemetry
	telErr    error
	lastGPU   string
	lastStart *time.Time
	lastEnd   *time.Time
}

func (s *stubTelemetryQuery) ListGPUsWithTelemetry(context.Context) ([]string, error) {
	return s.gpuIDs, s.listErr
}

func (s *stubTelemetryQuery) ListTelemetryForGPU(_ context.Context, gpuID string, start, end *time.Time) ([]domain.Telemetry, error) {
	s.lastGPU = gpuID
	s.lastStart = start
	s.lastEnd = end
	return s.telemetry, s.telErr
}

func newTestRouter(h *TelemetryHandler) http.Handler {
	return NewRouter(h)
}

func TestTelemetryHandler_ListGPUs_OK(t *testing.T) {
	stub := &stubTelemetryQuery{gpuIDs: []string{"gpu-a", "gpu-b"}}
	h := NewTelemetryHandler(stub)
	srv := newTestRouter(h)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/gpus", nil)
	srv.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status %d, body %s", rr.Code, rr.Body.String())
	}
	var got listGpusResponse
	if err := json.NewDecoder(rr.Body).Decode(&got); err != nil {
		t.Fatal(err)
	}
	if len(got.GPUs) != 2 || got.GPUs[0] != "gpu-a" || got.GPUs[1] != "gpu-b" {
		t.Fatalf("gpus: %#v", got.GPUs)
	}
}

func TestTelemetryHandler_ListGPUs_ServiceError(t *testing.T) {
	stub := &stubTelemetryQuery{listErr: errors.New("db down")}
	h := NewTelemetryHandler(stub)
	srv := newTestRouter(h)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/gpus", nil)
	srv.ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("status %d", rr.Code)
	}
}

func TestTelemetryHandler_GetGPUTelemetry_OK(t *testing.T) {
	ts := time.Date(2024, 3, 1, 12, 0, 0, 0, time.UTC)
	stub := &stubTelemetryQuery{
		telemetry: []domain.Telemetry{
			{
				ID: 1, MetricName: "util", GPUID: "gpu-a", Device: "d", UUID: "u",
				ModelName: "m", HostName: "h", Value: 0.5, LabelsRaw: "{}",
				ProcessedAtUnixNano: ts.UnixNano(), CreatedAt: ts,
			},
		},
	}
	h := NewTelemetryHandler(stub)
	srv := newTestRouter(h)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/gpus/gpu-a/telemetry", nil)
	srv.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status %d, body %s", rr.Code, rr.Body.String())
	}
	if stub.lastGPU != "gpu-a" {
		t.Fatalf("gpu id: %q", stub.lastGPU)
	}
	var body telemetryListResponse
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if len(body.Entries) != 1 || body.Entries[0].ID != 1 || body.Entries[0].MetricName != "util" {
		t.Fatalf("entries: %#v", body.Entries)
	}
}

func TestTelemetryHandler_GetGPUTelemetry_TimeFiltersPassed(t *testing.T) {
	stub := &stubTelemetryQuery{}
	h := NewTelemetryHandler(stub)
	srv := newTestRouter(h)

	rr := httptest.NewRecorder()
	u := "/api/v1/gpus/gpu-a/telemetry?start_time=2026-05-01T14:15:00Z&end_time=2026-05-03T06:30:00Z"
	req := httptest.NewRequest(http.MethodGet, u, nil)
	srv.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status %d, body %s", rr.Code, rr.Body.String())
	}
	if stub.lastStart == nil || stub.lastEnd == nil {
		t.Fatal("expected start/end to be set")
	}
	if stub.lastStart.Format(time.RFC3339) != "2026-05-01T14:15:00Z" {
		t.Fatalf("start: %v", stub.lastStart)
	}
	if stub.lastEnd.Format(time.RFC3339) != "2026-05-03T06:30:00Z" {
		t.Fatalf("end: %v", stub.lastEnd)
	}
}

func TestTelemetryHandler_GetGPUTelemetry_InvalidStartTime(t *testing.T) {
	stub := &stubTelemetryQuery{}
	h := NewTelemetryHandler(stub)
	srv := newTestRouter(h)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/gpus/gpu-a/telemetry?start_time=not-a-date", nil)
	srv.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status %d", rr.Code)
	}
}

func TestTelemetryHandler_GetGPUTelemetry_InvalidTimeWindowFromService(t *testing.T) {
	stub := &stubTelemetryQuery{telErr: domain.ErrInvalidTimeWindow}
	h := NewTelemetryHandler(stub)
	srv := newTestRouter(h)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/gpus/gpu-a/telemetry", nil)
	srv.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status %d", rr.Code)
	}
}

func TestTelemetryHandler_GetGPUTelemetry_ServiceError(t *testing.T) {
	stub := &stubTelemetryQuery{telErr: errors.New("db")}
	h := NewTelemetryHandler(stub)
	srv := newTestRouter(h)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/gpus/gpu-a/telemetry", nil)
	srv.ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("status %d", rr.Code)
	}
}
