# API examples

These routes are **GET** only: there is **no JSON request body**. Send parameters in the **path** and **query string**.

- **Base URL:** default **`http://localhost:8080`** (matches `HTTP_ADDR` `0.0.0.0:8080`). **Swagger UI** at **`/docs`**, OpenAPI YAML at **`/openapi.yaml`** (same host/port). If the devcontainer publishes **`8081→8080`** to the Mac, use **`http://localhost:8081`** on the host (still **`http://localhost:8080`** from inside the container).

Files:

| File | Purpose |
|------|---------|
| `http/api.http` | VS Code **REST Client** / IntelliJ HTTP requests |
| `test-cases/valid.json` | Valid calls: method, path, query, expected status |
| `test-cases/invalid.json` | Invalid calls: bad query / path, expected status |
| `responses/*.json` | Example **response** bodies (what you should see) |
| [`seed.sql`](seed.sql) | Inserts `gpu-001` / `gpu-002` / `gpu-003` with **`processed_at_unix_nano` matching** the `created_at` instants (required for time-filtered GETs) |

## Empty `gpus` or `entries`

1. **`GET /api/v1/gpus` returns `"gpus": []`** — the `telemetry` table has no rows (or wrong database). Load sample data: run [`seed.sql`](seed.sql) against the same DB the API uses (`DATABASE_URL`).

2. **`GET .../telemetry` with `start_time` / `end_time` returns `"entries": []`** — filters are **inclusive** on `processed_at_unix_nano` only (not `created_at`). If that column does not match the wall times you expect (e.g. placeholder values like `1000000001000000000`), every row is outside a 2026 window. **Omit** `start_time` / `end_time` to confirm rows exist, then fix nanoseconds (see `seed.sql`) or widen the window to include the actual stored nanos.

3. **Path `gpu_id` must match** the `gpu_id` column exactly (e.g. `gpu-001` after seeding), not a display name.

## `start_time` / `end_time` in the URL

Use **RFC3339** in query params, or common SQL-style values the API now normalizes:

- `2026-05-01T08:00:00Z` or `2026-05-01T08:00:00+00:00`
- `2026-05-01 08:00:00+00` (space between date and time; `+00` = UTC)

Encode for HTTP: spaces as **`%20`**, a literal **`+`** in the value as **`%2B`** if your client treats `+` as a space (HTML forms). Example:

`.../telemetry?start_time=2026-05-01%2008:00:00%2B00&end_time=2026-05-03%2018:45:00%2B00`

Or avoid `+` entirely: `...start_time=2026-05-01T08:00:00Z&end_time=2026-05-03T18:45:00Z`.

## Connection refused

1. **Match host/port to where the process listens.** Host-only `go run` / `make run` on the Mac → **`http://127.0.0.1:8080/healthz`**. API in the dev container with **`8081:8080`** published → **`http://127.0.0.1:8081/healthz`** on the Mac (that map does **not** reach a process that only runs on the host).

2. **Recreate the dev container** after changing `.devcontainer/docker-compose.yml` ports (`Dev Containers: Rebuild Container`).

3. When using Docker, confirm the publish: `docker ps` should show something like `127.0.0.1:8081->8080/tcp` on the `app` container.

4. If `8081` is busy, change the left side in compose (e.g. `127.0.0.1:18081:8080`) and use that host port in your HTTP client.
