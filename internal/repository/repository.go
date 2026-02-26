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

	TransactionSql Transaction
}

type Transaction interface {
	Begin(ctx context.Context, fn func(ctx context.Context) error) error
}

func (r *Repository) Transaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return r.TransactionSql.Begin(ctx, fn)
}
