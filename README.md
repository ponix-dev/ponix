# ponix

A monorepo for ponix software written in go

## Local Development Setup

### Prerequisites

- Go 1.24+
- Docker & Docker Compose
- Mage build tool
- Cloudflare account with tunnel credentials
- Tilt

### Docker Services

The local development environment uses Docker Compose to orchestrate the following services:

- **ponix-all-in-one**: Main application server (port 3001)
- **PostgreSQL**: Database server (port 5432)
- **NATS**: Message broker cluster (3 nodes)
- **InfluxDB**: Time-series database for metrics (port 8086)
- **Telegraf**: Metrics collector (reads from NATS, writes to InfluxDB)
- **Grafana LGTM Stack**: Observability platform (port 3000)
- **Cloudflared**: Cloudflare tunnel for secure external access

### Cloudflared Tunnel Configuration

The project uses Cloudflare tunnels to expose the local development API securely at `api.ponix.dev`.

#### Setup Requirements

1. **Credentials**: Place your Cloudflare tunnel credentials at:
   - `$HOME/.cloudflared/ponix-api.json` - Tunnel credentials
   - `$HOME/.cloudflared/cert.pem` - Cloudflare certificate

2. **Configuration**: The tunnel configuration is stored in `.cloudflared/cloudflared-config.yaml`:
   - Tunnel name: `ponix`
   - Hostname: `api.ponix.dev` → `ponix-all-in-one:3001`
   - Protocol: HTTP/2 with 2 HA connections
   - Metrics endpoint: `:8080`

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
   - API (tunnel): https://api.ponix.dev
   - Grafana: http://localhost:3000
   - InfluxDB: http://localhost:8086
   - NATS monitoring: http://localhost:8222

### Directory Structure

```
.cloudflared/          # Cloudflared tunnel configuration
.influxdb/             # InfluxDB and Telegraf configurations
cmd/                   # Application entry points
internal/              # Internal packages
schema/                # Database schema and queries
```

### Database Management

- Create migrations: `mage db:migrate <name>`
- Generate SQLC code: `mage db:gen`
- Schema location: `schema/schema.sql`

### Testing

```bash
go test ./...
```

### Environment Variables

Key environment variables are configured in `docker-compose.yaml`:
- Database connection settings
- TTN (The Things Network) configuration
- OpenTelemetry endpoints
- InfluxDB tokens and organization
