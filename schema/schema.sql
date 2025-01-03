CREATE TABLE
    organizations (
        id CHAR(20) PRIMARY KEY,
        name text NOT NULL,
        status integer NOT NULL
    );

CREATE TABLE
    users (
        id CHAR(20) PRIMARY KEY,
        organization_id CHAR(20) NOT NULL REFERENCES organizations (id),
        first_name text NOT NULL,
        last_name text NOT NULL,
        status integer NOT NULL
    );

CREATE TABLE
    systems (
        id CHAR(20) PRIMARY KEY,
        organization_id text NOT NULL,
        name text NOT NULL,
        status integer NOT NULL
    );

CREATE TABLE
    grow_mediums (id CHAR(20) PRIMARY KEY, medium_type integer);

CREATE TABLE
    tanks (id CHAR(20) PRIMARY KEY);

CREATE TABLE
    fields (id CHAR(20) PRIMARY KEY);

CREATE TABLE
    system_inputs (
        id CHAR(20) PRIMARY KEY,
        system_id CHAR(20) NOT NULL REFERENCES systems (id),
        name text NOT NULL,
        status integer NOT NULL,
        grow_medium_id CHAR(20) REFERENCES grow_mediums (id),
        tank_id CHAR(20) REFERENCES tanks (id),
        field_id CHAR(20) REFERENCES fields (id)
    );

CREATE TABLE
    network_servers (
        id CHAR(20) PRIMARY KEY,
        system_id CHAR(20) NOT NULL REFERENCES systems (id),
        name text NOT NULL,
        status integer NOT NULL,
        iot_platform integer NOT NULL
    );

CREATE TABLE
    gateways (
        id CHAR(20) PRIMARY KEY,
        system_id CHAR(20) NOT NULL REFERENCES systems (id),
        network_server_id CHAR(20) NOT NULL REFERENCES network_servers (id),
        name text NOT NULL,
        status integer NOT NULL
    );

CREATE TABLE
    end_devices (
        id CHAR(20) PRIMARY KEY,
        system_id CHAR(20) NOT NULL REFERENCES systems (id),
        network_server_id CHAR(20) NOT NULL REFERENCES network_servers (id),
        system_input_id CHAR(20) NOT NULL REFERENCES system_inputs (id),
        name text NOT NULL,
        status integer NOT NULL
    );
