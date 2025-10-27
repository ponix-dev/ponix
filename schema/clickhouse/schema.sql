CREATE TABLE `processed_envelopes` (
  `end_device_id` String NOT NULL,
  `occurred_at` DateTime64(3, 'UTC') NOT NULL,
  `processed_at` DateTime64(3, 'UTC') NOT NULL,
  `data` JSON NOT NULL,
  PRIMARY KEY (occurred_at,end_device_id)
) ENGINE = MergeTree()
ORDER BY (occurred_at,end_device_id)
PARTITION BY toYYYYMM(occurred_at);
