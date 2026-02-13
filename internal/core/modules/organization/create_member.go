package organization

import (
	"context"

	"github.com/netbill/auth-svc/internal/core/models"
)

func (m *Module) CreateOrgMember(ctx context.Context, member models.OrgMember) error {
	return m.repo.CreateOrgMember(ctx, member)
}
