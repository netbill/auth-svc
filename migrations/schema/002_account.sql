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
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ
);

CREATE TABLE account_emails (
    account_id UUID         PRIMARY KEY REFERENCES accounts(id) ON DELETE CASCADE,
    email      VARCHAR(254) NOT NULL UNIQUE,
    verified   BOOLEAN      NOT NULL DEFAULT FALSE,
    version    INTEGER      NOT NULL DEFAULT 1 CHECK ( version > 0 ),

    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ
);

CREATE TABLE account_passwords (
    account_id UUID    PRIMARY KEY REFERENCES accounts(id) ON DELETE CASCADE,
    hash       TEXT    NOT NULL,
    version    INTEGER NOT NULL DEFAULT 1 CHECK ( version > 0 ),

    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ
);

CREATE TABLE sessions (
    id         UUID    PRIMARY KEY DEFAULT uuid_generate_v4(),
    account_id UUID    NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    hash_token TEXT    NOT NULL UNIQUE,
    version    INTEGER NOT NULL DEFAULT 1 CHECK ( version > 0 ),

    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    last_used  TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX sessions_account_id_idx ON sessions(account_id);

-- Cascade soft-delete from accounts → account_emails, account_passwords
-- +migrate StatementBegin
CREATE OR REPLACE FUNCTION cascade_account_soft_delete()
RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
    IF OLD.deleted_at IS NULL AND NEW.deleted_at IS NOT NULL THEN
        UPDATE account_emails
        SET deleted_at = NEW.deleted_at
        WHERE account_id = NEW.id AND deleted_at IS NULL;

        UPDATE account_passwords
        SET deleted_at = NEW.deleted_at
        WHERE account_id = NEW.id AND deleted_at IS NULL;
    END IF;
    RETURN NEW;
END;
$$;
-- +migrate StatementEnd

CREATE TRIGGER tr_cascade_account_soft_delete
AFTER UPDATE ON accounts
FOR EACH ROW
EXECUTE FUNCTION cascade_account_soft_delete();

-- Prevent manual soft-delete of account_emails while account is active
-- +migrate StatementBegin
CREATE OR REPLACE FUNCTION forbid_manual_soft_delete_account_email()
RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
    IF OLD.deleted_at IS NULL AND NEW.deleted_at IS NOT NULL THEN
        IF pg_trigger_depth() > 0 THEN
            RETURN NEW;
        END IF;

        IF EXISTS (SELECT 1 FROM accounts WHERE id = NEW.account_id AND deleted_at IS NULL) THEN
            RAISE EXCEPTION 'cannot soft-delete account_emails while account % is active', NEW.account_id
                USING ERRCODE = '23503';
        END IF;
    END IF;
    RETURN NEW;
END;
$$;
-- +migrate StatementEnd

CREATE TRIGGER tr_forbid_manual_soft_delete_account_email
BEFORE UPDATE ON account_emails
FOR EACH ROW
EXECUTE FUNCTION forbid_manual_soft_delete_account_email();

-- Prevent manual soft-delete of account_passwords while account is active
-- +migrate StatementBegin
CREATE OR REPLACE FUNCTION forbid_manual_soft_delete_account_password()
RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
    IF OLD.deleted_at IS NULL AND NEW.deleted_at IS NOT NULL THEN
        IF pg_trigger_depth() > 0 THEN
            RETURN NEW;
        END IF;

        IF EXISTS (SELECT 1 FROM accounts WHERE id = NEW.account_id AND deleted_at IS NULL) THEN
            RAISE EXCEPTION 'cannot soft-delete account_passwords while account % is active', NEW.account_id
                USING ERRCODE = '23503';
        END IF;
    END IF;
    RETURN NEW;
END;
$$;
-- +migrate StatementEnd

CREATE TRIGGER tr_forbid_manual_soft_delete_account_password
BEFORE UPDATE ON account_passwords
FOR EACH ROW
EXECUTE FUNCTION forbid_manual_soft_delete_account_password();

-- +migrate Down
DROP TRIGGER IF EXISTS tr_forbid_manual_soft_delete_account_email ON account_emails;
DROP FUNCTION IF EXISTS forbid_manual_soft_delete_account_email();

DROP TRIGGER IF EXISTS tr_forbid_manual_soft_delete_account_password ON account_passwords;
DROP FUNCTION IF EXISTS forbid_manual_soft_delete_account_password();

DROP TRIGGER IF EXISTS tr_cascade_account_soft_delete ON accounts;
DROP FUNCTION IF EXISTS cascade_account_soft_delete();

DROP INDEX IF EXISTS sessions_account_id_idx;

DROP TABLE IF EXISTS sessions CASCADE;
DROP TABLE IF EXISTS account_passwords CASCADE;
DROP TABLE IF EXISTS account_emails CASCADE;
DROP TABLE IF EXISTS accounts CASCADE;

DROP TYPE IF EXISTS account_role;
