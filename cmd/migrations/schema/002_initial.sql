-- +migrate Up
CREATE TYPE "account_role" AS ENUM (
    'admin',
    'moderator',
    'user'
);

CREATE TYPE "account_status" AS ENUM (
    'active',
    'suspended',
    'deactivated'
);

CREATE TABLE accounts (
    id         UUID           PRIMARY KEY NOT NULL DEFAULT uuid_generate_v4(),
    username   VARCHAR(32)    NOT NULL UNIQUE,
    role       account_role   DEFAULT 'user'   NOT NULL,
    status     account_status DEFAULT 'active' NOT NULL,

    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    username_updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE account_emails (
    account_id UUID        NOT NULL PRIMARY KEY NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    email      VARCHAR(32) NOT NULL UNIQUE,
    verified   BOOLEAN     NOT NULL DEFAULT FALSE,
    updated_at TIMESTAMPTZ  NOT NULL DEFAULT now(),
    created_at TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE TABLE account_passwords (
    account_id UUID      NOT NULL PRIMARY KEY NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    hash       TEXT      NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE sessions (
    id         UUID PRIMARY KEY NOT NULL DEFAULT uuid_generate_v4(),
    account_id UUID NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    hash_token TEXT NOT NULL,
    last_used  TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- +migrate Down
DROP TABLE IF EXISTS sessions CASCADE;
DROP TABLE IF EXISTS account_passwords CASCADE;
DROP TABLE IF EXISTS account_emails CASCADE;
DROP TABLE IF EXISTS accounts CASCADE;

DROP TABLE IF EXISTS outbox_events CASCADE;
DROP TABLE IF EXISTS inbox_events CASCADE;

DROP TYPE IF EXISTS account_role;
DROP TYPE IF EXISTS account_status;
DROP TYPE IF EXISTS outbox_event_status;
DROP TYPE IF EXISTS inbox_event_status;

DROP EXTENSION IF EXISTS "uuid-ossp";
