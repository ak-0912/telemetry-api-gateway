.PHONY: build run run-local db-up test test-coverage clean

BIN := bin/api
COMPOSE := docker compose -f .devcontainer/docker-compose.yml

# Postgres URL when connecting from the host machine (Compose publishes db on 5433).
LOCAL_DATABASE_URL ?= postgresql://postgres:postgres@localhost:5433/telemetry

build:
	go build -o $(BIN) ./cmd/api

# Uses DATABASE_URL from the environment (devcontainer sets ...@db:5432/...).
run:
	go run ./cmd/api

# Use when `db` does not resolve (e.g. Go runs on the host, not inside Compose).
run-local:
	DATABASE_URL=$(LOCAL_DATABASE_URL) go run ./cmd/api

db-up:
	$(COMPOSE) up -d db

test:
	go test ./...

test-coverage:
	go test ./... -coverprofile=coverage.out -covermode=atomic
	go tool cover -func=coverage.out

clean:
	rm -f coverage.out $(BIN)
