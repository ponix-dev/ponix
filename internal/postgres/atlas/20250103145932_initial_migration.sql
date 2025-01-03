-- Create "systems" table
CREATE TABLE "systems" ("id" character(20) NOT NULL, "organization_id" text NOT NULL, "name" text NOT NULL, "status" integer NOT NULL, PRIMARY KEY ("id"));
-- Create "network_servers" table
CREATE TABLE "network_servers" ("id" character(20) NOT NULL, "system_id" character(20) NOT NULL, "name" text NOT NULL, "status" integer NOT NULL, "iot_platform" integer NOT NULL, PRIMARY KEY ("id"), CONSTRAINT "network_servers_system_id_fkey" FOREIGN KEY ("system_id") REFERENCES "systems" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION);
-- Create "fields" table
CREATE TABLE "fields" ("id" character(20) NOT NULL, PRIMARY KEY ("id"));
-- Create "grow_mediums" table
CREATE TABLE "grow_mediums" ("id" character(20) NOT NULL, "medium_type" integer NULL, PRIMARY KEY ("id"));
-- Create "tanks" table
CREATE TABLE "tanks" ("id" character(20) NOT NULL, PRIMARY KEY ("id"));
-- Create "system_inputs" table
CREATE TABLE "system_inputs" ("id" character(20) NOT NULL, "system_id" character(20) NOT NULL, "name" text NOT NULL, "status" integer NOT NULL, "grow_medium_id" character(20) NULL, "tank_id" character(20) NULL, "field_id" character(20) NULL, PRIMARY KEY ("id"), CONSTRAINT "system_inputs_field_id_fkey" FOREIGN KEY ("field_id") REFERENCES "fields" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, CONSTRAINT "system_inputs_grow_medium_id_fkey" FOREIGN KEY ("grow_medium_id") REFERENCES "grow_mediums" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, CONSTRAINT "system_inputs_system_id_fkey" FOREIGN KEY ("system_id") REFERENCES "systems" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, CONSTRAINT "system_inputs_tank_id_fkey" FOREIGN KEY ("tank_id") REFERENCES "tanks" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION);
-- Create "end_devices" table
CREATE TABLE "end_devices" ("id" character(20) NOT NULL, "system_id" character(20) NOT NULL, "network_server_id" character(20) NOT NULL, "system_input_id" character(20) NOT NULL, "name" text NOT NULL, "status" integer NOT NULL, PRIMARY KEY ("id"), CONSTRAINT "end_devices_network_server_id_fkey" FOREIGN KEY ("network_server_id") REFERENCES "network_servers" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, CONSTRAINT "end_devices_system_id_fkey" FOREIGN KEY ("system_id") REFERENCES "systems" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, CONSTRAINT "end_devices_system_input_id_fkey" FOREIGN KEY ("system_input_id") REFERENCES "system_inputs" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION);
-- Create "gateways" table
CREATE TABLE "gateways" ("id" character(20) NOT NULL, "system_id" character(20) NOT NULL, "network_server_id" character(20) NOT NULL, "name" text NOT NULL, "status" integer NOT NULL, PRIMARY KEY ("id"), CONSTRAINT "gateways_network_server_id_fkey" FOREIGN KEY ("network_server_id") REFERENCES "network_servers" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, CONSTRAINT "gateways_system_id_fkey" FOREIGN KEY ("system_id") REFERENCES "systems" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION);
-- Create "organizations" table
CREATE TABLE "organizations" ("id" character(20) NOT NULL, "name" text NOT NULL, "status" integer NOT NULL, PRIMARY KEY ("id"));
-- Create "users" table
CREATE TABLE "users" ("id" character(20) NOT NULL, "organization_id" character(20) NOT NULL, "first_name" text NOT NULL, "last_name" text NOT NULL, "status" integer NOT NULL, PRIMARY KEY ("id"), CONSTRAINT "users_organization_id_fkey" FOREIGN KEY ("organization_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION);
