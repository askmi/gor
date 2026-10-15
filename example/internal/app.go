package internal

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	gorep "gor/pkg/repository"
	gor "gor/pkg/server"

	"github.com/BurntSushi/toml"
	_ "github.com/jackc/pgx/v5/stdlib"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

var _ gorep.Repository[User, int] = (*UserRepository)(nil)

func init() {
	// https://pkg.go.dev/log/slog
	var h slog.Handler = slog.NewTextHandler(os.Stdout, nil)

	h = &TraceLogHandler{Handler: h}
	l := slog.New(h)
	slog.SetDefault(l)

	fileBytes, err := os.ReadFile("config.toml")
	if err != nil {
		panic("failed to read file: " + err.Error())
	}
	cfg = defaultConfig()
	err = toml.Unmarshal(fileBytes, &cfg)
	if err != nil {
		panic("failed to unmarshal TOML: " + err.Error())
	}
	slog.Info("config loaded", "config", cfg)

	db, err = setupDatabase(context.Background(), cfg.Database)
	if err != nil {
		panic("failed to set up database: " + err.Error())
	}
}

var cfg Config

// db is nil while the database is disabled; every user of it must check.
var db *sql.DB

const AppName = "gor-example-service"

type (
	Config struct {
		Server   ServerConfig   `toml:"server"`
		Database DatabaseConfig `toml:"database"`
		Cache    CacheConfig    `toml:"cache"`
	}

	ServerConfig struct {
		Port              int      `toml:"Port"`
		ReadTimeout       Duration `toml:"ReadTimeout"`
		WriteTimeout      Duration `toml:"WriteTimeout"`
		IdleTimeout       Duration `toml:"IdleTimeout"`
		ReadHeaderTimeout Duration `toml:"ReadHeaderTimeout"`
	}

	DatabaseConfig struct {
		DSN     string `toml:"dsn"`
		Enabled bool   `toml:"enabled"`
		// EnableMigration runs the pending migrations during startup. Turn it
		// off to apply the schema as a separate deployment step instead, which
		// is what several instances starting at once require.
		EnableMigration bool `toml:"enableMigration"`
	}

	CacheConfig struct {
		Enabled bool `toml:"enabled"`
	}

	// Duration reads a TOML string such as "5s" or "1m" into a time.Duration.
	// TOML has no duration type, so the value arrives as text.
	Duration time.Duration
)

func (d *Duration) UnmarshalText(b []byte) error {
	v, err := time.ParseDuration(string(b))
	if err != nil {
		return err
	}
	*d = Duration(v)
	return nil
}

// defaultConfig seeds values that TOML omission would otherwise leave at zero.
// For the timeouts zero means "no limit", not "use a sane default".
func defaultConfig() Config {
	var c Config
	c.Server.Port = 8080
	c.Server.ReadHeaderTimeout = Duration(5 * time.Second)
	c.Server.IdleTimeout = Duration(60 * time.Second)
	return c
}

// setupDatabase opens the pool and brings the schema up to date. It returns a
// nil handle when the database is disabled, leaving nothing to close.
//
// The migration is DDL, and the repository issues DML only, so the schema is
// created here, before the first statement is prepared against it. Running it
// in init also keeps it ahead of the listener: a schema failure stops the
// release rather than surfacing on requests the instance already accepted.
//
// sql.Open is lazy, so Ping is what actually proves the DSN and the server; the
// migration would report the same failure, but only when migrations are on.
func setupDatabase(ctx context.Context, c DatabaseConfig) (*sql.DB, error) {
	if !c.Enabled {
		slog.InfoContext(ctx, "database is disabled")
		return nil, nil
	}

	db, err := sql.Open("pgx", c.DSN)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}
	if err := db.PingContext(ctx); err != nil {
		return nil, errors.Join(fmt.Errorf("ping database: %w", err), db.Close())
	}

	if !c.EnableMigration {
		slog.InfoContext(ctx, "database migration is disabled")
		return db, nil
	}
	if err := Migrate(ctx, db); err != nil {
		return nil, errors.Join(err, db.Close())
	}
	return db, nil
}

func Run() {
	tracer := SetupTracer()
	meter, metricsHandler, err := SetupMeter()
	if err != nil {
		slog.Error("meter provider setup failed", "error", err)
		return
	}

	opts := gor.NewServerOpts().
		WithPort(cfg.Server.Port).
		WithReadHeaderTimeout(time.Duration(cfg.Server.ReadHeaderTimeout)).
		WithReadTimeout(time.Duration(cfg.Server.ReadTimeout)).
		WithWriteTimeout(time.Duration(cfg.Server.WriteTimeout)).
		WithIdleTimeout(time.Duration(cfg.Server.IdleTimeout)).
		WithMaxHeaderBytes(1 << 20)

	g := gor.NewEngine(opts...).
		EnableSignals().
		EnableProbes().
		OnShutdownWithContext(func(ctx context.Context) {
			if err := tracer.Shutdown(ctx); err != nil {
				slog.ErrorContext(ctx, "tracer provider shutdown failed", "error", err)
			} else {
				slog.InfoContext(ctx, "tracer provider closed")
			}
		}).
		OnShutdownWithContext(func(ctx context.Context) {
			if err := meter.Shutdown(ctx); err != nil {
				slog.ErrorContext(ctx, "meter provider shutdown failed", "error", err)
			} else {
				slog.InfoContext(ctx, "meter provider closed")
			}
		})

	if db != nil {
		g.OnShutdownWithContext(func(ctx context.Context) {
			if err := db.Close(); err != nil {
				slog.ErrorContext(ctx, "database close failed", "error", err)
			} else {
				slog.InfoContext(ctx, "database closed")
			}
		})
	}

	g.NewRouter("").
		HandleHTTP("GET /metrics", metricsHandler).
		HandleHTTP("/", http.FileServer(http.Dir("./static/")))

	v1 := g.NewRouter("/api/v1/")
	v1.
		UseErrorHandler(AppErrorHandler).
		Use(
			otelhttp.NewMiddleware(AppName),
			// https://go.dev/blog/defer-panic-and-recover
			gor.RecoveryMiddleware,
			gor.ResponseWriterStatusCodeMiddleware,
			gor.SimpleLoggingMiddleware,
			gor.BasicMiddleware,
			gor.AuthenticationMiddleware(UsernamePasswordAutenticator("admin:admin")),
		)

	h := H{NewService(new(Store))}
	v1.
		With(Authorize("admin")).
		// all authorized by role admin
		Delete("/user/{id}", h.DeleteUser).
		Put("/user", h.EditUser).
		Post("/user", UserCounter(h.AddUser)) // same as "POST /user"
	v1.
		// without authorization
		Get("/user/me", h.Me). // same as "GET /user/me"
		Get("/user/{id}", h.GetUser).
		Get("/user", h.SearchUser).
		//
		Get("/hello", Hello).
		Get("/trace", GetTrace).
		HandleHTTPFunc("GET /ws", WSHandler).
		HandleHTTPFunc("/", DefaultHandler)

	if err := g.Listen(); err != nil {
		slog.Error("app stopped with an error", "error", err)
	}
}
