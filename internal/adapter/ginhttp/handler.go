// Package ginhttp is the driving HTTP adapter. It translates Gin requests into
// application-layer calls and renders JSON responses.
package ginhttp

import (
	"context"
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"telemetry-api-gateway/internal/domain"
)

// ---------------------------------------------------------------------------
// Driving port (application interface consumed by this adapter)
// ---------------------------------------------------------------------------

// TelemetryQuery is the application-layer interface that the HTTP handlers
// depend on. It is satisfied by *telemetry.QueryService; defining it here
// keeps the adapter decoupled and enables unit-test stubs.
type TelemetryQuery interface {
	ListGPUsWithTelemetry(ctx context.Context) ([]string, error)
	ListTelemetryForGPU(ctx context.Context, gpuID string, start, end *time.Time) ([]domain.Telemetry, error)
}

// ---------------------------------------------------------------------------
// Handler
// ---------------------------------------------------------------------------

// TelemetryHandler adapts TelemetryQuery to Gin HTTP endpoints.
type TelemetryHandler struct {
	svc TelemetryQuery
}

// NewTelemetryHandler creates a TelemetryHandler wired to the given query service.
func NewTelemetryHandler(svc TelemetryQuery) *TelemetryHandler {
	return &TelemetryHandler{svc: svc}
}

// ListGPUs handles GET /api/v1/gpus.
func (h *TelemetryHandler) ListGPUs(c *gin.Context) {
	gpus, err := h.svc.ListGPUsWithTelemetry(c.Request.Context())
	if err != nil {
		log.Printf("ERROR ListGPUs: %v", err)
		c.JSON(http.StatusInternalServerError, errorBody{Error: "failed to list gpus"})
		return
	}
	c.JSON(http.StatusOK, listGpusResponse{GPUs: gpus})
}

// GetGPUTelemetry handles GET /api/v1/gpus/:id/telemetry.
func (h *TelemetryHandler) GetGPUTelemetry(c *gin.Context) {
	gpuID := strings.TrimSpace(c.Param("id"))
	if gpuID == "" {
		c.JSON(http.StatusBadRequest, errorBody{Error: "gpu id is required"})
		return
	}

	start, end, err := parseTimeWindow(c.Request)
	if err != nil {
		c.JSON(http.StatusBadRequest, errorBody{Error: err.Error()})
		return
	}

	entries, err := h.svc.ListTelemetryForGPU(c.Request.Context(), gpuID, start, end)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidTimeWindow) {
			c.JSON(http.StatusBadRequest, errorBody{Error: err.Error()})
			return
		}
		log.Printf("ERROR GetGPUTelemetry gpu=%s: %v", gpuID, err)
		c.JSON(http.StatusInternalServerError, errorBody{Error: "failed to query telemetry"})
		return
	}
	c.JSON(http.StatusOK, telemetryListResponse{Entries: toTelemetryDTOs(entries)})
}

// ---------------------------------------------------------------------------
// Time parsing helpers
// ---------------------------------------------------------------------------

// parseTimeWindow extracts optional start_time / end_time query parameters.
func parseTimeWindow(r *http.Request) (start, end *time.Time, err error) {
	q := r.URL.Query()
	if s := strings.TrimSpace(q.Get("start_time")); s != "" {
		t, perr := parseRFC3339Like(s)
		if perr != nil {
			return nil, nil, perr
		}
		start = &t
	}
	if s := strings.TrimSpace(q.Get("end_time")); s != "" {
		t, perr := parseRFC3339Like(s)
		if perr != nil {
			return nil, nil, perr
		}
		end = &t
	}
	return start, end, nil
}

// parseRFC3339Like parses an RFC 3339 (or RFC 3339 Nano) timestamp, after
// running normalizeTimeQueryParam to accept common non-standard variants.
func parseRFC3339Like(s string) (time.Time, error) {
	s = normalizeTimeQueryParam(s)
	if t, err := time.Parse(time.RFC3339Nano, s); err == nil {
		return t, nil
	}
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t, nil
	}
	return time.Time{}, errors.New(
		"start_time and end_time must be RFC 3339 timestamps " +
			"(space separator and +00 / +0000 UTC suffix also accepted)")
}

// normalizeTimeQueryParam rewrites common non-RFC variants into RFC 3339:
//
//	"2026-05-01 08:00:00+00"   → "2026-05-01T08:00:00Z"
//	"2026-05-01T08:00:00+00"   → "2026-05-01T08:00:00Z"
//	"2026-05-01T08:00:00+0000" → "2026-05-01T08:00:00Z"
func normalizeTimeQueryParam(s string) string {
	s = strings.TrimSpace(s)

	// Replace space date/time separator with 'T'.
	if len(s) >= 11 && s[10] == ' ' && !strings.Contains(s, "T") {
		s = s[:10] + "T" + strings.TrimSpace(s[11:])
	}

	// Rewrite bare UTC offset suffixes (+0000, +00) to Z.
	switch {
	case strings.HasSuffix(s, "+0000") && strings.Count(s, "+") == 1:
		s = strings.TrimSuffix(s, "+0000") + "Z"
	case strings.HasSuffix(s, "+00") && !strings.Contains(s, "+00:") && strings.Count(s, "+") == 1:
		s = strings.TrimSuffix(s, "+00") + "Z"
	}
	return s
}

// ---------------------------------------------------------------------------
// Response DTOs
// ---------------------------------------------------------------------------

type listGpusResponse struct {
	GPUs []string `json:"gpus"`
}

type telemetryListResponse struct {
	Entries []telemetryDTO `json:"entries"`
}

type telemetryDTO struct {
	ID                  int64     `json:"id"`
	MetricName          string    `json:"metric_name"`
	GPUID               string    `json:"gpu_id"`
	Device              string    `json:"device"`
	UUID                string    `json:"uuid"`
	ModelName           string    `json:"model_name"`
	HostName            string    `json:"host_name"`
	Value               float64   `json:"value"`
	LabelsRaw           string    `json:"labels_raw"`
	ProcessedAtUnixNano int64     `json:"processed_at_unix_nano"`
	CreatedAt           time.Time `json:"created_at"`
}

type errorBody struct {
	Error string `json:"error"`
}

func toTelemetryDTOs(rows []domain.Telemetry) []telemetryDTO {
	out := make([]telemetryDTO, 0, len(rows))
	for _, r := range rows {
		out = append(out, telemetryDTO{
			ID:                  r.ID,
			MetricName:          r.MetricName,
			GPUID:               r.GPUID,
			Device:              r.Device,
			UUID:                r.UUID,
			ModelName:           r.ModelName,
			HostName:            r.HostName,
			Value:               r.Value,
			LabelsRaw:           r.LabelsRaw,
			ProcessedAtUnixNano: r.ProcessedAtUnixNano,
			CreatedAt:           r.CreatedAt,
		})
	}
	return out
}
