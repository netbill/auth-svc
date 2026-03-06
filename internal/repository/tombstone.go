package repository

import (
	"context"

	"github.com/google/uuid"
)

type TombstonesSql interface {
	BuryAccount(ctx context.Context, accountID uuid.UUID) error
	BurySession(ctx context.Context, sessionID uuid.UUID) error
	BuryAccountSessions(ctx context.Context, accountID uuid.UUID) error

	AccountIsBuried(ctx context.Context, accountID uuid.UUID) (bool, error)
	SessionIsBuried(ctx context.Context, sessionID uuid.UUID) (bool, error)

	BuryOrgMember(ctx context.Context, orgMemberID uuid.UUID) error
	OrgMemberIsBuried(ctx context.Context, orgMemberID uuid.UUID) (bool, error)

	BuryOrganization(ctx context.Context, organizationID uuid.UUID) error
	OrganizationIsBuried(ctx context.Context, organizationID uuid.UUID) (bool, error)
}
