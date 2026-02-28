package pg

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/netbill/auth-svc/internal/repository"
	"github.com/netbill/pgdbx"
)

const (
	EntityTypeAccount = "account"
	EntityTypeSession = "session"
)

func scanTombstone(row pgx.Row) (r repository.TombstoneRow, err error) {
	err = row.Scan(
		&r.ID,
		&r.EntityType,
		&r.EntityID,
		&r.DeletedAt,
	)
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		return repository.TombstoneRow{}, nil
	case err != nil:
		return repository.TombstoneRow{}, fmt.Errorf("scanning tombstone: %w", err)
	}
	return r, nil
}

type tombstones struct {
	db *pgdbx.DB
}

func NewTombstonesQ(db *pgdbx.DB) repository.TombstonesSql {
	return &tombstones{db: db}
}

func (t *tombstones) BuryAccount(ctx context.Context, accountID uuid.UUID) error {
	_, err := t.db.Exec(ctx, `
		INSERT INTO tombstones (entity_type, entity_id)
		SELECT 'account', $1
		UNION ALL
		SELECT 'session', s.id FROM sessions s WHERE s.account_id = $1
		ON CONFLICT (entity_type, entity_id) DO NOTHING
	`, accountID)
	if err != nil {
		return fmt.Errorf("burying account: %w", err)
	}

	return nil
}

func (t *tombstones) BurySession(ctx context.Context, sessionID uuid.UUID) error {
	_, err := t.db.Exec(ctx, `
		INSERT INTO tombstones (entity_type, entity_id)
		VALUES ('session', $1)
		ON CONFLICT (entity_type, entity_id) DO NOTHING
	`, sessionID)
	if err != nil {
		return fmt.Errorf("burying session: %w", err)
	}

	return nil
}

func (t *tombstones) BuryAccountSessions(ctx context.Context, accountID uuid.UUID) error {
	_, err := t.db.Exec(ctx, `
		INSERT INTO tombstones (entity_type, entity_id)
		SELECT 'session', s.id FROM sessions s WHERE s.account_id = $1
		ON CONFLICT (entity_type, entity_id) DO NOTHING
	`, accountID)
	if err != nil {
		return fmt.Errorf("burying account sessions: %w", err)
	}

	return nil
}

func (t *tombstones) BuryOrgMember(ctx context.Context, orgMemberID uuid.UUID) error {
	_, err := t.db.Exec(ctx, `
		INSERT INTO tombstones (entity_type, entity_id)
		VALUES ('organization_member', $1)
		ON CONFLICT (entity_type, entity_id) DO NOTHING
	`, orgMemberID)
	if err != nil {
		return fmt.Errorf("burying org member: %w", err)
	}

	return nil
}

func (t *tombstones) OrgMemberIsBuried(ctx context.Context, orgMemberID uuid.UUID) (bool, error) {
	var exists bool
	err := t.db.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM tombstones
			WHERE entity_type = $1 AND entity_id = $2
		)
	`, "organization_member", orgMemberID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("checking org member is buried: %w", err)
	}

	return exists, nil
}

func (t *tombstones) BuryOrganization(ctx context.Context, orgID uuid.UUID) error {
	_, err := t.db.Exec(ctx, `
		INSERT INTO tombstones (entity_type, entity_id)
		SELECT 'organization', $1
		UNION ALL
		SELECT 'organization_member', om.id FROM organization_members om WHERE om.organization_id = $1
		ON CONFLICT (entity_type, entity_id) DO NOTHING
	`, orgID)
	if err != nil {
		return fmt.Errorf("burying organization: %w", err)
	}

	return nil
}

func (t *tombstones) OrganizationIsBuried(ctx context.Context, orgID uuid.UUID) (bool, error) {
	var exists bool
	err := t.db.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM tombstones
			WHERE entity_type = 'organization' AND entity_id = $1
		)
	`, orgID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("checking organization is buried: %w", err)
	}

	return exists, nil
}

func (t *tombstones) AccountIsBuried(ctx context.Context, accountID uuid.UUID) (bool, error) {
	var exists bool
	err := t.db.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM tombstones
			WHERE entity_type = $1 AND entity_id = $2
		)
	`, EntityTypeAccount, accountID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("checking account is buried: %w", err)
	}

	return exists, nil
}

func (t *tombstones) SessionIsBuried(ctx context.Context, sessionID uuid.UUID) (bool, error) {
	var exists bool
	err := t.db.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM tombstones
			WHERE entity_type = $1 AND entity_id = $2
		)
	`, EntityTypeSession, sessionID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("checking session is buried: %w", err)
	}

	return exists, nil
}
