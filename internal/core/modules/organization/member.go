package organization

import (
	"context"

	"github.com/google/uuid"
	"github.com/netbill/auth-svc/internal/core/models"
)

func (m *Module) CreateOrgMember(ctx context.Context, member models.OrgMember) error {
	buried, err := m.repo.OrgMemberIsBuried(ctx, member.ID)
	if err != nil {
		return err
	}
	if buried {
		return nil
	}

	buried, err = m.repo.AccountIsBuried(ctx, member.AccountID)
	if err != nil {
		return err
	}
	if buried {
		return nil
	}

	return m.repo.CreateOrgMember(ctx, member)
}

func (m *Module) DeleteOrgMembers(ctx context.Context, orgID uuid.UUID) error {
	return m.repo.Transaction(ctx, func(ctx context.Context) error {
		if err := m.repo.BuryOrgMembers(ctx, orgID); err != nil {
			return err
		}

		return m.repo.DeleteOrgMembers(ctx, orgID)
	})
}

func (m *Module) DeleteOrgMember(ctx context.Context, memberID uuid.UUID) error {
	return m.repo.Transaction(ctx, func(ctx context.Context) error {
		if err := m.repo.BuryOrgMember(ctx, memberID); err != nil {
			return err
		}

		return m.repo.DeleteOrgMember(ctx, memberID)
	})
}
