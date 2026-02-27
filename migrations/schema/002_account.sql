-- +migrate Up
CREATE TYPE "account_role" AS ENUM (
    'admin',
    'moderator',
    'user'
);

CREATE TABLE accounts (
    id         UUID         PRIMARY KEY DEFAULT uuid_generate_v4(),
    username   VARCHAR(32)  NOT NULL UNIQUE,
    role       account_role NOT NULL DEFAULT 'user',
    version    INTEGER      NOT NULL DEFAULT 1 CHECK ( version > 0 ),

    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE account_emails (
    account_id UUID        PRIMARY KEY REFERENCES accounts(id) ON DELETE CASCADE,
    email      VARCHAR(32) NOT NULL UNIQUE,
    verified   BOOLEAN     NOT NULL DEFAULT FALSE,
    version    INTEGER     NOT NULL DEFAULT 1 CHECK ( version > 0 ),

    created_at TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE TABLE account_passwords (
    account_id UUID    PRIMARY KEY REFERENCES accounts(id) ON DELETE CASCADE,
    hash       TEXT    NOT NULL,
    version    INTEGER NOT NULL DEFAULT 1 CHECK ( version > 0 ),

    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE sessions (
    id         UUID    PRIMARY KEY DEFAULT uuid_generate_v4(),
    account_id UUID    NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    hash_token TEXT    NOT NULL UNIQUE,
    version    INTEGER NOT NULL DEFAULT 1 CHECK ( version > 0 ),

    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    last_used  TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- +migrate StatementBegin
CREATE OR REPLACE FUNCTION forbid_delete_account_email_if_account_exists()
RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
    IF pg_trigger_depth() > 0 THEN
        RETURN OLD;
    END IF;

    IF EXISTS (SELECT 1 FROM accounts a WHERE a.id = OLD.account_id) THEN
        RAISE EXCEPTION 'cannot delete account_emails while account % exists', OLD.account_id
            USING ERRCODE = '23503';
    END IF;

    RETURN OLD;
END;
$$;
-- +migrate StatementEnd

CREATE TRIGGER tr_forbid_delete_account_email
BEFORE DELETE ON account_emails
FOR EACH ROW
EXECUTE FUNCTION forbid_delete_account_email_if_account_exists();

-- +migrate StatementBegin
CREATE OR REPLACE FUNCTION forbid_delete_account_password_if_account_exists()
RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
    IF pg_trigger_depth() > 0 THEN
        RETURN OLD;
    END IF;

    IF EXISTS (SELECT 1 FROM accounts a WHERE a.id = OLD.account_id) THEN
        RAISE EXCEPTION 'cannot delete account_passwords while account % exists', OLD.account_id
            USING ERRCODE = '23503';
    END IF;

    RETURN OLD;
END;
$$;
-- +migrate StatementEnd

CREATE TRIGGER tr_forbid_delete_account_password
BEFORE DELETE ON account_passwords
FOR EACH ROW
EXECUTE FUNCTION forbid_delete_account_password_if_account_exists();

CREATE TABLE tombstones (
    id           UUID        PRIMARY KEY DEFAULT uuid_generate_v4(),
    entity_type  VARCHAR(64) NOT NULL,  -- 'account', 'session', etc.
    entity_id    UUID        NOT NULL,
    deleted_at   TIMESTAMPTZ NOT NULL DEFAULT now(),

    UNIQUE (entity_type, entity_id)
);

-- +migrate Down
DROP TRIGGER IF EXISTS tr_forbid_delete_account_email ON account_emails;
DROP FUNCTION IF EXISTS forbid_delete_account_email_if_account_exists();

DROP TRIGGER IF EXISTS tr_forbid_delete_account_password ON account_passwords;
DROP FUNCTION IF EXISTS forbid_delete_account_password_if_account_exists();

DROP TABLE IF EXISTS sessions CASCADE;
DROP TABLE IF EXISTS account_passwords CASCADE;
DROP TABLE IF EXISTS account_emails CASCADE;
DROP TABLE IF EXISTS accounts CASCADE;

DROP TABLE IF EXISTS tombstones CASCADE;

DROP TYPE IF EXISTS account_role;
DROP TYPE IF EXISTS account_status;

DROP EXTENSION IF EXISTS "uuid-ossp";
