# Future Improvements

## API
- Pagination (`limit`/`offset`) on telemetry queries
- Filter by `metric_name` query param
- Aggregation endpoint (min/avg/max over time windows)
- Write/ingest endpoint (`POST /api/v1/telemetry`)

## Observability
- Structured JSON logging (`slog` or `zap`)
- Prometheus `/metrics` endpoint with request rate, error rate, latency histograms
- OpenTelemetry tracing (Gin → QueryService → Postgres)
- Deep readiness probe (`/readyz` that pings Postgres)

## Security
- Authentication (API key or JWT middleware)
- Rate limiting per client
- CORS configuration
- Input validation (gpu_id format, max time window)

## Database
- Expose connection pool settings (`MaxConns`, `MinConns`, `MaxConnLifetime`)
- Migration runner (`golang-migrate`) for multi-step schema evolution
- Table partitioning strategy for large-scale deployments

## Testing
- Integration tests with `testcontainers-go` for the Postgres adapter
- CI pipeline (lint, test, build image, Helm lint)
- Load testing (`k6` or `vegeta`)

## Deployment
- HorizontalPodAutoscaler and PodDisruptionBudget in Helm chart
- NetworkPolicy restricting DB access to API pods only
- Multi-environment values overlays (`values-prod.yaml`)
