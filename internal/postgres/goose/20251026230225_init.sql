-- +goose Up
-- +goose StatementBegin
-- Create "casbin_rule" table
CREATE TABLE IF NOT EXISTS "casbin_rule" ("id" serial NOT NULL, "ptype" character varying(100) NULL, "v0" character varying(100) NULL, "v1" character varying(100) NULL, "v2" character varying(100) NULL, "v3" character varying(100) NULL, "v4" character varying(100) NULL, "v5" character varying(100) NULL, PRIMARY KEY ("id"));
-- Create "organizations" table
CREATE TABLE IF NOT EXISTS "organizations" ("id" character(20) NOT NULL, "name" text NOT NULL, "status" integer NOT NULL, "created_at" timestamptz NOT NULL DEFAULT now(), "updated_at" timestamptz NOT NULL DEFAULT now(), PRIMARY KEY ("id"));
-- Create "end_devices" table
CREATE TABLE IF NOT EXISTS "end_devices" ("id" character(20) NOT NULL, "name" text NOT NULL, "description" text NULL, "organization_id" character(20) NOT NULL, "status" integer NOT NULL DEFAULT 1, "data_type" integer NOT NULL DEFAULT 0, "hardware_type" integer NOT NULL DEFAULT 1, "created_at" timestamptz NOT NULL DEFAULT now(), "updated_at" timestamptz NOT NULL DEFAULT now(), PRIMARY KEY ("id"), CONSTRAINT "end_devices_organization_id_fkey" FOREIGN KEY ("organization_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION);
-- Create index "idx_end_devices_hardware_type" to table: "end_devices"
CREATE INDEX IF NOT EXISTS "idx_end_devices_hardware_type" ON "end_devices" ("hardware_type");
-- Create index "idx_end_devices_org_id" to table: "end_devices"
CREATE INDEX IF NOT EXISTS "idx_end_devices_org_id" ON "end_devices" ("organization_id");
-- Create index "idx_end_devices_status" to table: "end_devices"
CREATE INDEX IF NOT EXISTS "idx_end_devices_status" ON "end_devices" ("status");
-- Create "lorawan_frequency_plans" table
CREATE TABLE IF NOT EXISTS "lorawan_frequency_plans" ("id" character varying(50) NOT NULL, "name" character varying(255) NOT NULL, "description" text NULL, "region" character varying(100) NOT NULL, "created_at" timestamptz NOT NULL DEFAULT now(), PRIMARY KEY ("id"));
-- Create "lorawan_hardware_types" table
CREATE TABLE IF NOT EXISTS "lorawan_hardware_types" ("id" character(20) NOT NULL, "name" character varying(255) NOT NULL, "description" text NULL, "manufacturer" character varying(255) NOT NULL, "model" character varying(255) NOT NULL, "firmware_version" character varying(100) NULL, "hardware_version" character varying(100) NULL, "profile" character varying(255) NULL, "lorawan_version" integer NOT NULL DEFAULT 0, "created_at" timestamptz NOT NULL DEFAULT now(), "updated_at" timestamptz NOT NULL DEFAULT now(), "deleted_at" timestamptz NULL, PRIMARY KEY ("id"));
-- Create "lorawan_configs" table
CREATE TABLE IF NOT EXISTS "lorawan_configs" ("id" character(20) NOT NULL, "end_device_id" character(20) NOT NULL, "device_eui" character(16) NOT NULL, "application_eui" character(16) NOT NULL, "application_id" character(20) NOT NULL, "application_key" character(32) NOT NULL, "network_key" character(32) NULL, "activation_method" integer NOT NULL DEFAULT 1, "frequency_plan_id" character varying(50) NOT NULL, "hardware_type_id" character(20) NOT NULL, "created_at" timestamptz NOT NULL DEFAULT now(), "updated_at" timestamptz NOT NULL DEFAULT now(), PRIMARY KEY ("id"), CONSTRAINT "unique_device_eui" UNIQUE ("device_eui"), CONSTRAINT "unique_end_device_id" UNIQUE ("end_device_id"), CONSTRAINT "lorawan_configs_end_device_id_fkey" FOREIGN KEY ("end_device_id") REFERENCES "end_devices" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "lorawan_configs_frequency_plan_id_fkey" FOREIGN KEY ("frequency_plan_id") REFERENCES "lorawan_frequency_plans" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, CONSTRAINT "lorawan_configs_hardware_type_id_fkey" FOREIGN KEY ("hardware_type_id") REFERENCES "lorawan_hardware_types" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, CONSTRAINT "valid_application_eui" CHECK (application_eui ~ '^[0-9A-Fa-f]{16}$'::text), CONSTRAINT "valid_application_key" CHECK (application_key ~ '^[0-9A-Fa-f]{32}$'::text), CONSTRAINT "valid_device_eui" CHECK (device_eui ~ '^[0-9A-Fa-f]{16}$'::text), CONSTRAINT "valid_network_key" CHECK ((network_key IS NULL) OR (network_key ~ '^[0-9A-Fa-f]{32}$'::text)));
-- Create index "idx_lorawan_configs_app_id" to table: "lorawan_configs"
CREATE INDEX IF NOT EXISTS "idx_lorawan_configs_app_id" ON "lorawan_configs" ("application_id");
-- Create index "idx_lorawan_configs_device_eui" to table: "lorawan_configs"
CREATE INDEX IF NOT EXISTS "idx_lorawan_configs_device_eui" ON "lorawan_configs" ("device_eui");
-- Create index "idx_lorawan_configs_hardware_type" to table: "lorawan_configs"
CREATE INDEX IF NOT EXISTS "idx_lorawan_configs_hardware_type" ON "lorawan_configs" ("hardware_type_id");
-- Create "users" table
CREATE TABLE IF NOT EXISTS "users" ("id" character(20) NOT NULL, "first_name" text NOT NULL, "last_name" text NOT NULL, "email" text NOT NULL, "created_at" timestamptz NOT NULL DEFAULT now(), "updated_at" timestamptz NOT NULL DEFAULT now(), PRIMARY KEY ("id"), CONSTRAINT "users_email_key" UNIQUE ("email"));
-- Create "user_organizations" table
CREATE TABLE IF NOT EXISTS "user_organizations" ("id" serial NOT NULL, "user_id" character(20) NOT NULL, "organization_id" character(20) NOT NULL, "role" character varying(20) NOT NULL, "created_at" timestamptz NOT NULL DEFAULT now(), PRIMARY KEY ("id"), CONSTRAINT "user_organizations_organization_id_fkey" FOREIGN KEY ("organization_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "user_organizations_user_id_fkey" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- Create index "idx_user_organizations_org_id" to table: "user_organizations"
CREATE INDEX IF NOT EXISTS "idx_user_organizations_org_id" ON "user_organizations" ("organization_id");
-- Create index "idx_user_organizations_user_id" to table: "user_organizations"
CREATE INDEX IF NOT EXISTS "idx_user_organizations_user_id" ON "user_organizations" ("user_id");
-- +goose StatementEnd
