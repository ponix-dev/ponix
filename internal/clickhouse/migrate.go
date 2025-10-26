package clickhouse

import (
	"context"
	"database/sql"
	"embed"
	"log/slog"

	_ "github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ponix-dev/ponix/internal/telemetry/stacktrace"
	"github.com/pressly/goose/v3"
)

//go:embed goose/*.sql
var migrations embed.FS

func RunMigrations(ctx context.Context, connUrl string) error {
	db, err := sql.Open("clickhouse", connUrl)
	if err != nil {
		return stacktrace.NewStackTraceError(err)
	}

	defer db.Close()

	slog.Info("applying goose migrations for clickhouse")

	err = goose.SetDialect("clickhouse")
	if err != nil {
		return stacktrace.NewStackTraceError(err)
	}

	goose.SetBaseFS(migrations)

	err = goose.UpContext(ctx, db, "goose")
	if err != nil {
		return stacktrace.NewStackTraceError(err)
	}

	return nil
}
