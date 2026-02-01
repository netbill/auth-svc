package organization

import (
	"context"

	"github.com/netbill/auth-svc/internal/core/models"
)

func (m *Module) CreateOrgMember(ctx context.Context, member models.Member) error {
	err := m.repo.CreateOrgMember(ctx, member)
	if err != nil {
		return err
	}

	return nil
}
