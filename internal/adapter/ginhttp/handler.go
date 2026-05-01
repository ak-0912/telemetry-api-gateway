package ginhttp

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"telemetry-api-gateway/internal/domain"
)

// TelemetryQuery is the application port used by HTTP handlers (enables mocking in tests).
type TelemetryQuery interface {
	ListGPUsWithTelemetry(ctx context.Context) ([]string, error)
	ListTelemetryForGPU(ctx context.Context, gpuID string, start, end *time.Time) ([]domain.Telemetry, error)
}

// TelemetryHandler is the Gin adapter for telemetry queries.
type TelemetryHandler struct {
	svc TelemetryQuery
}

// NewTelemetryHandler constructs the HTTP adapter.
func NewTelemetryHandler(svc TelemetryQuery) *TelemetryHandler {
	return &TelemetryHandler{svc: svc}
}

func (h *TelemetryHandler) ListGPUs(c *gin.Context) {
	ctx := c.Request.Context()
	gpus, err := h.svc.ListGPUsWithTelemetry(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorBody{Error: "failed to list gpus"})
		return
	}
	c.JSON(http.StatusOK, listGpusResponse{GPUs: gpus})
}

func (h *TelemetryHandler) GetGPUTelemetry(c *gin.Context) {
	ctx := c.Request.Context()
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
	entries, err := h.svc.ListTelemetryForGPU(ctx, gpuID, start, end)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidTimeWindow) {
			c.JSON(http.StatusBadRequest, errorBody{Error: err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, errorBody{Error: "failed to query telemetry"})
		return
	}
	c.JSON(http.StatusOK, telemetryListResponse{Entries: toTelemetryDTOs(entries)})
}

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

func parseRFC3339Like(s string) (time.Time, error) {
	if t, err := time.Parse(time.RFC3339Nano, s); err == nil {
		return t, nil
	}
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return time.Time{}, errors.New("start_time and end_time must be RFC3339 or RFC3339Nano timestamps")
	}
	return t, nil
}

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
	for _, t := range rows {
		out = append(out, telemetryDTOFromDomain(t))
	}
	return out
}

func telemetryDTOFromDomain(t domain.Telemetry) telemetryDTO {
	return telemetryDTO{
		ID:                  t.ID,
		MetricName:          t.MetricName,
		GPUID:               t.GPUID,
		Device:              t.Device,
		UUID:                t.UUID,
		ModelName:           t.ModelName,
		HostName:            t.HostName,
		Value:               t.Value,
		LabelsRaw:           t.LabelsRaw,
		ProcessedAtUnixNano: t.ProcessedAtUnixNano,
		CreatedAt:           t.CreatedAt,
	}
}
