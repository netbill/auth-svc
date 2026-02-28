package organization

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/netbill/auth-svc/internal/core/errx"
	"github.com/netbill/auth-svc/internal/core/models"
)

func (m *Module) Create(
	ctx context.Context,
	organization models.Organization,
) error {
	buried, err := m.repo.OrganizationIsBuried(ctx, organization.ID)
	if err != nil {
		return err
	}
	if buried {
		return errx.ErrorOrganizationDeleted.Raise(
			fmt.Errorf("organization with id %s is already deleted", organization.ID),
		)
	}

	return m.repo.CreateOrganization(ctx, organization)
}

func (m *Module) Get(
	ctx context.Context,
	organizationID uuid.UUID,
) (models.Organization, error) {
	organization, err := m.repo.GetOrganizationByID(ctx, organizationID)
	if err != nil {
		return models.Organization{}, err
	}

	return organization, nil
}

func (m *Module) Delete(ctx context.Context, orgID uuid.UUID) error {
	buried, err := m.repo.OrganizationIsBuried(ctx, orgID)
	if err != nil {
		return err
	}
	if buried {
		return errx.ErrorOrganizationDeleted.Raise(
			fmt.Errorf("organization with id %s is already deleted", orgID),
		)
	}

	return m.repo.Transaction(ctx, func(ctx context.Context) error {
		if err := m.repo.BuryOrganization(ctx, orgID); err != nil {
			return err
		}

		if err := m.repo.BuryOrganization(ctx, orgID); err != nil {
			return err
		}

		return m.repo.DeleteOrganization(ctx, orgID)
	})
}
