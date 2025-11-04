package domain

import (
	"context"
	"time"

	iotv1 "buf.build/gen/go/ponix/ponix/protocolbuffers/go/iot/v1"
	"github.com/ponix-dev/ponix/internal/telemetry"
	"github.com/ponix-dev/ponix/internal/telemetry/stacktrace"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// EndDeviceHistogram represents a single time bucket's histogram data
type EndDeviceHistogram struct {
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

// EnvelopeQuerier queries end device data from storage
type EnvelopeQuerier interface {
	QueryEndDeviceData(
		ctx context.Context,
		organizationID string,
		deviceIDs []string,
		startTime, endTime time.Time,
		fieldPath string,
		valueBuckets []float64,
	) ([]EndDeviceHistogram, error)
}

// EndDeviceDataManager orchestrates end device data query operations.
type EndDeviceDataManager struct {
	envelopeStore EnvelopeQuerier
	validator     Validate
}

// NewEndDeviceDataManager creates a new instance of EndDeviceDataManager.
func NewEndDeviceDataManager(envelopeStore EnvelopeQuerier, validator Validate) *EndDeviceDataManager {
	return &EndDeviceDataManager{
		envelopeStore: envelopeStore,
		validator:     validator,
	}
}

// QueryEndDeviceData queries time-series sensor data with histogram aggregation
func (mgr *EndDeviceDataManager) QueryEndDeviceData(
	ctx context.Context,
	req *iotv1.QueryEndDeviceDataRequest,
) (*iotv1.QueryEndDeviceDataResponse, error) {
	ctx, span := telemetry.Tracer().Start(ctx, "EndDeviceDataManager.QueryEndDeviceData")
	defer span.End()

	// Validate request
	err := mgr.validator(req)
	if err != nil {
		return nil, stacktrace.NewStackTraceError(err)
	}

	// Query data from store
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
	timeBucketInterval := CalculateTimeBucketInterval(
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
	response := ConvertToProtoResponse(results, deviceCount, timeBucketInterval)

	return response, nil
}

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

// ConvertToProtoResponse converts internal query results to protobuf response
func ConvertToProtoResponse(
	results []EndDeviceHistogram,
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
