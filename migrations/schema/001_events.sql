-- +migrate Up
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE outbox_events (
    event_id   UUID        PRIMARY KEY NOT NULL,
    topic      VARCHAR     NOT NULL,
    key        VARCHAR     NOT NULL,
    type       VARCHAR     NOT NULL,
    version    INT         NOT NULL,
    producer   VARCHAR     NOT NULL,
    payload    JSONB       NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT (now() AT TIME ZONE 'UTC')
);

-- +migrate Down
DROP TABLE IF EXISTS outbox_events;