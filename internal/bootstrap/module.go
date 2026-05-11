// Package bootstrap wires all application components via Uber Fx and manages
// the process lifecycle (database pool, migrations, HTTP server).
package bootstrap

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/fx"

	"telemetry-api-gateway/internal/adapter/ginhttp"
	"telemetry-api-gateway/internal/adapter/postgres"
	apptelemetry "telemetry-api-gateway/internal/application/telemetry"
	"telemetry-api-gateway/internal/platform/config"
)

// Module is the top-level Fx module that provides config, database, application
// services, HTTP adapter, and server lifecycle hooks.
//
// fx.Annotate + fx.As is used to bind concrete types to their port interfaces
// so Dig can satisfy constructor parameters declared as interfaces.
var Module = fx.Module("telemetry-api-gateway",
	fx.Provide(
		config.Load,
		newPool,
		fx.Annotate(postgres.NewTelemetryRepository, fx.As(new(apptelemetry.Repository))),
		fx.Annotate(apptelemetry.NewQueryService, fx.As(new(ginhttp.TelemetryQuery))),
		ginhttp.NewTelemetryHandler,
		ginhttp.NewRouter,
	),
	fx.Invoke(registerHTTPServer),
)

// newPool creates a pgxpool.Pool, runs schema migrations on start, and closes
// the pool on stop.
func newPool(lc fx.Lifecycle, cfg config.Config) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(context.Background(), cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("connect to postgres at %s: %w "+
			"(ensure Postgres is running and reachable; from a container the DB "+
			"must listen on 0.0.0.0 on the host port)",
			redactPostgresURL(cfg.DatabaseURL), err)
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			if err := postgres.Migrate(ctx, pool); err != nil {
				return fmt.Errorf("database migration: %w", err)
			}
			log.Println("database migration applied")
			return nil
		},
		OnStop: func(context.Context) error {
			pool.Close()
			log.Println("database pool closed")
			return nil
		},
	})
	return pool, nil
}

// redactPostgresURL replaces the password in a Postgres URL with "***" for safe logging.
func redactPostgresURL(raw string) string {
	u, err := url.Parse(raw)
	if err != nil || u.User == nil {
		return raw
	}
	name := u.User.Username()
	if _, set := u.User.Password(); set {
		u.User = url.UserPassword(name, "***")
	} else {
		u.User = url.User(name)
	}
	return u.String()
}

// registerHTTPServer starts the HTTP server on OnStart and gracefully shuts it
// down on OnStop.
func registerHTTPServer(lc fx.Lifecycle, cfg config.Config, h http.Handler) {
	srv := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           h,
		ReadHeaderTimeout: 10 * time.Second,
	}

	lc.Append(fx.Hook{
		OnStart: func(context.Context) error {
			log.Printf("http server listening on %s", cfg.HTTPAddr)
			go func() {
				if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
					log.Printf("http server error: %v", err)
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			log.Println("http server shutting down")
			shutdownCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
			defer cancel()
			return srv.Shutdown(shutdownCtx)
		},
	})
}
