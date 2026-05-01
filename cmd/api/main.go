// Package main is the process entrypoint; Fx composes config, DB, use cases, Gin, and HTTP lifecycle.
package main

import (
	"go.uber.org/fx"

	"telemetry-api-gateway/internal/bootstrap"
)

func main() {
	fx.New(
		bootstrap.Module,
	).Run()
}
