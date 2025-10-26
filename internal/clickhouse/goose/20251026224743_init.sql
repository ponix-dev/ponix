-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS `processed_envelopes` (`end_device_id` String, `occurred_at` DateTime64(3, 'UTC'), `processed_at` DateTime64(3, 'UTC'), `data` JSON) ENGINE = MergeTree PRIMARY KEY (`occurred_at`, `end_device_id`) ORDER BY (`occurred_at`, `end_device_id`) SETTINGS index_granularity = 8192;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS `processed_envelopes`;
-- +goose StatementEnd
