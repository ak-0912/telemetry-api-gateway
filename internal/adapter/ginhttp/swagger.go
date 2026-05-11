package ginhttp

import (
	"net/http"

	"github.com/gin-gonic/gin"

	apidoc "telemetry-api-gateway/api"
)

// swaggerUIPage is a self-contained HTML page that loads Swagger UI from a CDN
// and points it at the embedded /openapi.yaml served by this binary.
const swaggerUIPage = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8" />
  <title>Telemetry API — Swagger UI</title>
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5.11.0/swagger-ui.css" crossorigin="anonymous" />
  <style>body { margin: 0; } #swagger-ui { max-width: 100%; }</style>
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://unpkg.com/swagger-ui-dist@5.11.0/swagger-ui-bundle.js" crossorigin="anonymous"></script>
  <script src="https://unpkg.com/swagger-ui-dist@5.11.0/swagger-ui-standalone-preset.js" crossorigin="anonymous"></script>
  <script>
    window.onload = function () {
      window.ui = SwaggerUIBundle({
        url: window.location.origin + "/openapi.yaml",
        dom_id: "#swagger-ui",
        deepLinking: true,
        presets: [SwaggerUIBundle.presets.apis, SwaggerUIStandalonePreset],
        layout: "StandaloneLayout",
      });
    };
  </script>
</body>
</html>`

// registerSwaggerRoutes serves /docs, /docs/, and /openapi.yaml.
func registerSwaggerRoutes(r gin.IRoutes) {
	r.GET("/openapi.yaml", func(c *gin.Context) {
		c.Data(http.StatusOK, "application/yaml; charset=utf-8", apidoc.OpenAPIYAML)
	})
	serveDocsPage := func(c *gin.Context) {
		c.Header("Cache-Control", "no-cache")
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(swaggerUIPage))
	}
	r.GET("/docs", serveDocsPage)
	r.GET("/docs/", serveDocsPage)
}
