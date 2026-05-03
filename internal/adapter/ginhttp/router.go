package ginhttp

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// NewRouter mounts Gin routes (transport / driving adapter).
func NewRouter(h *TelemetryHandler) http.Handler {
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())
	r.GET("/healthz", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})
	registerSwaggerRoutes(r)
	registerTelemetryRoutes(r.Group("/api/v1"), h)
	return r
}

func registerTelemetryRoutes(v1 gin.IRoutes, h *TelemetryHandler) {
	v1.GET("/gpus", h.ListGPUs)
	v1.GET("/gpus/:id/telemetry", h.GetGPUTelemetry)
}
