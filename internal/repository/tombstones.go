package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type TombstoneRow struct {
	ID         uuid.UUID `db:"id"`
	EntityType string    `db:"entity_type"`
	EntityID   uuid.UUID `db:"entity_id"`
	DeletedAt  time.Time `db:"deleted_at"`
}

type TombstonesSql interface {
	BuryAccount(ctx context.Context, accountID uuid.UUID) error
	BurySession(ctx context.Context, sessionID uuid.UUID) error
	BuryAccountSessions(ctx context.Context, accountID uuid.UUID) error

	AccountIsBuried(ctx context.Context, accountID uuid.UUID) (bool, error)
	SessionIsBuried(ctx context.Context, sessionID uuid.UUID) (bool, error)

	BuryOrgMember(ctx context.Context, orgMemberID uuid.UUID) error
	OrgMemberIsBuried(ctx context.Context, orgMemberID uuid.UUID) (bool, error)

	BuryOrganization(ctx context.Context, orgID uuid.UUID) error
	OrganizationIsBuried(ctx context.Context, orgID uuid.UUID) (bool, error)
}
