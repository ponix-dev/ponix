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
