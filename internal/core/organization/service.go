package organization

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/netbill/auth-svc/internal/errx"
	"github.com/netbill/auth-svc/internal/models"
)

type Service struct {
	org       orgRepo
	member    memberRepo
	tombstone tombstoneRepo
	tx        transaction
}

type OrgCoreDeps struct {
	Org       orgRepo
	Member    memberRepo
	Tombstone tombstoneRepo
	Tx        transaction
}

func NewOrgModule(deps OrgCoreDeps) *Service {
	return &Service{
		org:       deps.Org,
		member:    deps.Member,
		tombstone: deps.Tombstone,
		tx:        deps.Tx,
	}
}

func (s *Service) Create(
	ctx context.Context,
	organization models.Organization,
) error {
	buried, err := s.tombstone.OrganizationIsBuried(ctx, organization.ID)
	if err != nil {
		return err
	}
	if buried {
		return errx.ErrorOrganizationDeleted.Raise(
			fmt.Errorf("organization with id %s is already deleted", organization.ID),
		)
	}

	return s.org.Create(ctx, organization)
}

func (s *Service) Get(
	ctx context.Context,
	organizationID uuid.UUID,
) (models.Organization, error) {
	organization, err := s.org.Get(ctx, organizationID)
	if err != nil {
		return models.Organization{}, err
	}

	return organization, nil
}

func (s *Service) Delete(ctx context.Context, organizationID uuid.UUID) error {
	buried, err := s.tombstone.OrganizationIsBuried(ctx, organizationID)
	if err != nil {
		return err
	}
	if buried {
		return errx.ErrorOrganizationDeleted.Raise(
			fmt.Errorf("organization with id %s is already deleted", organizationID),
		)
	}

	return s.tx.Transaction(ctx, func(ctx context.Context) error {
		if err := s.tombstone.BuryOrganization(ctx, organizationID); err != nil {
			return err
		}

		if err := s.tombstone.BuryOrganization(ctx, organizationID); err != nil {
			return err
		}

		return s.org.Delete(ctx, organizationID)
	})
}

func (s *Service) CreateMember(ctx context.Context, member models.OrgMember) error {
	buried, err := s.tombstone.OrgMemberIsBuried(ctx, member.ID)
	if err != nil {
		return err
	}
	if buried {
		return errx.ErrorOrgMemberDeleted.Raise(
			fmt.Errorf("org member with id %s is already deleted", member.ID),
		)
	}

	buried, err = s.tombstone.AccountIsBuried(ctx, member.AccountID)
	if err != nil {
		return err
	}
	if buried {
		return errx.ErrorAccountDeleted.Raise(
			fmt.Errorf("account with id %s is already deleted", member.AccountID),
		)
	}

	buried, err = s.tombstone.OrganizationIsBuried(ctx, member.OrganizationID)
	if err != nil {
		return err
	}
	if buried {
		return errx.ErrorOrganizationDeleted.Raise(
			fmt.Errorf("organization with id %s is already deleted", member.OrganizationID),
		)
	}

	return s.member.Create(ctx, member)
}

func (s *Service) DeleteMember(ctx context.Context, memberID uuid.UUID) error {
	buried, err := s.tombstone.OrgMemberIsBuried(ctx, memberID)
	if err != nil {
		return err
	}
	if buried {
		return errx.ErrorOrgMemberDeleted.Raise(
			fmt.Errorf("org member with id %s is already deleted", memberID),
		)
	}

	buried, err = s.tombstone.AccountIsBuried(ctx, memberID)
	if err != nil {
		return err
	}
	if buried {
		return errx.ErrorAccountDeleted.Raise(
			fmt.Errorf("account with id %s is already deleted", memberID),
		)
	}

	return s.tx.Transaction(ctx, func(ctx context.Context) error {
		if err := s.tombstone.BuryOrgMember(ctx, memberID); err != nil {
			return err
		}

		return s.member.Delete(ctx, memberID)
	})
}
