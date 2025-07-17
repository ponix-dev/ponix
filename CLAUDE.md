# Ponix Project - Claude Instructions

## Project Overview
This is a Go-based monorepo for ponix software using gRPC/Connect-RPC for communication and PostgreSQL for data persistence.

## Key Technologies
- Go 1.24
- Connect-RPC for API communication
- PostgreSQL with sqlc for type-safe SQL
- Atlas for database migrations
- OpenTelemetry for observability
- Chi router for HTTP routing
- Docker Compose for local development

## Build and Development Commands

### Mage Commands (primary build tool)
- `mage stack:up` - Start Docker dependencies (PostgreSQL, etc.)
- `mage stack:down` - Stop Docker dependencies
- `mage db:gen` - Generate database code with sqlc
- `mage db:migrate <name>` - Create new database migration

### Standard Go Commands
- `go run ./cmd/ponix-all-in-one` - Run the main application
- `go test ./...` - Run all tests
- `go mod tidy` - Clean up dependencies

### Database Operations
- Database migrations are in `internal/postgres/atlas/`
- SQL schema files are in `schema/`
- Generated database code is in `internal/postgres/sqlc/`
- Table and index sql operations should go in `schema/schema.sql`
- Queries should go in to specific entity named files under `schema`
- Whenever we add a new file under `schema` for queries, they need to be added to our `sqlc.yaml` file

## Project Structure
- `cmd/` - Application entry points
- `internal/` - Private application code
  - `connectrpc/` - Connect-RPC service implementations
  - `domain/` - Business domain models
  - `postgres/` - Database layer with migrations and generated code
  - `telemetry/` - OpenTelemetry instrumentation
- `schema/` - SQL schema definitions

## Development Workflow
1. Start dependencies: `mage stack:up`
2. Generate database code after schema changes: `mage db:gen`
3. Create migrations for schema changes: `mage db:migrate <migration_name>`
4. Run application: `go run ./cmd/ponix-all-in-one`

## Testing
Run tests with: `go test ./...`

## Authorization Pattern for RPC Services

### Service Structure
- Inject `*auth.Enforcer` into service constructors
- Create authorization helper: `authorizeRequest(ctx, action, orgID)`

### RPC Method Pattern
1. Extract `organizationID` from request
2. Call `authorizeRequest(ctx, action, orgID)` before business logic
3. Return `connect.CodePermissionDenied` if unauthorized

### Action Mapping
- `Create*` → `"create"`
- `Get*`, `List*` → `"read"`
- `Update*` → `"update"`
- `Delete*` → `"delete"`

### Context Requirements
- Extract `userID` from JWT/session in request context
- Use `enforcer.CanAccessEndDevice(ctx, userID, action, organizationID)`

## Code Quality
Always run `go fmt` and `go vet` before committing changes.

Never `if user if var1, err := SomeFunc(); err != nil {}` always do `if err != nil {}` after calling the function if it returns an error