package postgres

import (
	"github.com/jackc/pgx/v5/pgxpool"
	pgxadapter "github.com/pckhoi/casbin-pgx-adapter/v3"
	"github.com/ponix-dev/ponix/internal/telemetry/stacktrace"
)

// NewCasbinAdapter creates a new Casbin adapter for PostgreSQL using the provided connection pool.
func NewCasbinAdapter(pool *pgxpool.Pool) (*pgxadapter.Adapter, error) {
	connStr := pool.Config().ConnString()
	a, err := pgxadapter.NewAdapter(connStr, pgxadapter.WithTableName("casbin_rule"))
	if err != nil {
		return nil, stacktrace.NewStackTraceErrorf("failed to create pgx adapter: %w", err)
	}

	return a, nil
}
