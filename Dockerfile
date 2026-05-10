# syntax=docker/dockerfile:1

# --- Stage 1: compile ---
ARG GO_VERSION=1.25
FROM golang:${GO_VERSION}-bookworm AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/api ./cmd/api

# --- Stage 2: minimal runtime image ---
FROM gcr.io/distroless/static-debian12:nonroot AS runtime

WORKDIR /

COPY --from=build /out/api /api

# Matches internal/platform/config default; override with -e HTTP_ADDR=... if needed.
ENV HTTP_ADDR=0.0.0.0:8080

EXPOSE 8080

USER nonroot:nonroot

ENTRYPOINT ["/api"]
