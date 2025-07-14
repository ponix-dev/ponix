CREATE TABLE
    organizations (
        id CHAR(20) PRIMARY KEY,
        name text NOT NULL,
        status integer NOT NULL,
        created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
        updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
    );

CREATE TABLE
    users (
        id CHAR(20) PRIMARY KEY,
        organization_id CHAR(20) NOT NULL REFERENCES organizations (id),
        first_name text NOT NULL,
        last_name text NOT NULL,
        email text NOT NULL UNIQUE,
        created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
        updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
    );

-- LoRaWAN hardware types for device classification (maps to LoRaWANHardwareData)
CREATE TABLE lorawan_hardware_types (
    id CHAR(20) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    manufacturer VARCHAR(255) NOT NULL,
    model VARCHAR(255) NOT NULL,
    firmware_version VARCHAR(100),
    hardware_version VARCHAR(100),
    profile VARCHAR(255),
    lorawan_version INTEGER NOT NULL DEFAULT 0, -- maps to LORAWANVersion enum
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ -- soft delete timestamp
);

-- LoRaWAN frequency plans
CREATE TABLE lorawan_frequency_plans (
    id VARCHAR(50) PRIMARY KEY, -- e.g., "US_902_928", "EU_863_870"
    name VARCHAR(255) NOT NULL,
    description TEXT,
    region VARCHAR(100) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Generic end devices table (maps to EndDevice message)
CREATE TABLE end_devices (
    id CHAR(20) PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT,
    organization_id CHAR(20) NOT NULL REFERENCES organizations(id),
    status INTEGER NOT NULL DEFAULT 1, -- maps to EndDeviceStatus enum
    data_type INTEGER NOT NULL DEFAULT 0, -- maps to EndDeviceDataType enum
    hardware_type INTEGER NOT NULL DEFAULT 1, -- maps to EndDeviceHardwareType enum
    
    -- Audit fields
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- LoRaWAN-specific configuration (maps to LoRaWANConfig message)
CREATE TABLE lorawan_configs (
    id CHAR(20) PRIMARY KEY,
    end_device_id CHAR(20) NOT NULL REFERENCES end_devices(id) ON DELETE CASCADE,
    
    -- LoRaWAN identifiers
    device_eui CHAR(16) NOT NULL, -- 64-bit device EUI (hex)
    application_eui CHAR(16) NOT NULL, -- 64-bit application EUI (hex)
    application_id CHAR(20) NOT NULL, -- LoRaWAN application identifier
    
    -- LoRaWAN keys (encrypted storage recommended in production)
    application_key CHAR(32) NOT NULL, -- 128-bit application key (hex)
    network_key CHAR(32), -- 128-bit network key for LoRaWAN 1.1+ (hex)
    
    -- LoRaWAN configuration
    activation_method INTEGER NOT NULL DEFAULT 1, -- maps to ActivationMethod enum (OTAA default)
    frequency_plan_id VARCHAR(50) NOT NULL REFERENCES lorawan_frequency_plans(id),
    
    -- Hardware reference
    hardware_type_id CHAR(20) NOT NULL REFERENCES lorawan_hardware_types(id),
    
    -- Audit fields
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    -- Constraints
    CONSTRAINT unique_device_eui UNIQUE (device_eui),
    CONSTRAINT unique_end_device_id UNIQUE (end_device_id),
    CONSTRAINT valid_device_eui CHECK (device_eui ~ '^[0-9A-Fa-f]{16}$'),
    CONSTRAINT valid_application_eui CHECK (application_eui ~ '^[0-9A-Fa-f]{16}$'),
    CONSTRAINT valid_application_key CHECK (application_key ~ '^[0-9A-Fa-f]{32}$'),
    CONSTRAINT valid_network_key CHECK (network_key IS NULL OR network_key ~ '^[0-9A-Fa-f]{32}$')
);

-- Indexes for performance
CREATE INDEX idx_end_devices_org_id ON end_devices(organization_id);
CREATE INDEX idx_end_devices_status ON end_devices(status);
CREATE INDEX idx_end_devices_hardware_type ON end_devices(hardware_type);
CREATE INDEX idx_lorawan_configs_device_eui ON lorawan_configs(device_eui);
CREATE INDEX idx_lorawan_configs_app_id ON lorawan_configs(application_id);
CREATE INDEX idx_lorawan_configs_hardware_type ON lorawan_configs(hardware_type_id);

-- Insert default frequency plans
INSERT INTO lorawan_frequency_plans (id, name, description, region) VALUES
('US_902_928', 'US 902-928 MHz', 'North America LoRaWAN frequency plan', 'US'),
('EU_863_870', 'EU 863-870 MHz', 'European LoRaWAN frequency plan', 'EU'),
('AS_923', 'AS 923 MHz', 'Asia-Pacific LoRaWAN frequency plan', 'AS'),
('AU_915_928', 'AU 915-928 MHz', 'Australia LoRaWAN frequency plan', 'AU'),
('CN_470_510', 'CN 470-510 MHz', 'China LoRaWAN frequency plan', 'CN'),
('IN_865_867', 'IN 865-867 MHz', 'India LoRaWAN frequency plan', 'IN');