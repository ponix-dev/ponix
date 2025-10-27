package clickhouse

import (
	"context"
	"fmt"

	envelopev1 "buf.build/gen/go/ponix/ponix/protocolbuffers/go/envelope/v1"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
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
