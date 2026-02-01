package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/netbill/auth-svc/internal/core/models"
)

type OrganizationMemberRow struct {
	ID             uuid.UUID `db:"id"`
	AccountID      uuid.UUID `db:"account_id"`
	OrganizationID uuid.UUID `db:"organization_id"`

	SourceCreatedAt  time.Time `db:"source_created_at"`
	ReplicaCreatedAt time.Time `db:"replica_created_at"`
}

func (o OrganizationMemberRow) IsNil() bool {
	return o.ID == uuid.Nil
}

func (o OrganizationMemberRow) ToModel() models.Member {
	return models.Member{
		ID:             o.ID,
		AccountID:      o.AccountID,
		OrganizationID: o.OrganizationID,
		CreatedAt:      o.SourceCreatedAt,
	}
}

type OrganizationMembersQ interface {
	New() OrganizationMembersQ
	Insert(ctx context.Context, input OrganizationMemberRow) (OrganizationMemberRow, error)
	Delete(ctx context.Context) error

	FilterByID(id uuid.UUID) OrganizationMembersQ
	FilterByAccountID(accountID uuid.UUID) OrganizationMembersQ

	Exists(ctx context.Context) (bool, error)
}

func (r *Repository) CreateOrgMember(ctx context.Context, member models.Member) error {
	_, err := r.orgMembersQ().Insert(ctx, OrganizationMemberRow{
		ID:              member.ID,
		AccountID:       member.AccountID,
		OrganizationID:  member.OrganizationID,
		SourceCreatedAt: member.CreatedAt,
	})
	if err != nil {
		return fmt.Errorf("failed to insert organization member, cause: %w", err)
	}

	return err
}

func (r *Repository) DeleteOrgMember(ctx context.Context, memberID uuid.UUID) error {
	err := r.orgMembersQ().FilterByID(memberID).Delete(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete organization member with id %s, cause: %w", memberID, err)
	}

	return nil
}

func (r *Repository) ExistOrgMemberByAccount(ctx context.Context, accountID uuid.UUID) (bool, error) {
	exist, err := r.orgMembersQ().FilterByID(accountID).Exists(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to check existence of organization member with account id %s, cause: %w", accountID, err)
	}

	return exist, nil
}
