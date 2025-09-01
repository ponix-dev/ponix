# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

# Ponix Project - Claude Instructions

## Project Overview
Go-based monorepo for ponix IoT platform using Connect-RPC for API communication, PostgreSQL for data persistence, and Casbin for authorization.

## Key Technologies
- Go 1.24
- Connect-RPC (gRPC-compatible) with Protocol Buffers
- PostgreSQL with sqlc for type-safe SQL
- Atlas for database migrations
- Casbin for RBAC authorization
- OpenTelemetry for observability (logs, metrics, traces)
- Chi router for HTTP routing
- Docker Compose for local development
- NATS for messaging
- InfluxDB + Telegraf for metrics storage

## Build and Development Commands

### Mage Commands (primary build tool)
- `mage stack:up` - Start all Docker dependencies (PostgreSQL, NATS, InfluxDB, Grafana, etc.)
- `mage stack:down` - Stop all Docker dependencies
- `mage db:gen` - Generate database code with sqlc (run after modifying SQL files)
- `mage db:migrate <name>` - Create new database migration with Atlas

### Standard Go Commands
- `go run ./cmd/ponix-all-in-one` - Run the main application
- `go test ./...` - Run all tests
- `go test -run TestName ./package/...` - Run specific test
- `go mod tidy` - Clean up dependencies
- `go fmt ./...` - Format all Go code
- `go vet ./...` - Run static analysis

### Database Workflow
1. Modify schema in `schema/schema.sql` for tables/indexes
2. Add queries to entity-specific files under `schema/` (e.g., `schema/end_device.sql`)
3. If creating new query file, add it to `sqlc.yaml` queries section
4. Run `mage db:migrate <migration_name>` to create migration
5. Run `mage db:gen` to regenerate type-safe Go code

**File Locations:**
- Migrations: `internal/postgres/atlas/`
- Generated code: `internal/postgres/sqlc/`
- Schema definition: `schema/schema.sql`
- Query files: `schema/*.sql`

## High-Level Architecture

### Layered Architecture
```
┌─ cmd/ponix-all-in-one/         Entry point, service initialization
├─ internal/connectrpc/          RPC handlers (API layer)
├─ internal/domain/              Business logic and domain models
├─ internal/postgres/            Data persistence layer
│  ├─ atlas/                    Database migrations
│  └─ sqlc/                     Generated type-safe queries
├─ internal/casbin/              Authorization enforcement
└─ internal/telemetry/           OpenTelemetry instrumentation
```

### Core Services and Data Flow
1. **RPC Request** → ConnectRPC Handler
2. **Authentication** → Extract user from context (currently hardcoded as `dev-user-123`)
3. **Authorization** → Casbin enforcer checks permissions
4. **Business Logic** → Domain layer processes request
5. **Data Access** → PostgreSQL via SQLC-generated code
6. **Response** → ConnectRPC response with proper error codes

### Multi-Tenancy Model
- All entities scoped by `organization_id`
- Users can belong to multiple organizations with different roles
- Authorization enforced at organization level
- Role hierarchy: Super Admin > Org Admin > Org Member > Org Viewer

## Authorization System (Casbin)

### Authorization Pattern for RPC Services
```go
// Standard pattern in every RPC method:
func (s *Service) Method(ctx context.Context, req *Request) (*Response, error) {
    // 1. Extract organizationID from request
    orgID := req.GetOrganizationId()
    
    // 2. Authorize before business logic
    if err := s.authorizeRequest(ctx, "action", orgID); err != nil {
        return nil, err // Returns connect.CodePermissionDenied
    }
    
    // 3. Execute business logic
    // ...
}
```

### Action Mapping Convention
- `Create*` methods → `"create"` action
- `Get*`, `List*` methods → `"read"` action
- `Update*` methods → `"update"` action
- `Delete*` methods → `"delete"` action

### Casbin Model
- Format: `sub, obj, act, org` (subject, object, action, organization)
- Policies stored in PostgreSQL `casbin_rule` table
- Separate enforcers per domain (User, Organization, EndDevice, LoRaWAN)

### Context Requirements
- User ID extracted from request context (set by AuthenticationInterceptor)
- Organization ID extracted from request payload
- Enforcer checks: `CanAccess*(ctx, userID, action, organizationID)`

## Service Integration Points

### Connect-RPC Services
- **Organization Service** - Organization and user association management
- **User Service** - User CRUD operations
- **End Device Service** - IoT device management
- **LoRaWAN Service** - LoRaWAN-specific configurations

### External Integrations
- **The Things Network (TTN)** - LoRaWAN network server integration
- **NATS** - Message broker for webhook events
- **InfluxDB** - Time-series metrics storage via Telegraf

### Service Dependencies
- Services initialized in `cmd/ponix-all-in-one/main.go`
- Dependency injection via constructor parameters
- Shared database connection pool
- Common telemetry and authorization middleware

## Code Style Requirements

### Error Handling
```go
// CORRECT - Check error immediately
result, err := SomeFunction()
if err != nil {
    return nil, err
}

// INCORRECT - Never embed error check in declaration
if result, err := SomeFunction(); err != nil {
    return nil, err
}
```

### Code Quality
- Always run `go fmt ./...` before committing
- Run `go vet ./...` for static analysis
- Ensure proper OpenTelemetry spans for observability
- Use structured logging with `slog`
- Validate inputs using `protovalidate` annotations in proto files