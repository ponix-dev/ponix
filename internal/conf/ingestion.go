package conf

import "time"

type IngestionConfig struct {
	ClickHouseAddr                   string        `env:"CLICKHOUSE_ADDR"`
	ClickHouseUser                   string        `env:"CLICKHOUSE_USER"`
	ClickHousePass                   string        `env:"CLICKHOUSE_PASS"`
	ClickHouseDB                     string        `env:"CLICKHOUSE_DB"`
	ClickHouseProcessedEnvelopeTable string        `env:"CLICKHOUSE_PROCESSED_ENVELOPE_TABLE"`
	NatsURL                          string        `env:"NATS_URL"`
	NatsProcessedEnvelopeStream      string        `env:"NATS_PROCESSED_ENVELOPE_STREAM"`
	NatsProcessedEnvelopeSubject     string        `env:"NATS_PROCESSED_ENVELOPE_SUBJECT"`
	NatsProcessedEnvelopeBatchSize   int           `env:"NATS_PROCESSED_ENVELOPE_BATCH_SIZE"`
	NatsProcessedEnvelopeBatchWait   time.Duration `env:"NATS_PROCESSED_ENVELOPE_BATCH_WAIT"`
}
