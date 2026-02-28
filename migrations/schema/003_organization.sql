-- +migrate Up

CREATE TABLE organizations (
    id      UUID PRIMARY KEY NOT NULL DEFAULT uuid_generate_v4(),
    source_created_at  TIMESTAMPTZ NOT NULL,
    replica_created_at TIMESTAMPTZ NOT NULL DEFAULT (now() at time zone 'utc')
);

CREATE TABLE organization_members (
      id              UUID PRIMARY KEY NOT NULL,
      account_id      UUID NOT NULL REFERENCES accounts(id) ON DELETE RESTRICT,
      organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,

      source_created_at  TIMESTAMPTZ NOT NULL,
      replica_created_at TIMESTAMPTZ NOT NULL DEFAULT (now() AT TIME ZONE 'utc'),

      CONSTRAINT organization_members_unique_member UNIQUE (account_id, organization_id)
);

CREATE INDEX organization_members_account_id_idx ON organization_members(account_id);
CREATE INDEX organization_members_organization_id_idx ON organization_members(organization_id);

-- +migrate Down
DROP TABLE IF EXISTS organization_members;
DROP TABLE IF EXISTS organizations;