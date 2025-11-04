# Organization-Scoped Sensor Data Query Implementation Plan

**Date**: 2025-11-04
**Feature**: Query sensor data from ClickHouse with Prometheus-style histograms for time-series visualization

---

## Overview

Implement a Connect-RPC endpoint that allows web/mobile UIs to query time-series sensor data stored in ClickHouse. The endpoint will:
- Filter data by organization (with authorization enforcement)
- Filter by specific devices or all devices in an organization
- Accept time window parameters from the client
- Return Prometheus-style histogram data with auto-calculated time buckets
- Support configurable value buckets for distribution analysis

---

## Current State Analysis

### Existing Infrastructure
- **ClickHouse Table**: `processed_envelopes` with columns:
  - `end_device_id` (String)
  - `occurred_at` (DateTime64)
  - `processed_at` (DateTime64)
  - `data` (JSON) - flexible sensor data
  - **Missing**: `organization_id` for efficient filtering

- **Data Flow**: IoT Data → Webhook → Domain Layer → NATS JetStream → Consumer → ClickHouse

- **Authorization**: Casbin-based with organization-level permissions (`CanReadEndDevice`)

- **Protobuf Structure**:
  - `ProcessedEnvelope` (envelope/v1) - current ingestion format
  - `EndDeviceDataRequest/Response` (iot/v1) - existing but empty stub at [end_device.go:164-205](internal/connectrpc/end_device.go#L164)

### Key Constraints
- Protobuf definitions hosted on buf.build (can be modified in `ponix-protobuf` repo)
- PostgreSQL has `organization_id` on `end_devices` table
- ClickHouse data currently has no direct organization reference
- Multi-tenant architecture requires organization-scoped queries

---

## Desired End State

Users authenticated to an organization can query their sensor data through a Connect-RPC endpoint that returns:
- Time-bucketed histogram data (Prometheus-style cumulative buckets)
- Filtered by organization (enforced via Casbin authorization)
- Optional device-specific filtering
- Auto-calculated time granularity based on query window
- Configurable value distribution buckets

### Success Verification
```bash
# After implementation, verify:
# 1. ClickHouse migration applied
make clickhouse-migrate

# 2. Protobuf updated and generated
cd ../ponix-protobuf && buf push && cd ../ponix
go mod tidy

# 3. All tests pass
go test ./...

# 4. Service starts without errors
go run ./cmd/ponix-all-in-one

# 5. Can query sensor data via RPC with proper authorization
# (Manual testing via UI or grpcurl)
```

---

## What We're NOT Doing

- Real-time streaming of sensor data (this is query-only)
- Complex aggregations beyond histograms (no percentiles, stddev, etc. for now)
- Query caching or optimization (can be added later)
- GraphQL or REST endpoints (Connect-RPC only)
- Historical data backfill of `organization_id` (new data only)

---

## Implementation Approach

We'll use a phased approach to:
1. Add organization context to ClickHouse schema and ingestion pipeline
2. Define protobuf messages for histogram queries
3. Implement ClickHouse query logic with histogram bucketing
4. Wire up RPC handler with authorization
5. Test end-to-end with various time ranges

---

## Phase 1: ClickHouse Schema Migration

### Overview
Add `organization_id` to the ClickHouse `processed_envelopes` table for efficient organization-scoped queries.

### Changes Required

#### 1. Create ClickHouse Migration
**File**: `internal/clickhouse/goose/<timestamp>_add_organization_id.sql`

```sql
-- +goose Up
ALTER TABLE processed_envelopes
ADD COLUMN organization_id String NOT NULL DEFAULT '';

-- Update the table engine to include organization_id in the ORDER BY
-- Note: In ClickHouse, we need to recreate the table to change ORDER BY
CREATE TABLE processed_envelopes_new (
  `organization_id` String NOT NULL,
  `end_device_id` String NOT NULL,
  `occurred_at` DateTime64(3, 'UTC') NOT NULL,
  `processed_at` DateTime64(3, 'UTC') NOT NULL,
  `data` JSON NOT NULL,
  PRIMARY KEY (organization_id, occurred_at, end_device_id)
) ENGINE = MergeTree()
ORDER BY (organization_id, occurred_at, end_device_id)
PARTITION BY toYYYYMM(occurred_at);

-- Copy data (if any exists)
INSERT INTO processed_envelopes_new
SELECT '' as organization_id, end_device_id, occurred_at, processed_at, data
FROM processed_envelopes;

-- Swap tables
RENAME TABLE processed_envelopes TO processed_envelopes_old;
RENAME TABLE processed_envelopes_new TO processed_envelopes;

-- Drop old table after verification
DROP TABLE processed_envelopes_old;

-- +goose Down
-- Reverse migration: recreate original table structure
CREATE TABLE processed_envelopes_old (
  `end_device_id` String NOT NULL,
  `occurred_at` DateTime64(3, 'UTC') NOT NULL,
  `processed_at` DateTime64(3, 'UTC') NOT NULL,
  `data` JSON NOT NULL,
  PRIMARY KEY (occurred_at, end_device_id)
) ENGINE = MergeTree()
ORDER BY (occurred_at, end_device_id)
PARTITION BY toYYYYMM(occurred_at);

INSERT INTO processed_envelopes_old
SELECT end_device_id, occurred_at, processed_at, data
FROM processed_envelopes;

RENAME TABLE processed_envelopes TO processed_envelopes_new;
RENAME TABLE processed_envelopes_old TO processed_envelopes;
DROP TABLE processed_envelopes_new;
```

#### 2. Update ClickHouse Schema Definition
**File**: `schema/clickhouse/schema.sql`

```sql
CREATE TABLE `processed_envelopes` (
  `organization_id` String NOT NULL,
  `end_device_id` String NOT NULL,
  `occurred_at` DateTime64(3, 'UTC') NOT NULL,
  `processed_at` DateTime64(3, 'UTC') NOT NULL,
  `data` JSON NOT NULL,
  PRIMARY KEY (organization_id, occurred_at, end_device_id)
) ENGINE = MergeTree()
ORDER BY (organization_id, occurred_at, end_device_id)
PARTITION BY toYYYYMM(occurred_at);
```

### Success Criteria

#### Automated Verification:
- [ ] Migration runs successfully: `mage clickhouse:migrate`
- [ ] Schema file updated and committed
- [ ] Service starts without ClickHouse errors

#### Manual Verification:
- [ ] Verify table structure in ClickHouse: `DESCRIBE processed_envelopes`
- [ ] Confirm ORDER BY includes organization_id

---

## Phase 2: Update Data Ingestion Pipeline

### Overview
Modify the ingestion pipeline to include `organization_id` when storing envelopes to ClickHouse.

### Changes Required

#### 1. Update ProcessedEnvelope Protobuf
**File**: `/Users/srall/development/personal/ponix-protobuf/envelope/v1/envelope_processed.proto`

```protobuf
syntax = "proto3";

package envelope.v1;

import "google/protobuf/struct.proto";
import "google/protobuf/timestamp.proto";

message ProcessedEnvelope {
  string organization_id = 1;
  string end_device_id = 2;
  google.protobuf.Timestamp occurred_at = 3;
  google.protobuf.Struct data = 4;
  google.protobuf.Timestamp processed_at = 5;
}
```

**Note**: Field numbers changed to accommodate new `organization_id` field.

#### 2. Push Updated Protobuf
**Commands**:
```bash
cd /Users/srall/development/personal/ponix-protobuf
buf push
```

#### 3. Update Go Module Dependencies
**File**: `go.mod` (update buf.build dependencies)
```bash
cd /Users/srall/development/personal/ponix
go get buf.build/gen/go/ponix/ponix/protocolbuffers/go@latest
go mod tidy
```

#### 4. Update Domain Layer to Include Organization ID
**File**: `internal/domain/data_envelope.go`

Modify `IngestDataEnvelope` to accept and pass `organizationID`:

```go
func (mgr *DataEnvelopeManager) IngestDataEnvelope(ctx context.Context, envelope *envelopev1.DataEnvelope, organizationID string) error {
	ctx, span := telemetry.Tracer().Start(ctx, "IngestDataEnvelope")
	defer span.End()

	processedEnvelope := envelopev1.ProcessedEnvelope_builder{
		OrganizationId: organizationID,
		EndDeviceId:    envelope.GetEndDeviceId(),
		OccurredAt:     envelope.GetOccurredAt(),
		Data:           envelope.GetData(),
		ProcessedAt:    timestamppb.New(time.Now().UTC()),
	}.Build()

	return mgr.producer.ProduceProcessedEnvelope(ctx, processedEnvelope)
}
```

#### 5. Update ClickHouse EnvelopeStore
**File**: `internal/clickhouse/envelope.go`

Modify `StoreProcessedEnvelopes` to include `organization_id`:

```go
err = batch.Append(
	envelope.GetOrganizationId(),
	envelope.GetEndDeviceId(),
	envelope.GetOccurredAt().AsTime(),
	envelope.GetProcessedAt().AsTime(),
	string(dataJson),
)
```

#### 6. Update Webhook/Ingestion Callers
Find all callers of `IngestDataEnvelope` and ensure they pass `organizationID`. This will require:
- Looking up the device's organization from PostgreSQL
- Passing it through the ingestion call

**Research needed**: Find webhook endpoint and update it.

### Success Criteria

#### Automated Verification:
- [x] Protobuf pushed successfully: `buf push` completes
- [x] Go modules updated: `go mod tidy` succeeds
- [x] Code compiles: `go build ./...`
- [x] Tests pass: `go test ./internal/domain/...`
- [x] Tests pass: `go test ./internal/clickhouse/...`

#### Manual Verification:
- [ ] Send test data through webhook, verify organization_id stored in ClickHouse
- [ ] Check NATS messages include organization_id
- [ ] Verify no data loss during transition

**Implementation Note**: After completing this phase and all automated verification passes, pause here for manual confirmation that data is flowing correctly before proceeding to the next phase.

---

## Phase 3: Protobuf Definitions for Query API

### Overview
Define Connect-RPC message types for querying sensor data with Prometheus-style histograms.

### Changes Required

#### 1. Create Sensor Data Query Messages
**File**: `/Users/srall/development/personal/ponix-protobuf/iot/v1/sensor_data.proto`

```protobuf
syntax = "proto3";

package iot.v1;

import "buf/validate/validate.proto";
import "google/protobuf/timestamp.proto";
import "google/protobuf/duration.proto";

// Request to query sensor data with histogram aggregation
message QueryEndDeviceDataRequest {
  // Organization ID (required, will be validated against user's access)
  string organization_id = 1 [(buf.validate.field).required = true];

  // Optional: Filter by specific device IDs. If empty, query all devices in org.
  repeated string end_device_ids = 2;

  // Time range for query (required)
  google.protobuf.Timestamp start_time = 3 [(buf.validate.field).required = true];
  google.protobuf.Timestamp end_time = 4 [(buf.validate.field).required = true];

  // JSON path to the field to histogram (e.g., "flow_rate", "temperature")
  string field_path = 5 [(buf.validate.field).required = true];

  // Value bucket boundaries for histogram (Prometheus-style "le" buckets)
  // Example: [0, 10, 50, 100, 500] creates buckets: <=0, <=10, <=50, <=100, <=500, +Inf
  repeated double value_buckets = 6 [(buf.validate.field).repeated.min_items = 1];
}

// Response containing time-series histogram data
message QueryEndDeviceDataResponse {
  // Array of time-bucketed histograms
  repeated TimeSeriesHistogram time_series = 1;

  // Metadata about the query
  QueryMetadata metadata = 2;
}

// A single time bucket with histogram data
message TimeSeriesHistogram {
  // Start timestamp of this time bucket
  google.protobuf.Timestamp bucket_start = 1;

  // End timestamp of this time bucket
  google.protobuf.Timestamp bucket_end = 2;

  // Prometheus-style histogram buckets (cumulative counts)
  repeated HistogramBucket buckets = 3;

  // Total count of observations in this time bucket
  uint64 count = 4;

  // Sum of all observed values in this time bucket
  double sum = 5;
}

// Prometheus-style cumulative histogram bucket
message HistogramBucket {
  // Upper bound (le = "less than or equal to")
  // Use special value +Inf (represented as max float64) for the last bucket
  double le = 1;

  // Cumulative count of observations <= le
  uint64 cumulative_count = 2;
}

// Query metadata
message QueryMetadata {
  // Total number of data points across all time buckets
  uint64 total_count = 1;

  // Time bucket interval that was auto-calculated
  google.protobuf.Duration time_bucket_interval = 2;

  // Number of time buckets returned
  uint32 time_bucket_count = 3;

  // Number of devices included in the query
  uint32 device_count = 4;
}
```

#### 2. Update EndDeviceService
**File**: `/Users/srall/development/personal/ponix-protobuf/iot/v1/end_device.proto`

Replace the existing `EndDeviceData` RPC with new query endpoint:

```protobuf
import "iot/v1/sensor_data.proto";

service EndDeviceService {
  rpc CreateEndDevice(CreateEndDeviceRequest) returns (CreateEndDeviceResponse);
  rpc EndDevice(EndDeviceRequest) returns (EndDeviceResponse);
  rpc OrganizationEndDevices(OrganizationEndDevicesRequest) returns (OrganizationEndDevicesResponse);

  // New: Query sensor data with histogram aggregation
  rpc QueryEndDeviceData(QueryEndDeviceDataRequest) returns (QueryEndDeviceDataResponse);
}
```

Remove old `EndDeviceDataRequest` and `EndDeviceDataResponse` messages (lines 73-81).

#### 3. Push Updated Protobufs
**Commands**:
```bash
cd /Users/srall/development/personal/ponix-protobuf
buf push
cd /Users/srall/development/personal/ponix
go get buf.build/gen/go/ponix/ponix/protocolbuffers/go@latest
go get buf.build/gen/go/ponix/ponix/connectrpc/go@latest
go mod tidy
```

### Success Criteria

#### Automated Verification:
- [x] Protobuf lint passes: `buf lint`
- [x] Protobuf push succeeds: `buf push`
- [x] Go code generates: `go mod tidy`
- [x] Code compiles: `go build ./...`

#### Manual Verification:
- [ ] Review generated Go types for correctness
- [ ] Verify Connect-RPC service includes new `QueryEndDeviceData` method

**Implementation Note**: After completing this phase, verify the generated code looks correct before implementing the query logic.

---

## Phase 4: ClickHouse Query Store Implementation

### Overview
Implement the ClickHouse query logic to retrieve and aggregate sensor data into Prometheus-style histograms.

### Changes Required

#### 1. Add Time Bucket Calculation Helper
**File**: `internal/clickhouse/time_bucket.go` (new file)

```go
package clickhouse

import (
	"time"

	"google.golang.org/protobuf/types/known/durationpb"
)

// CalculateTimeBucketInterval determines the appropriate time bucket size
// based on the query time range. Uses simple, round values for bucketing.
func CalculateTimeBucketInterval(startTime, endTime time.Time) time.Duration {
	duration := endTime.Sub(startTime)

	switch {
	case duration < time.Hour:
		return 5 * time.Minute // < 1 hour: 5-minute buckets
	case duration < 6*time.Hour:
		return 15 * time.Minute // < 6 hours: 15-minute buckets
	case duration < 24*time.Hour:
		return time.Hour // < 1 day: 1-hour buckets
	case duration < 7*24*time.Hour:
		return 6 * time.Hour // < 1 week: 6-hour buckets
	default:
		return 24 * time.Hour // >= 1 week: 1-day buckets
	}
}

// ToProtoDuration converts time.Duration to google.protobuf.Duration
func ToProtoDuration(d time.Duration) *durationpb.Duration {
	return durationpb.New(d)
}
```

#### 2. Add Sensor Data Query Method
**File**: `internal/clickhouse/envelope.go`

Add new method to `EnvelopeStore`:

```go
import (
	iotv1 "buf.build/gen/go/ponix/ponix/protocolbuffers/go/iot/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
	"math"
)

// QueryEndDeviceDataResult represents a single time bucket's histogram data
type QueryEndDeviceDataResult struct {
	BucketStart time.Time
	BucketEnd   time.Time
	Buckets     []HistogramBucketResult
	Count       uint64
	Sum         float64
}

// HistogramBucketResult represents a single histogram bucket
type HistogramBucketResult struct {
	LE              float64
	CumulativeCount uint64
}

// QueryEndDeviceData retrieves sensor data for an organization with histogram aggregation.
// It returns time-bucketed histograms based on the query parameters.
func (es *EnvelopeStore) QueryEndDeviceData(
	ctx context.Context,
	organizationID string,
	deviceIDs []string,
	startTime, endTime time.Time,
	fieldPath string,
	valueBuckets []float64,
) ([]QueryEndDeviceDataResult, error) {
	ctx, span := telemetry.Tracer().Start(ctx, "QueryEndDeviceData")
	defer span.End()

	// Calculate appropriate time bucket interval
	timeBucketInterval := CalculateTimeBucketInterval(startTime, endTime)

	// Build the ClickHouse query
	query := es.buildHistogramQuery(organizationID, deviceIDs, startTime, endTime, fieldPath, valueBuckets, timeBucketInterval)

	rows, err := es.db.Query(ctx, query,
		organizationID,
		startTime,
		endTime,
	)
	if err != nil {
		return nil, stacktrace.NewStackTraceError(err)
	}
	defer rows.Close()

	// Parse results
	results := make([]QueryEndDeviceDataResult, 0)
	for rows.Next() {
		var bucketStart time.Time
		var bucketEnd time.Time
		var count uint64
		var sum float64
		var bucketCounts []uint64 // Cumulative counts for each value bucket

		err := rows.Scan(&bucketStart, &bucketEnd, &count, &sum, &bucketCounts)
		if err != nil {
			return nil, stacktrace.NewStackTraceError(err)
		}

		// Build histogram buckets
		buckets := make([]HistogramBucketResult, len(valueBuckets))
		for i, le := range valueBuckets {
			buckets[i] = HistogramBucketResult{
				LE:              le,
				CumulativeCount: bucketCounts[i],
			}
		}

		// Add +Inf bucket
		buckets = append(buckets, HistogramBucketResult{
			LE:              math.Inf(1),
			CumulativeCount: count,
		})

		results = append(results, QueryEndDeviceDataResult{
			BucketStart: bucketStart,
			BucketEnd:   bucketEnd,
			Buckets:     buckets,
			Count:       count,
			Sum:         sum,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, stacktrace.NewStackTraceError(err)
	}

	return results, nil
}

// buildHistogramQuery constructs the ClickHouse SQL query for histogram aggregation
func (es *EnvelopeStore) buildHistogramQuery(
	organizationID string,
	deviceIDs []string,
	startTime, endTime time.Time,
	fieldPath string,
	valueBuckets []float64,
	timeBucketInterval time.Duration,
) string {
	// Convert time bucket interval to seconds for ClickHouse
	intervalSeconds := int(timeBucketInterval.Seconds())

	// Build device filter clause
	deviceFilter := ""
	if len(deviceIDs) > 0 {
		deviceFilter = fmt.Sprintf("AND end_device_id IN ('%s')", strings.Join(deviceIDs, "','"))
	}

	// Build histogram bucket expressions (cumulative counts)
	bucketExpressions := make([]string, len(valueBuckets))
	for i, le := range valueBuckets {
		bucketExpressions[i] = fmt.Sprintf("countIf(value <= %f)", le)
	}

	query := fmt.Sprintf(`
		SELECT
			toStartOfInterval(occurred_at, INTERVAL %d SECOND) as bucket_start,
			toStartOfInterval(occurred_at, INTERVAL %d SECOND) + INTERVAL %d SECOND as bucket_end,
			count(*) as count,
			sum(value) as sum,
			[%s] as bucket_counts
		FROM (
			SELECT
				occurred_at,
				JSONExtractFloat(data, '%s') as value
			FROM %s
			WHERE organization_id = ?
			  AND occurred_at >= ?
			  AND occurred_at < ?
			  %s
			  AND JSONHas(data, '%s')
		)
		GROUP BY bucket_start, bucket_end
		ORDER BY bucket_start
	`,
		intervalSeconds,
		intervalSeconds,
		intervalSeconds,
		strings.Join(bucketExpressions, ", "),
		fieldPath,
		es.table,
		deviceFilter,
		fieldPath,
	)

	return query
}

// ConvertToProtoResponse converts internal query results to protobuf response
func ConvertToProtoResponse(
	results []QueryEndDeviceDataResult,
	deviceCount int,
	timeBucketInterval time.Duration,
) *iotv1.QueryEndDeviceDataResponse {
	timeSeries := make([]*iotv1.TimeSeriesHistogram, len(results))

	totalCount := uint64(0)
	for i, result := range results {
		buckets := make([]*iotv1.HistogramBucket, len(result.Buckets))
		for j, bucket := range result.Buckets {
			buckets[j] = &iotv1.HistogramBucket{
				Le:              bucket.LE,
				CumulativeCount: bucket.CumulativeCount,
			}
		}

		timeSeries[i] = &iotv1.TimeSeriesHistogram{
			BucketStart: timestamppb.New(result.BucketStart),
			BucketEnd:   timestamppb.New(result.BucketEnd),
			Buckets:     buckets,
			Count:       result.Count,
			Sum:         result.Sum,
		}

		totalCount += result.Count
	}

	metadata := &iotv1.QueryMetadata{
		TotalCount:         totalCount,
		TimeBucketInterval: ToProtoDuration(timeBucketInterval),
		TimeBucketCount:    uint32(len(results)),
		DeviceCount:        uint32(deviceCount),
	}

	return &iotv1.QueryEndDeviceDataResponse{
		TimeSeries: timeSeries,
		Metadata:   metadata,
	}
}
```

### Success Criteria

#### Automated Verification:
- [x] Code compiles: `go build ./internal/clickhouse/...`
- [x] Unit tests pass: `go test ./internal/clickhouse/...`
- [x] ClickHouse query syntax is valid

#### Manual Verification:
- [ ] Test with sample data in ClickHouse
- [ ] Verify histogram buckets are cumulative
- [ ] Verify time bucketing works correctly for different ranges
- [ ] Check query performance with large datasets

**Implementation Note**: Create unit tests for time bucket calculation and histogram aggregation logic before proceeding.

---

## Phase 5: Domain Layer & RPC Handler

### Overview
Wire up the RPC handler with authorization, domain logic, and ClickHouse store integration.

### Changes Required

#### 1. Update EndDeviceAuthorizer Interface
**File**: `internal/connectrpc/end_device.go`

The existing `CanReadEndDevice` method should already cover reading sensor data, so no changes needed to the interface.

#### 2. Implement QueryEndDeviceData RPC Handler
**File**: `internal/connectrpc/end_device.go`

Replace the stub `EndDeviceData` method with `QueryEndDeviceData`:

```go
// QueryEndDeviceData handles RPC requests to query time-series sensor data.
// Returns Prometheus-style histogram data for visualization.
// Requires super admin privileges or device read permission in the organization.
func (handler *EndDeviceHandler) QueryEndDeviceData(
	ctx context.Context,
	req *connect.Request[iotv1.QueryEndDeviceDataRequest],
) (*connect.Response[iotv1.QueryEndDeviceDataResponse], error) {
	ctx, span := telemetry.Tracer().Start(ctx, "QueryEndDeviceData")
	defer span.End()

	// Extract user from context
	userId, ok := domain.GetUserFromContext(ctx)
	if !ok {
		return nil, connect.NewError(connect.CodeUnauthenticated, fmt.Errorf("user not authenticated"))
	}

	// Extract organization from request
	organizationID := req.Msg.GetOrganizationId()
	if organizationID == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("organization ID is required"))
	}

	// Authorization check
	allowed := false
	if domain.IsSuperAdminFromContext(ctx) {
		allowed = true
	} else {
		can, err := handler.authorizer.CanReadEndDevice(ctx, userId, organizationID)
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("authorization check failed: %w", err))
		}
		allowed = can
	}

	if !allowed {
		return nil, connect.NewError(
			connect.CodePermissionDenied,
			fmt.Errorf("user %s not authorized to read sensor data in organization %s", userId, organizationID),
		)
	}

	// Validate time range
	startTime := req.Msg.GetStartTime().AsTime()
	endTime := req.Msg.GetEndTime().AsTime()
	if endTime.Before(startTime) {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("end_time must be after start_time"))
	}

	// Query sensor data
	response, err := handler.endDeviceManager.QueryEndDeviceData(ctx, req.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to query sensor data: %w", err))
	}

	return connect.NewResponse(response), nil
}
```

Remove the old `EndDeviceData` method (lines 164-205).

#### 3. Update EndDeviceManager Interface
**File**: `internal/connectrpc/end_device.go`

Add new method to interface:

```go
type EndDeviceManager interface {
	CreateEndDevice(ctx context.Context, createReq *iotv1.CreateEndDeviceRequest, organization string) (*iotv1.EndDevice, error)
	QueryEndDeviceData(ctx context.Context, req *iotv1.QueryEndDeviceDataRequest) (*iotv1.QueryEndDeviceDataResponse, error)
}
```

#### 4. Implement Domain Manager Method
**File**: `internal/domain/end_device.go`

Add `QueryEndDeviceData` method to `EndDeviceManager`:

```go
import (
	"buf.build/gen/go/ponix/ponix/protocolbuffers/go/iot/v1"
	"github.com/ponix-dev/ponix/internal/clickhouse"
)

type EndDeviceManager struct {
	store          EndDeviceStore
	ttnClient      TTNClient
	applicationId  string
	idGenerator    func() string
	validator      Validator
	envelopeStore  EnvelopeQuerier // New dependency
}

// EnvelopeQuerier queries sensor data from ClickHouse
type EnvelopeQuerier interface {
	QueryEndDeviceData(
		ctx context.Context,
		organizationID string,
		deviceIDs []string,
		startTime, endTime time.Time,
		fieldPath string,
		valueBuckets []float64,
	) ([]clickhouse.QueryEndDeviceDataResult, error)
}

func NewEndDeviceManager(
	store EndDeviceStore,
	ttnClient TTNClient,
	applicationId string,
	idGenerator func() string,
	validator Validator,
	envelopeStore EnvelopeQuerier,
) *EndDeviceManager {
	return &EndDeviceManager{
		store:         store,
		ttnClient:     ttnClient,
		applicationId: applicationId,
		idGenerator:   idGenerator,
		validator:     validator,
		envelopeStore: envelopeStore,
	}
}

// QueryEndDeviceData queries time-series sensor data with histogram aggregation
func (mgr *EndDeviceManager) QueryEndDeviceData(
	ctx context.Context,
	req *iotv1.QueryEndDeviceDataRequest,
) (*iotv1.QueryEndDeviceDataResponse, error) {
	ctx, span := telemetry.Tracer().Start(ctx, "EndDeviceManager.QueryEndDeviceData")
	defer span.End()

	// Validate request
	err := mgr.validator(req)
	if err != nil {
		return nil, stacktrace.NewStackTraceError(err)
	}

	// Query ClickHouse
	results, err := mgr.envelopeStore.QueryEndDeviceData(
		ctx,
		req.GetOrganizationId(),
		req.GetEndDeviceIds(),
		req.GetStartTime().AsTime(),
		req.GetEndTime().AsTime(),
		req.GetFieldPath(),
		req.GetValueBuckets(),
	)
	if err != nil {
		return nil, stacktrace.NewStackTraceError(err)
	}

	// Calculate time bucket interval
	timeBucketInterval := clickhouse.CalculateTimeBucketInterval(
		req.GetStartTime().AsTime(),
		req.GetEndTime().AsTime(),
	)

	// Determine device count
	deviceCount := len(req.GetEndDeviceIds())
	if deviceCount == 0 {
		// Query all devices in organization - would need to fetch count from PostgreSQL
		// For now, we can leave this as 0 or implement a separate query
		deviceCount = 0
	}

	// Convert results to proto response
	response := clickhouse.ConvertToProtoResponse(results, deviceCount, timeBucketInterval)

	return response, nil
}
```

#### 5. Update Dependency Injection
**File**: `cmd/ponix-all-in-one/main.go`

Update the EndDeviceManager initialization to include envelopeStore:

```go
edMgr := domain.NewEndDeviceManager(
	edStore,
	ttnClient,
	cfg.ApplicationId,
	xid.StringId,
	protobuf.Validate,
	envelopeStore, // Add this parameter
)
```

### Success Criteria

#### Automated Verification:
- [x] Code compiles: `go build ./...`
- [x] Unit tests pass: `go test ./internal/domain/...`
- [x] Unit tests pass: `go test ./internal/connectrpc/...`
- [ ] Service starts: `go run ./cmd/ponix-all-in-one`

#### Manual Verification:
- [ ] Can call QueryEndDeviceData RPC with valid authorization
- [ ] Unauthorized users receive permission denied error
- [ ] Invalid time ranges return appropriate errors
- [ ] Response includes correct histogram data

**Implementation Note**: After completing this phase, test the full RPC flow with a client or grpcurl before proceeding to comprehensive testing.

---

## Phase 6: Testing & Validation

### Overview
Create comprehensive tests and validate the implementation end-to-end.

### Changes Required

#### 1. Unit Tests for Time Bucket Calculation
**File**: `internal/clickhouse/time_bucket_test.go` (new file)

```go
package clickhouse

import (
	"testing"
	"time"
)

func TestCalculateTimeBucketInterval(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		expected time.Duration
	}{
		{"30 minutes", 30 * time.Minute, 5 * time.Minute},
		{"1 hour", time.Hour, 5 * time.Minute},
		{"3 hours", 3 * time.Hour, 15 * time.Minute},
		{"12 hours", 12 * time.Hour, time.Hour},
		{"1 day", 24 * time.Hour, time.Hour},
		{"3 days", 3 * 24 * time.Hour, 6 * time.Hour},
		{"1 week", 7 * 24 * time.Hour, 6 * time.Hour},
		{"1 month", 30 * 24 * time.Hour, 24 * time.Hour},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start := time.Now()
			end := start.Add(tt.duration)
			result := CalculateTimeBucketInterval(start, end)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}
```

#### 2. Integration Tests for ClickHouse Queries
**File**: `internal/clickhouse/envelope_test.go`

Add test for `QueryEndDeviceData`:

```go
func TestEnvelopeStore_QueryEndDeviceData(t *testing.T) {
	// Skip if no ClickHouse connection available
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	// Setup: Create test data in ClickHouse
	// Test: Query with various parameters
	// Assert: Verify histogram structure and data
}
```

#### 3. RPC Handler Tests
**File**: `internal/connectrpc/end_device_test.go` (new file)

```go
package connectrpc

import (
	"context"
	"testing"
	"time"

	iotv1 "buf.build/gen/go/ponix/ponix/protocolbuffers/go/iot/v1"
	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestEndDeviceHandler_QueryEndDeviceData_Authorization(t *testing.T) {
	// Test cases:
	// - Unauthenticated user returns error
	// - Unauthorized user returns permission denied
	// - Super admin can access
	// - Org member with read permission can access
}

func TestEndDeviceHandler_QueryEndDeviceData_Validation(t *testing.T) {
	// Test cases:
	// - Missing organization_id returns error
	// - Invalid time range returns error
	// - Missing field_path returns error
	// - Empty value_buckets returns error
}
```

#### 4. Manual Testing Script
**File**: `scripts/test_query_sensor_data.sh` (new file)

```bash
#!/bin/bash
# Manual test script for QueryEndDeviceData RPC

# Variables
BASE_URL="http://localhost:8080"
ORG_ID="test-org-123"
DEVICE_ID="test-device-456"

# Test 1: Query last hour of data
echo "Test 1: Query last hour"
grpcurl -plaintext -d '{
  "organization_id": "'$ORG_ID'",
  "end_device_ids": ["'$DEVICE_ID'"],
  "start_time": "'$(date -u -v-1H +%Y-%m-%dT%H:%M:%SZ)'",
  "end_time": "'$(date -u +%Y-%m-%dT%H:%M:%SZ)'",
  "field_path": "flow_rate",
  "value_buckets": [0, 10, 50, 100, 500]
}' localhost:8080 iot.v1.EndDeviceService/QueryEndDeviceData

# Test 2: Query all devices in organization
echo "Test 2: Query all devices"
grpcurl -plaintext -d '{
  "organization_id": "'$ORG_ID'",
  "start_time": "'$(date -u -v-1d +%Y-%m-%dT%H:%M:%SZ)'",
  "end_time": "'$(date -u +%Y-%m-%dT%H:%M:%SZ)'",
  "field_path": "temperature",
  "value_buckets": [0, 20, 40, 60, 80, 100]
}' localhost:8080 iot.v1.EndDeviceService/QueryEndDeviceData
```

#### 5. Documentation Update
**File**: `README.md` or `docs/sensor-data-query.md`

Document the new API endpoint with:
- Request/response examples
- Histogram bucket configuration guide
- Time bucketing behavior
- Authorization requirements
- Common use cases

### Success Criteria

#### Automated Verification:
- [ ] All unit tests pass: `go test ./...`
- [ ] Integration tests pass: `go test -short=false ./internal/clickhouse/...`
- [ ] Linting passes: `golangci-lint run`
- [ ] Code formatting correct: `go fmt ./...`

#### Manual Verification:
- [ ] Query data for last hour, verify 5-minute buckets
- [ ] Query data for last week, verify 6-hour buckets
- [ ] Query data for last month, verify 1-day buckets
- [ ] Verify histogram buckets are cumulative (each bucket count >= previous)
- [ ] Verify authorization blocks unauthorized users
- [ ] Verify organization scoping works correctly (can't access other org's data)
- [ ] Test with missing data (some time buckets have no observations)
- [ ] Test with various value bucket configurations
- [ ] Verify performance with large datasets (1M+ data points)

**Implementation Note**: Complete all automated tests before manual testing. Document any edge cases or limitations discovered during testing.

---

## Performance Considerations

1. **ClickHouse Indexing**: The `ORDER BY (organization_id, occurred_at, end_device_id)` ensures efficient filtering and time-range queries.

2. **Query Optimization**:
   - Partition pruning by month automatically limits data scanned
   - Organization filtering uses primary key
   - Time range filtering uses primary key

3. **Response Size**: Large time ranges may return many buckets. Consider:
   - Adding a max bucket count limit
   - Pagination for very large queries
   - Client-side downsampling

4. **Concurrent Queries**: ClickHouse handles concurrent analytical queries well, but consider rate limiting per organization.

---

## Migration Notes

### Data Migration
- **No backfill required**: Old data without `organization_id` will have empty string
- New data will include `organization_id` from ingestion forward
- Consider cleanup job to populate `organization_id` for historical data if needed

### Rollback Strategy
If issues arise:
1. Revert protobuf changes and re-push old version
2. Run ClickHouse down migration to remove `organization_id` column
3. Revert code changes
4. Redeploy previous version

---

## References

- Prometheus Histogram Design: https://prometheus.io/docs/concepts/metric_types/#histogram
- ClickHouse Aggregation Functions: https://clickhouse.com/docs/en/sql-reference/aggregate-functions
- Connect-RPC Documentation: https://connectrpc.com/docs/
- Existing code patterns:
  - Authorization: [end_device.go:62-77](internal/connectrpc/end_device.go#L62-L77)
  - ClickHouse store: [envelope.go:25-58](internal/clickhouse/envelope.go#L25-L58)
  - Domain managers: [data_envelope.go:22-34](internal/domain/data_envelope.go#L22-L34)
