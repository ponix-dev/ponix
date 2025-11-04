-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS `processed_envelopes_new` (
  `organization_id` String NOT NULL,
  `end_device_id` String NOT NULL,
  `occurred_at` DateTime64(3, 'UTC') NOT NULL,
  `processed_at` DateTime64(3, 'UTC') NOT NULL,
  `data` JSON NOT NULL
) ENGINE = MergeTree()
PRIMARY KEY (organization_id, occurred_at, end_device_id)
ORDER BY (organization_id, occurred_at, end_device_id)
PARTITION BY toYYYYMM(occurred_at)
SETTINGS index_granularity = 8192;
-- +goose StatementEnd

-- +goose StatementBegin
INSERT INTO processed_envelopes_new
SELECT '' as organization_id, end_device_id, occurred_at, processed_at, data
FROM processed_envelopes;
-- +goose StatementEnd

-- +goose StatementBegin
RENAME TABLE processed_envelopes TO processed_envelopes_old;
-- +goose StatementEnd

-- +goose StatementBegin
RENAME TABLE processed_envelopes_new TO processed_envelopes;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE IF EXISTS processed_envelopes_old;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS `processed_envelopes_old` (
  `end_device_id` String NOT NULL,
  `occurred_at` DateTime64(3, 'UTC') NOT NULL,
  `processed_at` DateTime64(3, 'UTC') NOT NULL,
  `data` JSON NOT NULL
) ENGINE = MergeTree()
PRIMARY KEY (occurred_at, end_device_id)
ORDER BY (occurred_at, end_device_id)
PARTITION BY toYYYYMM(occurred_at)
SETTINGS index_granularity = 8192;
-- +goose StatementEnd

-- +goose StatementBegin
INSERT INTO processed_envelopes_old
SELECT end_device_id, occurred_at, processed_at, data
FROM processed_envelopes;
-- +goose StatementEnd

-- +goose StatementBegin
RENAME TABLE processed_envelopes TO processed_envelopes_new;
-- +goose StatementEnd

-- +goose StatementBegin
RENAME TABLE processed_envelopes_old TO processed_envelopes;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE IF EXISTS processed_envelopes_new;
-- +goose StatementEnd
