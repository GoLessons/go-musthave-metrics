CREATE SCHEMA IF NOT EXISTS "metrics";

CREATE TABLE IF NOT EXISTS "metrics"."metrics" (
    "name" text NOT NULL,
    "type" text NOT NULL,
    "delta" bigint DEFAULT NULL,
    "value" double precision,
    PRIMARY KEY ("name", "type")
);
