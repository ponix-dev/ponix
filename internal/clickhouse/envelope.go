package clickhouse

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"

	envelopev1 "buf.build/gen/go/ponix/ponix/protocolbuffers/go/envelope/v1"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"

	"github.com/ponix-dev/ponix/internal/domain"
	"github.com/ponix-dev/ponix/internal/telemetry"
	"github.com/ponix-dev/ponix/internal/telemetry/stacktrace"
)

type EnvelopeStore struct {
	db    driver.Conn
	table string
}

func NewEnvelopeStore(db driver.Conn, table string) *EnvelopeStore {
	return &EnvelopeStore{
		db:    db,
		table: table,
	}
}

func (es *EnvelopeStore) StoreProcessedEnvelopes(ctx context.Context, envelopes ...*envelopev1.ProcessedEnvelope) error {
	ctx, span := telemetry.Tracer().Start(ctx, "StoreProcessedEnvelopes")
	defer span.End()

	batch, err := es.db.PrepareBatch(ctx, fmt.Sprintf("INSERT INTO %s", es.table))
	if err != nil {
		return stacktrace.NewStackTraceError(err)
	}
	defer batch.Close()

	for _, envelope := range envelopes {
		dataJson, err := envelope.GetData().MarshalJSON()
		if err != nil {
			return stacktrace.NewStackTraceError(err)
		}

		err = batch.Append(
			envelope.GetOrganizationId(),
			envelope.GetEndDeviceId(),
			envelope.GetOccurredAt().AsTime(),
			envelope.GetProcessedAt().AsTime(),
			string(dataJson),
		)
		if err != nil {
			return stacktrace.NewStackTraceError(err)
		}
	}

	err = batch.Send()
	if err != nil {
		return stacktrace.NewStackTraceError(err)
	}

	return nil
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
) ([]domain.EndDeviceHistogram, error) {
	ctx, span := telemetry.Tracer().Start(ctx, "QueryEndDeviceData")
	defer span.End()

	// Calculate appropriate time bucket interval
	timeBucketInterval := CalculateTimeBucketInterval(startTime, endTime)

	// Build the ClickHouse query
	query := es.buildHistogramQuery(fieldPath, valueBuckets, timeBucketInterval, len(deviceIDs) > 0)

	// Build query args
	args := []any{organizationID, startTime, endTime}
	if len(deviceIDs) > 0 {
		args = append(args, deviceIDs)
	}

	rows, err := es.db.Query(ctx, query, args...)
	if err != nil {
		return nil, stacktrace.NewStackTraceError(err)
	}
	defer rows.Close()

	// Parse results
	results := make([]domain.EndDeviceHistogram, 0)
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
		buckets := make([]domain.HistogramBucketResult, len(valueBuckets))
		for i, le := range valueBuckets {
			buckets[i] = domain.HistogramBucketResult{
				LE:              le,
				CumulativeCount: bucketCounts[i],
			}
		}

		// Add +Inf bucket
		buckets = append(buckets, domain.HistogramBucketResult{
			LE:              math.Inf(1),
			CumulativeCount: count,
		})

		results = append(results, domain.EndDeviceHistogram{
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
	fieldPath string,
	valueBuckets []float64,
	timeBucketInterval time.Duration,
	includeDeviceFilter bool,
) string {
	// Convert time bucket interval to seconds for ClickHouse
	intervalSeconds := int(timeBucketInterval.Seconds())

	// Build device filter placeholder
	deviceFilter := ""
	if includeDeviceFilter {
		deviceFilter = "AND end_device_id IN (?)"
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
