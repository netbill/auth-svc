package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/netbill/auth-svc/internal/core/models"
	"github.com/netbill/auth-svc/internal/repository/pgdb"
)

func (r Repository) CreateOrgMember(ctx context.Context, member models.Member) error {
	_, err := r.orgMembersQ(ctx).Insert(ctx, pgdb.OrganizationMemberInsertInput{
		ID:              member.ID,
		AccountID:       member.AccountID,
		OrganizationID:  member.OrganizationID,
		SourceCreatedAt: member.CreatedAt,
	})
	return err
}

func (r Repository) DeleteOrgMember(ctx context.Context, memberID uuid.UUID) error {
	return r.orgMembersQ(ctx).FilterByID(memberID).Delete(ctx)
}

func (r Repository) ExistOrgMemberByAccount(ctx context.Context, accountID uuid.UUID) (bool, error) {
	return r.orgMembersQ(ctx).FilterByID(accountID).Exists(ctx)
}
