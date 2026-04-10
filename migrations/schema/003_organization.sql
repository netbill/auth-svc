-- +migrate Up

CREATE TABLE organizations (
    id      UUID PRIMARY KEY NOT NULL DEFAULT uuid_generate_v4(),

    created_at TIMESTAMPTZ NOT NULL,
    deleted_at TIMESTAMPTZ,
    sync_at    TIMESTAMPTZ NOT NULL DEFAULT (now() at time zone 'utc')
);

CREATE TABLE organization_members (
      id              UUID PRIMARY KEY NOT NULL,
      account_id      UUID NOT NULL REFERENCES accounts(id) ON DELETE RESTRICT,
      organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,

      created_at TIMESTAMPTZ NOT NULL,
      deleted_at TIMESTAMPTZ,
      sync_at    TIMESTAMPTZ NOT NULL DEFAULT (now() at time zone 'utc'),

      CONSTRAINT organization_members_unique_member UNIQUE (account_id, organization_id)
);

CREATE INDEX organization_members_account_id_idx ON organization_members(account_id);
CREATE INDEX organization_members_organization_id_idx ON organization_members(organization_id);

-- Cascade soft-delete from organizations → organization_members
-- +migrate StatementBegin
CREATE OR REPLACE FUNCTION cascade_organization_soft_delete()
RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
    IF OLD.deleted_at IS NULL AND NEW.deleted_at IS NOT NULL THEN
        UPDATE organization_members
        SET deleted_at = NEW.deleted_at
        WHERE organization_id = NEW.id AND deleted_at IS NULL;
    END IF;
    RETURN NEW;
END;
$$;
-- +migrate StatementEnd

CREATE TRIGGER tr_cascade_organization_soft_delete
AFTER UPDATE ON organizations
FOR EACH ROW
EXECUTE FUNCTION cascade_organization_soft_delete();

-- +migrate Down
DROP TRIGGER IF EXISTS tr_cascade_organization_soft_delete ON organizations;
DROP FUNCTION IF EXISTS cascade_organization_soft_delete();

DROP INDEX IF EXISTS organization_members_account_id_idx;
DROP INDEX IF EXISTS organization_members_organization_id_idx;

DROP TABLE IF EXISTS organization_members;
DROP TABLE IF EXISTS organizations;
