package organization

import (
	"context"

	"github.com/google/uuid"
	"github.com/netbill/auth-svc/internal/core/models"
)

type Module struct {
	repo repo
}

func New(repo repo) *Module {
	return &Module{repo: repo}
}

type repo interface {
	CreateOrgMember(ctx context.Context, member models.OrgMember) error
	DeleteOrgMember(ctx context.Context, memberID uuid.UUID) error
	DeleteOrgMembers(ctx context.Context, accountID uuid.UUID) error

	BuryOrgMember(ctx context.Context, orgMemberID uuid.UUID) error
	BuryOrgMembers(ctx context.Context, orgID uuid.UUID) error
	OrgMemberIsBuried(ctx context.Context, orgMemberID uuid.UUID) (bool, error)

	AccountIsBuried(ctx context.Context, accountID uuid.UUID) (bool, error)

	Transaction(ctx context.Context, fn func(ctx context.Context) error) error
}
