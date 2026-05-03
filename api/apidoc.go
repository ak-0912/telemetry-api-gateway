// Package apidoc embeds the OpenAPI document shipped with the service (served at /openapi.yaml).
package apidoc

import _ "embed"

//go:embed openapi.yaml
var OpenAPIYAML []byte
