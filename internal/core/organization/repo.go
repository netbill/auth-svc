package organization

import (
	"context"

	"github.com/google/uuid"
	"github.com/netbill/auth-svc/internal/models"
)

type orgRepo interface {
	Create(ctx context.Context, organization models.Organization) error
	Get(ctx context.Context, organizationID uuid.UUID) (models.Organization, error)
	Delete(ctx context.Context, accountID uuid.UUID) error
}

type memberRepo interface {
	Create(ctx context.Context, member models.OrgMember) error
	Delete(ctx context.Context, memberID uuid.UUID) error
}

type tombstoneRepo interface {
	OrganizationIsBuried(ctx context.Context, organizationID uuid.UUID) (bool, error)
	BuryOrganization(ctx context.Context, organizationID uuid.UUID) error

	BuryOrgMember(ctx context.Context, orgMemberID uuid.UUID) error
	OrgMemberIsBuried(ctx context.Context, orgMemberID uuid.UUID) (bool, error)

	AccountIsBuried(ctx context.Context, accountID uuid.UUID) (bool, error)
}

type transaction interface {
	Transaction(ctx context.Context, fn func(ctx context.Context) error) error
}
