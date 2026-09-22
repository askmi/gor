package internal

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"log/slog"

	"github.com/pressly/goose/v3"
)

// migrations holds the DDL. Embedding keeps the schema in the binary, so a
// deployment carries the migrations that match its code and needs no files
// alongside it.
//
//go:embed migrations/*.sql
var migrations embed.FS

// Migrate applies every pending migration, in version order, and reports the
// schema version it settled on.
//
// Migrating at startup suits a single-instance deployment. Concurrent instances
// would race, and goose takes no lock by default, so set database.enableMigration
// to false and apply the schema as its own deployment step instead.
func Migrate(ctx context.Context, db *sql.DB) error {
	goose.SetBaseFS(migrations)
	goose.SetLogger(gooseLogger{})

	if err := goose.SetDialect(string(goose.DialectPostgres)); err != nil {
		return fmt.Errorf("migrate dialect: %w", err)
	}
	if err := goose.UpContext(ctx, db, "migrations"); err != nil {
		return fmt.Errorf("migrate up: %w", err)
	}

	version, err := goose.GetDBVersionContext(ctx, db)
	if err != nil {
		return fmt.Errorf("migrate version: %w", err)
	}
	slog.InfoContext(ctx, "database schema is up to date", "version", version)
	return nil
}

// gooseLogger routes goose's own output through slog, so migration progress
// lands in the same stream, and the same format, as the rest of the logs.
type gooseLogger struct{}

func (gooseLogger) Printf(format string, v ...any) {
	slog.Info("goose: " + fmt.Sprintf(format, v...))
}

func (gooseLogger) Fatalf(format string, v ...any) {
	// goose calls Fatalf for errors it also returns; log and let the caller decide.
	slog.Error("goose: " + fmt.Sprintf(format, v...))
}
