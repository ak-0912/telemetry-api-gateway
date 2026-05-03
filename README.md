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

The OpenAPI contract lives in [`api/openapi.yaml`](api/openapi.yaml). With the server running, open **Swagger UI** at **`/docs`** on the same host and port as the API — for example **`http://127.0.0.1:8080/docs`** when the process listens on `:8080` (default `HTTP_ADDR`). If Docker publishes the app as **`8081→8080`**, use **`http://127.0.0.1:8081/docs`** instead (see [examples/README.md](examples/README.md)). The page loads the embedded spec from **`/openapi.yaml`**. The browser fetches Swagger UI assets from **unpkg**; you can still use `/openapi.yaml` with any other OpenAPI client offline.
