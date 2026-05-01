# telemetry-api-gateway

REST API exposing GPU telemetry backed by PostgreSQL.

## Layout (DDD / clean architecture)

| Path | Role |
|------|------|
| [`cmd/api`](cmd/api) | Process entrypoint |
| [`internal/domain`](internal/domain) | Entities and domain errors |
| [`internal/application`](internal/application) | Use cases and ports (e.g. `telemetry.Repository`) |
| [`internal/adapter/ginhttp`](internal/adapter/ginhttp) | Gin HTTP driving adapter |
| [`internal/adapter/postgres`](internal/adapter/postgres) | Postgres driven adapter |
| [`internal/platform/config`](internal/platform/config) | Environment config |
| [`internal/bootstrap`](internal/bootstrap) | **[Fx](https://github.com/uber-go/fx)** module: DB pool lifecycle, migrate, HTTP server lifecycle |

## REST stack

The HTTP API uses **[Gin](https://github.com/gin-gonic/gin)** for routing and JSON, **[Uber Fx](https://github.com/uber-go/fx)** for compile-time dependency injection and lifecycle (`OnStart` / `OnStop`), and `net/http` for serving. Entry: [`cmd/api/main.go`](cmd/api/main.go).

The OpenAPI contract lives in [`api/openapi.yaml`](api/openapi.yaml).
