package repository

import (
	"context"
)

type Repository struct {
	AccountsSql      AccountsQ
	AccountEmailsSql AccountEmailsQ
	AccountPassSql   AccountPasswordsQ
	SessionsSql      SessionsQ
	OrgMembersSql    OrganizationMembersQ
	OrganizationsSql OrganizationsQ
	TombstonesSql
	TransactionSql
}

type TransactionSql interface {
	Transaction(ctx context.Context, fn func(ctx context.Context) error) error
}

func (r *Repository) Transaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return r.TransactionSql.Transaction(ctx, fn)
}
