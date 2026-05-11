package ginhttp

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSwaggerRoutes_DocsAndOpenAPI(t *testing.T) {
	t.Parallel()
	h := &TelemetryHandler{}
	r := NewRouter(h)

	for _, path := range []string{"/docs", "/docs/", "/openapi.yaml"} {
		t.Run(path, func(t *testing.T) {
			t.Parallel()
			req := httptest.NewRequest(http.MethodGet, path, nil)
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)
			if rec.Code != http.StatusOK {
				t.Fatalf("%s: status %d", path, rec.Code)
			}
			if path == "/openapi.yaml" && !strings.Contains(rec.Body.String(), "openapi:") {
				t.Fatalf("%s: expected openapi preamble in body", path)
			}
		})
	}
}
