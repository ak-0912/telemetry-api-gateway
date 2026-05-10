package bootstrap

import (
	"context"
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

// Module is the Fx wiring for the telemetry API process.
//
// Dig only satisfies constructor parameters of interface type when that interface is
// registered explicitly: fx.Annotate(..., fx.As(new(Port))) binds the concrete constructor
// result to the port type (e.g. *TelemetryRepository → Repository, *QueryService → TelemetryQuery).
var Module = fx.Module("telemetry-api-gateway",
	fx.Provide(
		config.Load,
		newPool,
		fx.Annotate(
			postgres.NewTelemetryRepository,
			fx.As(new(apptelemetry.Repository)),
		),
		fx.Annotate(
			apptelemetry.NewQueryService,
			fx.As(new(ginhttp.TelemetryQuery)),
		),
		ginhttp.NewTelemetryHandler,
		ginhttp.NewRouter,
	),
	fx.Invoke(registerHTTPServer),
)

func newPool(lc fx.Lifecycle, cfg config.Config) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(context.Background(), cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("%w\npostgres: check DATABASE_URL / DATABASE_HOST; from a container the DB must listen on 0.0.0.0 (not only 127.0.0.1) on the host port (default 5433). attempted: %s\nhint: ensure Postgres is running and reachable (e.g. sibling compose publishing 5433:5432 on the host)",
			err, redactPostgresURL(cfg.DatabaseURL))
	}
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			return postgres.Migrate(ctx, pool)
		},
		OnStop: func(context.Context) error {
			pool.Close()
			return nil
		},
	})
	return pool, nil
}

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

func registerHTTPServer(lc fx.Lifecycle, cfg config.Config, h http.Handler) {
	srv := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           h,
		ReadHeaderTimeout: 10 * time.Second,
	}
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			log.Printf("listening on %s", cfg.HTTPAddr)
			go func() {
				if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					log.Printf("http: %v", err)
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			shutdownCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
			defer cancel()
			return srv.Shutdown(shutdownCtx)
		},
	})
}
