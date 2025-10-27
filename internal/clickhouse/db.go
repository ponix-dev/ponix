package clickhouse

import (
	"context"
	"fmt"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/ponix-dev/ponix/internal/telemetry/stacktrace"
)

func NewUrl(database string, username string, password string, addr string) string {
	return fmt.Sprintf("clickhouse://%s:%s@%s/%s", username, password, addr, database)
}

func NewConnection(ctx context.Context, database string, username string, password string, addr string) (driver.Conn, error) {
	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{addr},
		Auth: clickhouse.Auth{
			Database: database,
			Username: username,
			Password: password,
		},
	})
	if err != nil {
		return nil, stacktrace.NewStackTraceError(err)
	}

	err = conn.Ping(ctx)
	if err != nil {
		return nil, stacktrace.NewStackTraceError(err)
	}

	return conn, nil
}
