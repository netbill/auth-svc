-- +migrate Up
CREATE TYPE "user_role" AS ENUM (
    'admin',
    'moderator',
    'user'
);

CREATE TABLE users (
    id          UUID      PRIMARY KEY DEFAULT uuid_generate_v4(),
    role        user_role NOT NULL DEFAULT 'user',
    username    VARCHAR(32) NOT NULL UNIQUE,
    pseudonym   VARCHAR(128),
    description VARCHAR(255),
    avatar_key  TEXT,
    version     INTEGER      NOT NULL DEFAULT 1 CHECK ( version > 0 ),

    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ
);

CREATE TABLE user_emails (
    user_id UUID         PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    email      VARCHAR(254) NOT NULL UNIQUE,
    verified   BOOLEAN      NOT NULL DEFAULT FALSE,
    version    INTEGER      NOT NULL DEFAULT 1 CHECK ( version > 0 ),

    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ
);

CREATE TABLE user_passwords (
    user_id UUID    PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    hash       TEXT    NOT NULL,
    version    INTEGER NOT NULL DEFAULT 1 CHECK ( version > 0 ),

    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ
);

CREATE TABLE sessions (
    id         UUID    PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID    NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    hash_token TEXT    NOT NULL UNIQUE,
    version    INTEGER NOT NULL DEFAULT 1 CHECK ( version > 0 ),

    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    last_used  TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX sessions_user_id_idx ON sessions(user_id);

-- Cascade soft-delete from users → user_emails, user_passwords
-- +migrate StatementBegin
CREATE OR REPLACE FUNCTION cascade_user_soft_delete()
RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
    IF OLD.deleted_at IS NULL AND NEW.deleted_at IS NOT NULL THEN
        UPDATE user_emails
        SET deleted_at = NEW.deleted_at
        WHERE user_id = NEW.id AND deleted_at IS NULL;

        UPDATE user_passwords
        SET deleted_at = NEW.deleted_at
        WHERE user_id = NEW.id AND deleted_at IS NULL;
    END IF;
    RETURN NEW;
END;
$$;
-- +migrate StatementEnd

CREATE TRIGGER tr_cascade_user_soft_delete
AFTER UPDATE ON users
FOR EACH ROW
EXECUTE FUNCTION cascade_user_soft_delete();

-- Prevent manual soft-delete of user_emails while user is active
-- +migrate StatementBegin
CREATE OR REPLACE FUNCTION forbid_manual_soft_delete_user_email()
RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
    IF OLD.deleted_at IS NULL AND NEW.deleted_at IS NOT NULL THEN
        IF pg_trigger_depth() > 0 THEN
            RETURN NEW;
        END IF;

        IF EXISTS (SELECT 1 FROM users WHERE id = NEW.user_id AND deleted_at IS NULL) THEN
            RAISE EXCEPTION 'cannot soft-delete user_emails while user % is active', NEW.user_id
                USING ERRCODE = '23503';
        END IF;
    END IF;
    RETURN NEW;
END;
$$;
-- +migrate StatementEnd

CREATE TRIGGER tr_forbid_manual_soft_delete_user_email
BEFORE UPDATE ON user_emails
FOR EACH ROW
EXECUTE FUNCTION forbid_manual_soft_delete_user_email();

-- Prevent manual soft-delete of user_passwords while user is active
-- +migrate StatementBegin
CREATE OR REPLACE FUNCTION forbid_manual_soft_delete_user_password()
RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
    IF OLD.deleted_at IS NULL AND NEW.deleted_at IS NOT NULL THEN
        IF pg_trigger_depth() > 0 THEN
            RETURN NEW;
        END IF;

        IF EXISTS (SELECT 1 FROM users WHERE id = NEW.user_id AND deleted_at IS NULL) THEN
            RAISE EXCEPTION 'cannot soft-delete user_passwords while user % is active', NEW.user_id
                USING ERRCODE = '23503';
        END IF;
    END IF;
    RETURN NEW;
END;
$$;
-- +migrate StatementEnd

CREATE TRIGGER tr_forbid_manual_soft_delete_user_password
BEFORE UPDATE ON user_passwords
FOR EACH ROW
EXECUTE FUNCTION forbid_manual_soft_delete_user_password();

-- +migrate Down
DROP TRIGGER IF EXISTS tr_forbid_manual_soft_delete_user_email ON user_emails;
DROP FUNCTION IF EXISTS forbid_manual_soft_delete_user_email();

DROP TRIGGER IF EXISTS tr_forbid_manual_soft_delete_user_password ON user_passwords;
DROP FUNCTION IF EXISTS forbid_manual_soft_delete_user_password();

DROP TRIGGER IF EXISTS tr_cascade_user_soft_delete ON users;
DROP FUNCTION IF EXISTS cascade_user_soft_delete();

DROP INDEX IF EXISTS sessions_user_id_idx;

DROP TABLE IF EXISTS sessions CASCADE;
DROP TABLE IF EXISTS user_passwords CASCADE;
DROP TABLE IF EXISTS user_emails CASCADE;
DROP TABLE IF EXISTS users CASCADE;

DROP TYPE IF EXISTS user_role;
