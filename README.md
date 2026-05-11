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

The OpenAPI contract lives in [`api/openapi.yaml`](api/openapi.yaml). With the server running, open **Swagger UI** at **`/docs`** on the same host and port as the API. **`make run`** / **`make run-local`** default to **`http://127.0.0.1:18080/docs`** when **`HTTP_ADDR`** is unset (avoids `:8080` clashes on the host). Use **`HTTP_ADDR=0.0.0.0:8080 make run`** if you want port **8080**. In the dev container, Compose sets **`HTTP_ADDR`** to **`:8080`**; from the Mac use **`http://127.0.0.1:8081/docs`** when Docker maps **`8081→8080`**. The page loads **`/openapi.yaml`**; Swagger assets load from **unpkg**.

## Run the API

The process connects to PostgreSQL at startup (migrations run on `OnStart`). **Postgres must be reachable** before `make run` succeeds.

- **Dev container** ([`.devcontainer/docker-compose.yml`](.devcontainer/docker-compose.yml)): Compose defines only the **`app`** service. Postgres must run **elsewhere** (e.g. a sibling compose stack publishing **`127.0.0.1:5433→5432`**). Default `DATABASE_URL` uses **`host.docker.internal:5433`** (`telemetry` / `telemetry`). Start that DB before **`make run`**. Override **`DATABASE_URL`** if your DB host or network differs (e.g. shared Docker network with hostname **`db`**).
- **Host terminal** (Postgres on **`127.0.0.1:5433`**): use **`make run-local`**, or set **`DATABASE_URL`** to match your instance.

### `bind: address already in use`

Something else is already listening on the port in **`HTTP_ADDR`**. **`make run`** defaults **`HTTP_ADDR`** to **`0.0.0.0:18080`** only when it is **unset**; inside the dev container **`HTTP_ADDR`** is **`0.0.0.0:8080`** from Compose. Check listeners with **`lsof -nP -iTCP:PORT -sTCP:LISTEN`**. Stray API processes: **`make stop`**. Second **`make run`** in the same environment: stop the first instance. To force another port: **`HTTP_ADDR=0.0.0.0:9090 make run`**.

## Kubernetes (Helm)

Deploy with the chart under [`helm/telemetry-api-gateway`](helm/telemetry-api-gateway). Defaults (image, **`database.url`**, ingress, **`fullnameOverride`**) live in [`values.yaml`](helm/telemetry-api-gateway/values.yaml). The chart usually creates a Secret named **`telemetry-api-gateway`** with key **`DATABASE_URL`** (unless **`database.existingSecret`** is set). Create the namespace with **`--create-namespace`**.

**kind:** build and load the image so the tag matches the chart default (**`docker.io/library/telemetry-api-gateway:latest`**):

```bash
docker build -t telemetry-api-gateway:latest .
kind load docker-image telemetry-api-gateway:latest --name <your-kind-cluster>
```

### Install

```bash
helm upgrade --install telemetry-api-gateway ./helm/telemetry-api-gateway \
  --namespace telemetry --create-namespace
```

Override only what you need, for example **`--set ingress.hosts[0].host=api.example.com`**.

If DB credentials come from an existing Secret:

```bash
helm upgrade --install telemetry-api-gateway ./helm/telemetry-api-gateway \
  --namespace telemetry --create-namespace \
  --set database.createSecret=false \
  --set database.existingSecret=telemetry-db-credentials \
  --set database.existingSecretKey=DATABASE_URL
```

### Check Postgres from the cluster (same view as the API pod)

The runtime image is **distroless** (no shell), so you cannot rely on **`kubectl exec ... -- sh`** into **`telemetry-api-gateway`** pods.

**1. See what URL the app uses (Helm release):**

```bash
helm get values telemetry-api-gateway -n telemetry -o yaml
```

**2. Confirm the Secret exists and which key holds the URL:**

```bash
kubectl get secret telemetry-api-gateway -n telemetry -o yaml
# Key is usually DATABASE_URL (see database.existingSecretKey in values).
```

If you use **`database.existingSecret`**, replace **`telemetry-api-gateway`** below with that secret name.

**3. TCP reachability** (replace host with the hostname from **`database.url`**, e.g. your Postgres **`Service`** DNS name):

```bash
kubectl run -n telemetry netcheck --rm -i --restart=Never --image=busybox:1.36 \
  --command -- sh -c 'nc -zv telemetry-db-postgres 5432'
```

Exit code **`0`** means something accepted the connection on that port (not a full auth check).

**4. Full check with `psql`** (same connection string as the app): run a one-off pod in **`telemetry`** so cluster DNS and **NetworkPolicies** match what **`telemetry-api-gateway`** sees:

```bash
NS=telemetry
SECRET=telemetry-api-gateway
URL=$(kubectl get secret "$SECRET" -n "$NS" -o jsonpath='{.data.DATABASE_URL}' | base64 -d)
kubectl run "pgcheck-$(date +%s)" -n "$NS" --rm -i --restart=Never \
  --image=postgres:16-alpine \
  --command -- psql "$URL" -c 'select 1 as ok'
```

You should see a row **`1`**. If this fails but **`nc`** works, suspect **credentials**, **database name**, or **TLS/sslmode**.

**5. Indirect signal from the app**

```bash
kubectl logs -n telemetry deploy/telemetry-api-gateway --tail=100
kubectl describe pod -n telemetry -l app.kubernetes.io/name=telemetry-api-gateway
```

Startup or migration errors often mention **connection refused**, **timeout**, or **authentication failed**.
