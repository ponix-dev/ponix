# ponix

A monorepo for ponix software written in go

## Local Development Setup

### Prerequisites

- Go 1.24+
- Docker & Docker Compose
- Mage build tool
- Tilt

### Docker Services

The local development environment uses Docker Compose to orchestrate the following services:

- **ponix-all-in-one**: Main application server (port 3001)
- **PostgreSQL**: Relational database for metadata and authorization (port 5432)
- **ClickHouse**: Columnar database for time-series IoT data storage (ports 8123, 9000)
- **NATS**: Message broker with JetStream for event streaming (ports 4222, 8222, 6222)
- **Grafana LGTM Stack**: Observability platform with OpenTelemetry (port 3000)

### Quick Start

1. **Start all services**:
   ```bash
   tilt up
   ```

2. **Run the application** (if not using Docker):
   ```bash
   go run ./cmd/ponix-all-in-one
   ```

3. **Access the services**:
   - API (local): http://localhost:3001
   - Grafana (LGTM): http://localhost:3000
   - ClickHouse HTTP: http://localhost:8123
   - NATS monitoring: http://localhost:8222

### Directory Structure

```
cmd/                   # Application entry points
internal/
  ├─ clickhouse/       # ClickHouse connection and data storage
  ├─ nats/             # NATS JetStream producers and consumers
  ├─ postgres/         # PostgreSQL migrations and SQLC queries
  ├─ domain/           # Business logic and domain models
  └─ connectrpc/       # RPC service handlers
schema/
  ├─ postgres/         # PostgreSQL schema and queries
  └─ clickhouse/       # ClickHouse schema definitions
```

### Database Management

**PostgreSQL (Metadata & Authorization)**:
- Create migrations: `mage db:migrate <name>`
- Generate SQLC code: `mage db:gen`
- Schema location: `schema/postgres/schema.sql`

**ClickHouse (Time-Series IoT Data)**:
- Migrations auto-run on startup from `internal/clickhouse/goose/`
- Schema location: `schema/clickhouse/schema.sql`

### Testing

```bash
go test ./...
```

### Environment Variables

Key environment variables are configured in `docker-compose.yaml`:
- **PostgreSQL**: Database connection settings (relational data)
- **ClickHouse**: Time-series database connection (IoT data)
- **NATS**: JetStream configuration for event streaming
- **TTN**: The Things Network integration settings
- **OpenTelemetry**: OTLP endpoint for observability
