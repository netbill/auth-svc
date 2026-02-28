package organization

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/netbill/auth-svc/internal/core/errx"
	"github.com/netbill/auth-svc/internal/core/models"
)

func (m *Module) CreateOrgMember(ctx context.Context, member models.OrgMember) error {
	buried, err := m.repo.OrgMemberIsBuried(ctx, member.ID)
	if err != nil {
		return err
	}
	if buried {
		return errx.ErrorOrgMemberDeleted.Raise(
			fmt.Errorf("org member with id %s is already deleted", member.ID),
		)
	}

	buried, err = m.repo.AccountIsBuried(ctx, member.AccountID)
	if err != nil {
		return err
	}
	if buried {
		return errx.ErrorAccountDeleted.Raise(
			fmt.Errorf("account with id %s is already deleted", member.AccountID),
		)
	}

	buried, err = m.repo.OrganizationIsBuried(ctx, member.OrganizationID)
	if err != nil {
		return err
	}
	if buried {
		return errx.ErrorOrganizationDeleted.Raise(
			fmt.Errorf("organization with id %s is already deleted", member.OrganizationID),
		)
	}

	return m.repo.CreateOrgMember(ctx, member)
}

func (m *Module) DeleteOrgMember(ctx context.Context, memberID uuid.UUID) error {
	buried, err := m.repo.OrgMemberIsBuried(ctx, memberID)
	if err != nil {
		return err
	}
	if buried {
		return errx.ErrorOrgMemberDeleted.Raise(
			fmt.Errorf("org member with id %s is already deleted", memberID),
		)
	}

	buried, err = m.repo.AccountIsBuried(ctx, memberID)
	if err != nil {
		return err
	}
	if buried {
		return errx.ErrorAccountDeleted.Raise(
			fmt.Errorf("account with id %s is already deleted", memberID),
		)
	}

	return m.repo.Transaction(ctx, func(ctx context.Context) error {
		if err := m.repo.BuryOrgMember(ctx, memberID); err != nil {
			return err
		}

		return m.repo.DeleteOrgMember(ctx, memberID)
	})
}
