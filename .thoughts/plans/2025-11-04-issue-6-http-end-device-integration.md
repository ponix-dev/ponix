# HTTP End Device Integration Implementation Plan

## Overview

Add support for HTTP-based end device data ingestion to complement the existing LoRaWAN integration. This enables devices that communicate via HTTP to send telemetry data to Ponix using ConnectRPC, broadening platform compatibility and simplifying end-to-end testing with mock devices.

## Current State Analysis

After thorough research of the codebase:
- **All APIs are ConnectRPC-based** - No raw HTTP webhooks, consistent architecture (main.go:203-261)
- **hardware_type field** already serves as integration type discriminator (schema/postgres/schema.sql:74)
- **ProcessedEnvelope pipeline** fully operational through NATS → ClickHouse
- **LoRaWAN-only** implementation with hardcoded checks (internal/domain/end_device.go:98-103)
- **No device validation** in IngestDataEnvelope - currently trusts organization_id from caller

### Key Discoveries:
- The `hardware_type` enum in end_devices table is the integration type concept
- System uses polymorphic configuration pattern with type-specific tables
- DataEnvelopeManager.IngestDataEnvelope() doesn't validate device exists or belongs to organization
- NATS subject pattern includes device ID for routing: `processed_envelopes.{device_id}`
- ConnectRPC endpoints provide consistent middleware (auth, validation, tracing)

## Desired End State

After implementation:
- ConnectRPC endpoint for ingesting device data (e.g., `IngestDeviceData`)
- HTTP devices stored in end_devices table with `hardware_type = END_DEVICE_HARDWARE_TYPE_HTTP`
- Domain validation ensures device exists and belongs to specified organization before ingestion
- HTTP device data flowing through same NATS → ClickHouse pipeline as LoRaWAN
- Ability to create HTTP end devices via RPC API
- Mock devices can send test data via ConnectRPC client

### Verification:
- Create an HTTP end device via CreateEndDevice RPC
- Call IngestDeviceData RPC with device telemetry
- Attempt to ingest data for non-existent device (should fail)
- Attempt to ingest data for device in different organization (should fail)
- Query data via QueryEndDeviceData RPC to confirm storage

## What We're NOT Doing

- Authentication/authorization for ingestion endpoint (uses existing interceptors)
- Payload transformation or validation beyond JSON parsing
- Device type templates or payload mapping (issue #5)
- Modification of existing LoRaWAN functionality
- Changes to ProcessedEnvelope structure or NATS/ClickHouse flow

## Additional Changes

- **Deprecating `data_type` field**: The `EndDeviceDataType` enum is not compatible with flexible JSON payloads. The field will be marked as deprecated in protobuf and made nullable in the database. Existing devices will retain their data_type value, but new devices don't need to set it.

## Implementation Approach

1. Add ConnectRPC endpoint for data ingestion (consistent with existing architecture)
2. Add device validation in domain layer to prevent unauthorized ingestion
3. Extend polymorphic device configuration pattern to support HTTP hardware type
4. Reuse existing ProcessedEnvelope pipeline without modifications

## Phase 1: Protobuf & Database Schema Updates

### Overview
Add HTTP device support to data model and define ingestion RPC contract.

### Changes Required:

#### 1. Protobuf Definition Updates
**Note**: Protobuf definitions are in external repository at `buf.build/ponix/ponix`. These changes need to be made there first, then update go.mod dependency.

**File**: `iot/v1/end_device.proto`

Add HTTP hardware type:
```protobuf
enum EndDeviceHardwareType {
  END_DEVICE_HARDWARE_TYPE_UNSPECIFIED = 0;
  END_DEVICE_HARDWARE_TYPE_LORAWAN = 1;
  END_DEVICE_HARDWARE_TYPE_HTTP = 2;  // New addition
}
```

Deprecate `data_type` field (not compatible with flexible JSON payloads):
```protobuf
message EndDevice {
  string id = 1 [(buf.validate.field).required = true];
  string name = 2 [(buf.validate.field).required = true];
  EndDeviceStatus status = 3 [(buf.validate.field).required = true];
  EndDeviceDataType data_type = 4 [deprecated = true];  // Deprecated: use flexible data field in ingestion
  EndDeviceHardwareType hardware_type = 5 [(buf.validate.field).required = true];
  string description = 6;

  // Hardware-specific configuration
  oneof hardware_config {
    iot.v1.LoRaWANConfig lorawan_config = 7;
  }
}

message CreateEndDeviceRequest {
  string name = 1 [(buf.validate.field).required = true];
  string description = 2;
  string hardware_type_id = 3 [(buf.validate.field).required = true];
  EndDeviceHardwareType hardware_type = 4 [(buf.validate.field).required = true];
  EndDeviceDataType data_type = 5 [deprecated = true];  // Deprecated: not required for HTTP devices
}
```

Add new service to `iot/v1/ingestion.proto` (new file):
```protobuf
syntax = "proto3";

package iot.v1;

import "google/protobuf/struct.proto";
import "google/protobuf/timestamp.proto";

// DataIngestionService handles device data ingestion
service DataIngestionService {
  // IngestDeviceData accepts telemetry data from devices
  rpc IngestDeviceData(IngestDeviceDataRequest) returns (IngestDeviceDataResponse);
}

message IngestDeviceDataRequest {
  // Organization ID that owns the device
  string organization_id = 1;

  // End device ID sending the data
  string end_device_id = 2;

  // Timestamp when data was collected (optional, defaults to server time)
  google.protobuf.Timestamp occurred_at = 3;

  // Flexible JSON data payload
  google.protobuf.Struct data = 4;
}

message IngestDeviceDataResponse {
  // Success confirmation
  bool success = 1;

  // Optional message
  string message = 2;
}
```

#### 2. Update go.mod Dependency
**File**: `go.mod`

After protobuf changes are published, run:
```bash
go get buf.build/gen/go/ponix/ponix/protocolbuffers/go@latest
go mod tidy
```

#### 3. Database Migration
**File**: Create `internal/postgres/goose/YYYYMMDDHHMMSS_add_http_device_support.sql`

```sql
-- +goose Up
-- Migration to add HTTP device configuration support
-- HTTP devices don't need a separate config table as they have no special fields

-- Add constraint to validate hardware_type values
ALTER TABLE end_devices
ADD CONSTRAINT check_hardware_type CHECK (hardware_type IN (0, 1, 2));

-- Create index for querying HTTP devices
CREATE INDEX IF NOT EXISTS idx_end_devices_hardware_type_http
ON end_devices(organization_id, hardware_type)
WHERE hardware_type = 2;

-- Make data_type nullable (deprecated field, not meaningful with flexible JSON payloads)
ALTER TABLE end_devices
ALTER COLUMN data_type DROP NOT NULL;

-- +goose Down
-- Reverse data_type nullability
ALTER TABLE end_devices
ALTER COLUMN data_type SET NOT NULL;

DROP INDEX IF EXISTS idx_end_devices_hardware_type_http;
ALTER TABLE end_devices DROP CONSTRAINT IF EXISTS check_hardware_type;
```

### Success Criteria:

#### Automated Verification:
- [x] Protobuf compiles and generates Go code: `buf generate`
- [x] Application compiles with new protobufs: `go build ./cmd/ponix-all-in-one`
- [ ] Database migration applies cleanly: `go run ./cmd/ponix-all-in-one` (migrations run on startup)
- [ ] Migration rollback works: Test down migration

#### Manual Verification:
- [ ] Generated RPC service code exists for DataIngestionService
- [ ] PostgreSQL schema updated with constraint and index
- [ ] Can insert test HTTP device record directly in PostgreSQL
- [ ] Existing LoRaWAN devices unaffected

---

## Phase 2: Domain Layer - Device Validation

### Overview
Add validation in DataEnvelopeManager to verify device exists and belongs to organization before ingestion.

### Changes Required:

#### 1. Add EndDevice Store Interface Method
**File**: `internal/domain/end_device.go` (add to EndDeviceStorer interface around line 18)

```go
type EndDeviceStorer interface {
    AddEndDevice(ctx context.Context, endDevice *iotv1.EndDevice, organizationID string) error
    // Add this method:
    GetEndDeviceWithOrganization(ctx context.Context, endDeviceID string) (*iotv1.EndDevice, string, error)
}
```

#### 2. Implement SQL Query
**File**: `schema/postgres/end_device.sql` (add new query)

```sql
-- name: GetEndDeviceWithOrganization :one
SELECT id, name, description, organization_id, status, data_type, hardware_type, created_at, updated_at
FROM end_devices
WHERE id = $1;
```

#### 3. Regenerate SQLC Code
Run after adding query:
```bash
mage db:gen
```

#### 4. Implement Store Method
**File**: `internal/postgres/end_device.go` (add new method)

```go
// GetEndDeviceWithOrganization retrieves an end device and its organization ID
func (store *EndDeviceStore) GetEndDeviceWithOrganization(ctx context.Context, endDeviceID string) (*iotv1.EndDevice, string, error) {
    ctx, span := telemetry.Tracer().Start(ctx, "GetEndDeviceWithOrganization")
    defer span.End()

    queries := sqlc.New(store.pool)

    endDeviceRow, err := queries.GetEndDeviceWithOrganization(ctx, endDeviceID)
    if err != nil {
        if errors.Is(err, pgx.ErrNoRows) {
            return nil, "", stacktrace.NewStackTraceErrorf("end device not found: %s", endDeviceID)
        }
        return nil, "", stacktrace.NewStackTraceError(err)
    }

    // Build protobuf EndDevice from row (minimal version without config)
    endDeviceBuilder := iotv1.EndDevice_builder{
        Id:           endDeviceRow.ID,
        Name:         endDeviceRow.Name,
        Status:       iotv1.EndDeviceStatus(endDeviceRow.Status),
        HardwareType: iotv1.EndDeviceHardwareType(endDeviceRow.HardwareType),
    }

    if endDeviceRow.Description.Valid {
        endDeviceBuilder.Description = endDeviceRow.Description.String
    }

    // Data type is deprecated but may still be set on older devices
    if endDeviceRow.DataType.Valid {
        endDeviceBuilder.DataType = iotv1.EndDeviceDataType(endDeviceRow.DataType.Int32)
    }

    return endDeviceBuilder.Build(), endDeviceRow.OrganizationID, nil
}
```

#### 5. Add Validation to DataEnvelopeManager
**File**: `internal/domain/data_envelope.go` (update IngestDataEnvelope around line 38)

```go
type DataEnvelopeManager struct {
    producer         ProcessedEnvelopeProducer
    store            ProcessedEnvelopeStorer
    endDeviceStore   EndDeviceStorer  // Add this dependency
}

// Update constructor
func NewDataEnvelopeManager(
    producer ProcessedEnvelopeProducer,
    store ProcessedEnvelopeStorer,
    endDeviceStore EndDeviceStorer,  // Add parameter
) *DataEnvelopeManager {
    return &DataEnvelopeManager{
        producer:       producer,
        store:          store,
        endDeviceStore: endDeviceStore,
    }
}

func (mgr *DataEnvelopeManager) IngestDataEnvelope(ctx context.Context, envelope *envelopev1.DataEnvelope, organizationID string) error {
    ctx, span := telemetry.Tracer().Start(ctx, "IngestDataEnvelope")
    defer span.End()

    // VALIDATION: Verify device exists and belongs to organization
    device, deviceOrgID, err := mgr.endDeviceStore.GetEndDeviceWithOrganization(ctx, envelope.GetEndDeviceId())
    if err != nil {
        return stacktrace.NewStackTraceErrorf("device validation failed: %w", err)
    }

    if deviceOrgID != organizationID {
        return stacktrace.NewStackTraceErrorf(
            "organization mismatch: device %s belongs to %s, but data sent for %s",
            envelope.GetEndDeviceId(),
            deviceOrgID,
            organizationID,
        )
    }

    // Build ProcessedEnvelope with validation complete
    processedEnvelope := envelopev1.ProcessedEnvelope_builder{
        OrganizationId: organizationID,
        EndDeviceId:    envelope.GetEndDeviceId(),
        OccurredAt:     envelope.GetOccurredAt(),
        Data:           envelope.GetData(),
        ProcessedAt:    timestamppb.New(time.Now().UTC()),
    }.Build()

    // Publish to NATS
    err = mgr.producer.ProduceProcessedEnvelope(ctx, processedEnvelope)
    if err != nil {
        return stacktrace.NewStackTraceError(err)
    }

    return nil
}
```

#### 6. Update Main Application Wiring
**File**: `cmd/ponix-all-in-one/main.go` (update DataEnvelopeManager creation around line 165)

```go
// Create data envelope manager with device validation
envelopeMgr := domain.NewDataEnvelopeManager(
    processedEnvelopeProducer,
    envelopeStore,
    edStore,  // Add end device store for validation
)
```

### Success Criteria:

#### Automated Verification:
- [x] SQLC generates code: `mage db:gen`
- [x] Application compiles: `go build ./cmd/ponix-all-in-one`
- [ ] Unit test for validation passes: `go test ./internal/domain/...`

#### Manual Verification:
- [ ] Attempting to ingest data for non-existent device fails with clear error
- [ ] Attempting to ingest data with wrong organization_id fails
- [ ] Valid ingestion still succeeds

---

## Phase 3: ConnectRPC Ingestion Handler

### Overview
Create ConnectRPC handler for device data ingestion.

### Changes Required:

#### 1. Create Ingestion Handler
**File**: `internal/connectrpc/ingestion.go` (new file)

```go
package connectrpc

import (
    "context"
    "fmt"

    envelopev1 "buf.build/gen/go/ponix/ponix/protocolbuffers/go/envelope/v1"
    iotv1 "buf.build/gen/go/ponix/ponix/protocolbuffers/go/iot/v1"
    "connectrpc.com/connect"
    "github.com/ponix-dev/ponix/internal/domain"
    "github.com/ponix-dev/ponix/internal/telemetry"
    "google.golang.org/protobuf/types/known/timestamppb"
    "time"
)

// DataEnvelopeManager handles device data ingestion operations
type DataEnvelopeManager interface {
    IngestDataEnvelope(ctx context.Context, envelope *envelopev1.DataEnvelope, organizationID string) error
}

// IngestionHandler implements ConnectRPC handlers for device data ingestion
type IngestionHandler struct {
    envelopeManager DataEnvelopeManager
}

// NewIngestionHandler creates a new ingestion handler
func NewIngestionHandler(envelopeManager DataEnvelopeManager) *IngestionHandler {
    return &IngestionHandler{
        envelopeManager: envelopeManager,
    }
}

// IngestDeviceData handles RPC requests to ingest device telemetry data
// No authorization required for MVP (future enhancement)
func (handler *IngestionHandler) IngestDeviceData(
    ctx context.Context,
    req *connect.Request[iotv1.IngestDeviceDataRequest],
) (*connect.Response[iotv1.IngestDeviceDataResponse], error) {
    ctx, span := telemetry.Tracer().Start(ctx, "IngestDeviceData")
    defer span.End()

    // Validate required fields
    if req.Msg.GetOrganizationId() == "" {
        return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("organization_id is required"))
    }
    if req.Msg.GetEndDeviceId() == "" {
        return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("end_device_id is required"))
    }
    if req.Msg.GetData() == nil {
        return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("data is required"))
    }

    // Use provided occurred_at or default to now
    occurredAt := req.Msg.GetOccurredAt()
    if occurredAt == nil {
        occurredAt = timestamppb.New(time.Now().UTC())
    }

    // Build DataEnvelope
    envelope := envelopev1.DataEnvelope_builder{
        EndDeviceId: req.Msg.GetEndDeviceId(),
        OccurredAt:  occurredAt,
        Data:        req.Msg.GetData(),
    }.Build()

    // Ingest with validation (checks device exists and belongs to org)
    err := handler.envelopeManager.IngestDataEnvelope(ctx, envelope, req.Msg.GetOrganizationId())
    if err != nil {
        telemetry.RecordError(span, err)
        return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("ingestion failed: %w", err))
    }

    // Return success
    return connect.NewResponse(iotv1.IngestDeviceDataResponse_builder{
        Success: true,
        Message: "Data ingested successfully",
    }.Build()), nil
}
```

#### 2. Register Handler in Main Application
**File**: `cmd/ponix-all-in-one/main.go` (add after line 241)

```go
// Create ingestion handler
ingestionHandler := connectrpc.NewIngestionHandler(envelopeMgr)

// Update server creation
srv, err := mux.New(
    mux.WithPort(cfg.ServerPort),
    mux.WithLogger(logger),
    // ... existing handlers ...

    // Add data ingestion handler (no auth interceptors for MVP)
    mux.WithHandler(iotv1connect.NewDataIngestionServiceHandler(
        ingestionHandler,
        connect.WithInterceptors(protovalidateInterceptor),
    )),
)
```

### Success Criteria:

#### Automated Verification:
- [x] Application compiles: `go build ./cmd/ponix-all-in-one`
- [ ] Handler unit tests pass: `go test ./internal/connectrpc/...`
- [ ] Application starts without errors: `go run ./cmd/ponix-all-in-one`

#### Manual Verification:
- [ ] IngestDeviceData endpoint accessible via Connect client
- [ ] Returns proper error for missing required fields
- [ ] Returns proper error for non-existent device
- [ ] Returns proper error for organization mismatch
- [ ] Successful ingestion returns 200 with success message

---

## Phase 4: EndDevice CRUD Extensions

### Overview
Enable creation and management of HTTP end devices via RPC.

### Changes Required:

#### 1. Update EndDevice Domain Manager
**File**: `internal/domain/end_device.go`

Update `buildEndDeviceFromRequest` method (around lines 98-103):

```go
func (mgr *EndDeviceManager) buildEndDeviceFromRequest(ctx context.Context, endDeviceId string, createReq *iotv1.CreateEndDeviceRequest) (*iotv1.EndDevice, error) {
    endDeviceBuilder := iotv1.EndDevice_builder{
        Id:           endDeviceId,
        Name:         createReq.GetName(),
        Description:  createReq.GetDescription(),
        Status:       iotv1.EndDeviceStatus_END_DEVICE_STATUS_PENDING,
        HardwareType: createReq.GetHardwareType(),
        // Note: data_type is deprecated and not set
    }

    switch createReq.GetHardwareType() {
    case iotv1.EndDeviceHardwareType_END_DEVICE_HARDWARE_TYPE_LORAWAN:
        lorawanConfig, err := mgr.buildLoRaWANConfig(ctx, createReq)
        if err != nil {
            return nil, stacktrace.NewStackTraceError(err)
        }
        endDeviceBuilder.LorawanConfig = lorawanConfig
    case iotv1.EndDeviceHardwareType_END_DEVICE_HARDWARE_TYPE_HTTP:
        // HTTP devices don't need additional configuration
        // Just validate and continue
    default:
        return nil, stacktrace.NewStackTraceErrorf("unsupported hardware type: %v", createReq.GetHardwareType())
    }

    return endDeviceBuilder.Build(), nil
}
```

Update `CreateEndDevice` method (around lines 65-68):

```go
func (mgr *EndDeviceManager) CreateEndDevice(ctx context.Context, createReq *iotv1.CreateEndDeviceRequest, organizationId string) (*iotv1.EndDevice, error) {
    // ... existing validation and ID generation ...

    endDevice, err := mgr.buildEndDeviceFromRequest(ctx, endDeviceId, createReq)
    if err != nil {
        return nil, stacktrace.NewStackTraceError(err)
    }

    // Only register with external systems for LoRaWAN devices
    switch endDevice.GetHardwareType() {
    case iotv1.EndDeviceHardwareType_END_DEVICE_HARDWARE_TYPE_LORAWAN:
        err = mgr.endDeviceRegister.RegisterEndDevice(ctx, endDevice)
        if err != nil {
            return nil, stacktrace.NewStackTraceError(err)
        }
    case iotv1.EndDeviceHardwareType_END_DEVICE_HARDWARE_TYPE_HTTP:
        // HTTP devices don't need external registration
        // Continue to storage
    }

    // Store device in database
    err = mgr.endDeviceStore.AddEndDevice(ctx, endDevice, organizationId)
    if err != nil {
        return nil, stacktrace.NewStackTraceError(err)
    }

    return endDevice, nil
}
```

#### 2. Update PostgreSQL Store
**File**: `internal/postgres/end_device.go`

Update `AddEndDevice` method (around lines 67-89):

```go
func (store *EndDeviceStore) AddEndDevice(ctx context.Context, endDevice *iotv1.EndDevice, organizationID string) error {
    ctx, span := telemetry.Tracer().Start(ctx, "AddEndDevice")
    defer span.End()

    tx, err := store.pool.Begin(ctx)
    if err != nil {
        return stacktrace.NewStackTraceError(err)
    }
    defer tx.Rollback(ctx)

    txQueries := sqlc.New(tx)

    // Insert base end device
    endDeviceParams := sqlc.CreateEndDeviceParams{
        ID:             endDevice.GetId(),
        Name:           endDevice.GetName(),
        Description:    pgtype.Text{String: endDevice.GetDescription(), Valid: endDevice.GetDescription() != ""},
        OrganizationID: organizationID,
        Status:         int32(endDevice.GetStatus()),
        DataType:       pgtype.Int4{Int32: int32(endDevice.GetDataType()), Valid: false}, // Deprecated field
        HardwareType:   int32(endDevice.GetHardwareType()),
    }

    _, err = txQueries.CreateEndDevice(ctx, endDeviceParams)
    if err != nil {
        return stacktrace.NewStackTraceError(err)
    }

    // Add hardware-specific configuration if needed
    switch endDevice.GetHardwareType() {
    case iotv1.EndDeviceHardwareType_END_DEVICE_HARDWARE_TYPE_LORAWAN:
        lorawanConfig := endDevice.GetLorawanConfig()
        if lorawanConfig == nil {
            return stacktrace.NewStackTraceErrorf("LoRaWAN device requires lorawan_config")
        }

        lorawanParams := sqlc.CreateLoRaWANConfigParams{
            // ... existing LoRaWAN params ...
        }
        _, err = txQueries.CreateLoRaWANConfig(ctx, lorawanParams)
        if err != nil {
            return stacktrace.NewStackTraceError(err)
        }
    case iotv1.EndDeviceHardwareType_END_DEVICE_HARDWARE_TYPE_HTTP:
        // HTTP devices don't have additional config tables
        // No additional operations needed
    default:
        return stacktrace.NewStackTraceErrorf("unsupported hardware type: %v", endDevice.GetHardwareType())
    }

    return tx.Commit(ctx)
}
```

### Success Criteria:

#### Automated Verification:
- [x] Application compiles: `go build ./cmd/ponix-all-in-one`
- [ ] Domain tests pass: `go test ./internal/domain/...`
- [ ] PostgreSQL store tests pass: `go test ./internal/postgres/...`

#### Manual Verification:
- [ ] Can create HTTP end device via CreateEndDevice RPC
- [ ] HTTP device appears in database with hardware_type = 2
- [ ] No LoRaWAN config created for HTTP devices
- [ ] Existing LoRaWAN device creation still works

---

## Phase 5: Integration & End-to-End Testing

### Overview
Validate the complete flow from device creation to data ingestion and querying.

### Changes Required:

#### 1. Create Integration Test
**File**: `test/integration/http_device_test.go`

```go
package integration

import (
    "context"
    "net/http"
    "testing"
    "time"

    iotv1 "buf.build/gen/go/ponix/ponix/protocolbuffers/go/iot/v1"
    "connectrpc.com/connect"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "google.golang.org/protobuf/types/known/structpb"
    "google.golang.org/protobuf/types/known/timestamppb"
)

func TestHTTPDeviceIntegration(t *testing.T) {
    ctx := context.Background()
    orgID := "test-org-123"
    baseURL := "http://localhost:8080"

    // Step 1: Create HTTP end device
    deviceClient := iotv1connect.NewEndDeviceServiceClient(http.DefaultClient, baseURL)

    createReq := connect.NewRequest(&iotv1.CreateEndDeviceRequest{
        Name:           "Test HTTP Device",
        Description:    "Integration test device",
        OrganizationId: orgID,
        HardwareType:   iotv1.EndDeviceHardwareType_END_DEVICE_HARDWARE_TYPE_HTTP,
        HardwareTypeId: "http-generic-v1", // Generic HTTP device type
        // Note: data_type is deprecated and not set
    })

    createResp, err := deviceClient.CreateEndDevice(ctx, createReq)
    require.NoError(t, err)
    assert.NotEmpty(t, createResp.Msg.GetEndDevice().GetId())
    deviceID := createResp.Msg.GetEndDevice().GetId()

    // Step 2: Ingest device data
    ingestionClient := iotv1connect.NewDataIngestionServiceClient(http.DefaultClient, baseURL)

    telemetryData, _ := structpb.NewStruct(map[string]interface{}{
        "temperature": 23.5,
        "humidity":    65.2,
        "battery":     87.0,
    })

    ingestReq := connect.NewRequest(&iotv1.IngestDeviceDataRequest{
        OrganizationId: orgID,
        EndDeviceId:    deviceID,
        Data:           telemetryData,
    })

    ingestResp, err := ingestionClient.IngestDeviceData(ctx, ingestReq)
    require.NoError(t, err)
    assert.True(t, ingestResp.Msg.GetSuccess())

    // Step 3: Test validation - wrong organization
    wrongOrgReq := connect.NewRequest(&iotv1.IngestDeviceDataRequest{
        OrganizationId: "wrong-org-id",
        EndDeviceId:    deviceID,
        Data:           telemetryData,
    })

    _, err = ingestionClient.IngestDeviceData(ctx, wrongOrgReq)
    assert.Error(t, err)
    assert.Equal(t, connect.CodeInvalidArgument, connect.CodeOf(err))

    // Step 4: Test validation - non-existent device
    nonExistentReq := connect.NewRequest(&iotv1.IngestDeviceDataRequest{
        OrganizationId: orgID,
        EndDeviceId:    "non-existent-device",
        Data:           telemetryData,
    })

    _, err = ingestionClient.IngestDeviceData(ctx, nonExistentReq)
    assert.Error(t, err)

    // Step 5: Wait for NATS processing
    time.Sleep(2 * time.Second)

    // Step 6: Query device data to verify storage
    dataClient := iotv1connect.NewEndDeviceDataServiceClient(http.DefaultClient, baseURL)

    queryReq := connect.NewRequest(&iotv1.QueryEndDeviceDataRequest{
        OrganizationId: orgID,
        EndDeviceIds:   []string{deviceID},
        FieldPath:      "temperature",
        StartTime:      timestamppb.New(time.Now().Add(-5 * time.Minute)),
        EndTime:        timestamppb.New(time.Now().Add(1 * time.Minute)),
    })

    queryResp, err := dataClient.QueryEndDeviceData(ctx, queryReq)
    require.NoError(t, err)
    assert.NotEmpty(t, queryResp.Msg.GetHistogram())
    assert.Greater(t, queryResp.Msg.GetHistogram().GetSampleCount(), uint64(0))
}
```

#### 2. Add Manual Testing Script
**File**: `scripts/test_http_device.sh`

```bash
#!/bin/bash

# Test HTTP device end-to-end flow with ConnectRPC

ORG_ID="test-org-123"
BASE_URL="http://localhost:8080"

echo "=== HTTP Device Integration Test ==="

# Test 1: Create HTTP device
echo -e "\n1. Creating HTTP end device..."
grpcurl -plaintext -d '{
  "name": "Test HTTP Device",
  "description": "Manual test device",
  "organization_id": "'$ORG_ID'",
  "hardware_type": "END_DEVICE_HARDWARE_TYPE_HTTP",
  "hardware_type_id": "http-generic-v1"
}' localhost:8080 iot.v1.EndDeviceService/CreateEndDevice

# Extract device ID from response (manual step)
read -p "Enter created device ID: " DEVICE_ID

# Test 2: Ingest valid data
echo -e "\n2. Ingesting device data (valid)..."
grpcurl -plaintext -d '{
  "organization_id": "'$ORG_ID'",
  "end_device_id": "'$DEVICE_ID'",
  "data": {
    "temperature": 22.5,
    "humidity": 60.0,
    "pressure": 1013.25
  }
}' localhost:8080 iot.v1.DataIngestionService/IngestDeviceData

# Test 3: Ingest with wrong organization (should fail)
echo -e "\n3. Ingesting with wrong organization (should fail)..."
grpcurl -plaintext -d '{
  "organization_id": "wrong-org-id",
  "end_device_id": "'$DEVICE_ID'",
  "data": {"temperature": 25.0}
}' localhost:8080 iot.v1.DataIngestionService/IngestDeviceData

# Test 4: Ingest for non-existent device (should fail)
echo -e "\n4. Ingesting for non-existent device (should fail)..."
grpcurl -plaintext -d '{
  "organization_id": "'$ORG_ID'",
  "end_device_id": "non-existent-device",
  "data": {"temperature": 25.0}
}' localhost:8080 iot.v1.DataIngestionService/IngestDeviceData

echo -e "\n=== Test Complete ==="
```

### Success Criteria:

#### Automated Verification:
- [ ] Integration test passes: `go test ./test/integration/... -tags=integration`
- [ ] All existing tests still pass: `go test ./...`
- [ ] No linting errors: `golangci-lint run`
- [ ] Application starts without errors: `mage stack:up && go run ./cmd/ponix-all-in-one`

#### Manual Verification:
- [ ] Create HTTP device via RPC (check database for hardware_type = 2)
- [ ] Ingest telemetry data via IngestDeviceData RPC
- [ ] Ingestion fails for non-existent device with clear error
- [ ] Ingestion fails for organization mismatch with clear error
- [ ] Data appears in ClickHouse: `SELECT * FROM processed_envelopes WHERE end_device_id = '{device_id}'`
- [ ] Data retrievable via QueryEndDeviceData RPC
- [ ] NATS messages published with correct subject: `processed_envelopes.{device_id}`
- [ ] No errors in application logs during ingestion

**Implementation Note**: After completing this phase and all automated verification passes, pause here for manual confirmation from the human that the manual testing was successful before declaring implementation complete.

---

## Testing Strategy

### Unit Tests:
- DataEnvelopeManager device validation logic
- IngestionHandler request validation
- Domain manager HTTP hardware type support
- PostgreSQL store handling of HTTP devices

### Integration Tests:
- Complete flow: device creation → data ingestion → query
- Validation scenarios (non-existent device, wrong organization)
- Concurrent ingestion requests
- LoRaWAN flow still works (regression test)

### Manual Testing Steps:
1. Start full stack: `mage stack:up && go run ./cmd/ponix-all-in-one`
2. Create HTTP device via gRPCurl or Connect client
3. Ingest test data via IngestDeviceData RPC
4. Verify validation works (wrong org, missing device)
5. Check ClickHouse: `SELECT * FROM processed_envelopes WHERE end_device_id = 'test-device'`
6. Query data via RPC: `QueryEndDeviceData`
7. Verify NATS message in stream: `nats stream view processed_envelopes`

## Performance Considerations

- Device validation adds one PostgreSQL query per ingestion (acceptable for MVP)
- Consider caching device-to-organization mapping if ingestion rate becomes high
- ConnectRPC provides automatic request size limits and timeouts
- Reuses existing batching in NATS consumer for efficient ClickHouse writes
- Struct payload allows flexible JSON without schema changes

## Migration Notes

- Database migration is backward compatible (adds constraint, doesn't modify data)
- Existing LoRaWAN devices unaffected
- DataEnvelopeManager signature changes - requires updating call site in main.go
- No data migration required as this adds new functionality

## References

- Original ticket: GitHub Issue #6
- Related issue: GitHub Issue #5 (End Device Type System - deferred)
- ProcessedEnvelope flow: `internal/domain/data_envelope.go:38-51`
- EndDevice CRUD: `internal/domain/end_device.go:43-156`
- ConnectRPC pattern: `internal/connectrpc/end_device_data.go:32-80`
- NATS configuration: `internal/conf/ingestion.go:6-16`
