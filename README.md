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

The OpenAPI contract lives in [`api/openapi.yaml`](api/openapi.yaml). With the server running, open **Swagger UI** at **`/docs`** on the same host and port as the API. **`make run`** / **`make run-local`** default to **`http://127.0.0.1:18080/docs`** when **`HTTP_ADDR`** is unset (avoids `:8080` clashes on the host). Use **`HTTP_ADDR=0.0.0.0:8080 make run`** if you want port **8080**. In the dev container, Compose sets **`HTTP_ADDR`** to **`:8080`**; from the Mac use **`http://127.0.0.1:8081/docs`** when Docker maps **`8081→8080`** (see [examples/README.md](examples/README.md)). The page loads **`/openapi.yaml`**; Swagger assets load from **unpkg**.

## Run the API

The process connects to PostgreSQL at startup (migrations run on `OnStart`). **Postgres must be reachable** before `make run` succeeds.

- **Dev container** ([`.devcontainer/docker-compose.yml`](.devcontainer/docker-compose.yml)): Compose defines only the **`app`** service. Postgres must run **elsewhere** (e.g. a sibling compose stack publishing **`127.0.0.1:5433→5432`**). Default `DATABASE_URL` uses **`host.docker.internal:5433`** (`telemetry` / `telemetry`). Start that DB before **`make run`**. Override **`DATABASE_URL`** if your DB host or network differs (e.g. shared Docker network with hostname **`db`**).
- **Host terminal** (Postgres on **`127.0.0.1:5433`**): use **`make run-local`**, or set **`DATABASE_URL`** to match your instance.

### `bind: address already in use`

Something else is already listening on the port in **`HTTP_ADDR`**. **`make run`** defaults **`HTTP_ADDR`** to **`0.0.0.0:18080`** only when it is **unset**; inside the dev container **`HTTP_ADDR`** is **`0.0.0.0:8080`** from Compose. Check listeners with **`lsof -nP -iTCP:PORT -sTCP:LISTEN`**. Stray API processes: **`make stop`**. Second **`make run`** in the same environment: stop the first instance. To force another port: **`HTTP_ADDR=0.0.0.0:9090 make run`**.
