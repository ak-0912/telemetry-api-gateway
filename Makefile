.PHONY: build run run-local stop test test-coverage clean

BIN := bin/api

# Postgres published on host :5433 (e.g. sibling telemetry-db compose). Native `go run` from host.
LOCAL_DATABASE_URL ?= postgres://telemetry:telemetry@127.0.0.1:5433/telemetry?sslmode=disable

# When HTTP_ADDR is unset, Makefile picks this port so host `make run` avoids :8080 (often busy with Cursor, mq-server, etc.).
# Dev container sets HTTP_ADDR=0.0.0.0:8080 via Compose — unchanged. Override anytime: HTTP_ADDR=0.0.0.0:8080 make run
DEFAULT_LISTEN_ADDR ?= 0.0.0.0:18080

build:
	go build -o $(BIN) ./cmd/api

# Uses DATABASE_URL if set; otherwise config.Load uses defaultDatabaseURL() host discovery (see internal/platform/config).
run:
	@if [ -z "$$HTTP_ADDR" ]; then export HTTP_ADDR="$(DEFAULT_LISTEN_ADDR)"; fi && go run ./cmd/api

# Postgres on localhost (e.g. Mac terminal). Same HTTP_ADDR rule as `run`.
run-local:
	@if [ -z "$$HTTP_ADDR" ]; then export HTTP_ADDR="$(DEFAULT_LISTEN_ADDR)"; fi && \
	export DATABASE_URL="$(LOCAL_DATABASE_URL)" && go run ./cmd/api

# Ports to scan for stray LISTENers after pkill (host or container).
FREE_PORTS ?= 8080 8081 18080

# Stop `go run` / $(BIN) and try to release :8080 / :8081 when the listener is this app (not Docker/Cursor).
stop:
	@echo "Stopping telemetry-api-gateway API..."
	@-pkill -f "go run ./cmd/api" 2>/dev/null && echo "  stopped: go run ./cmd/api" || true
	@-pkill -f "telemetry-api-gateway/cmd/api" 2>/dev/null && echo "  stopped: go run .../cmd/api" || true
	@-pkill -f "[b]in/api" 2>/dev/null && echo "  stopped: $(BIN)" || true
	@echo "Checking LISTEN on ports $(FREE_PORTS) (skip Docker/Cursor)..."
	@for port in $(FREE_PORTS); do \
	  pids=$$(lsof -tiTCP:$$port -sTCP:LISTEN 2>/dev/null || true); \
	  for pid in $$pids; do \
	    [ -z "$$pid" ] && continue; \
	    cmd=$$(ps -p $$pid -o command= 2>/dev/null || ps -p $$pid -o args= 2>/dev/null || true); \
	    case "$$cmd" in *docker*|*Docker*|*com.docker*|*Cursor*) continue;; esac; \
	    case "$$cmd" in *go*|*cmd/api*|*bin/api*|*telemetry-api-gateway*|*go-build*) \
	      kill $$pid 2>/dev/null && echo "  freed port $$port (PID $$pid)" || true;; \
	    esac; \
	  done; \
	done
	@echo "Done. If :8081 is still in use, it is often docker-proxy — run: docker compose -f .devcontainer/docker-compose.yml stop"

test:
	go test ./...

test-coverage:
	go test ./... -coverprofile=coverage.out -covermode=atomic
	go tool cover -func=coverage.out

clean:
	rm -f coverage.out $(BIN)
