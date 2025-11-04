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
