package repository

import (
	"context"
)

type Repository struct {
	AccountsQ      AccountsQ
	AccountEmailsQ AccountEmailsQ
	AccountPassQ   AccountPasswordsQ
	SessionsQ      SessionsQ
	OrgMembersQ    OrganizationMembersQ

	Transactioner
}

type Transactioner interface {
	Transaction(ctx context.Context, fn func(ctx context.Context) error) error
}
