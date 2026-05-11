OpenAPI (Swagger) specification for your REST APIs

API Requirements

Program : go
REST APIs - Expose REST (OpenAPI) on the API gateway for clients.

DB stores - 
CREATE TABLE IF NOT EXISTS telemetry (
    id BIGSERIAL PRIMARY KEY,
    metric_name TEXT NOT NULL,
    gpu_id TEXT NOT NULL,
    device TEXT NOT NULL,
    uuid TEXT NOT NULL,
    model_name TEXT NOT NULL,
    host_name TEXT NOT NULL,
    value DOUBLE PRECISION NOT NULL,
    labels_raw TEXT NOT NULL,
    processed_at_unix_nano BIGINT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

Design and implement the following endpoints:
1. List All GPUs
    • Return a list of all GPUs for which telemetry data is available.
2. Query Telemetry by GPU
    • Return all telemetry entries for a specific GPU, ordered by time.
    • Support optional time window filters:
        o start_time (inclusive)
        o end_time (inclusive)

Example API Design:
    GET /api/v1/gpus
    GET /api/v1/gpus/{id}/telemetry
    GET /api/v1/gpus/{id}/telemetry?start_time=...&end_time=...

Postgres DB

postgres:
    image: postgres:16
    restart: unless-stopped
    environment:
      POSTGRES_DB: telemetry
      POSTGRES_USER: telemetry
      POSTGRES_PASSWORD: telemetry

