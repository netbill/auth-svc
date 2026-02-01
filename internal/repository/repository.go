package repository

import (
	"context"
)

type Repository struct {
	accountSql       AccountsQ
	accountEmails    AccountEmailsQ
	accountPasswords AccountPasswordsQ
	sessionsSql      SessionsQ
	orgMemberSql     OrganizationMembersQ

	Transactioner
}

type Transactioner interface {
	Transaction(ctx context.Context, fn func(ctx context.Context) error) error
}

func NewRepository(
	transactioner Transactioner,
	accountSql AccountsQ,
	accountEmails AccountEmailsQ,
	accountPasswords AccountPasswordsQ,
	sessionsSql SessionsQ,
	orgMemberSql OrganizationMembersQ,
) *Repository {
	return &Repository{
		accountSql:       accountSql,
		accountEmails:    accountEmails,
		accountPasswords: accountPasswords,
		sessionsSql:      sessionsSql,
		orgMemberSql:     orgMemberSql,
		Transactioner:    transactioner,
	}
}

func (r *Repository) accountsQ() AccountsQ {
	return r.accountSql.New()
}

func (r *Repository) accountEmailsQ() AccountEmailsQ {
	return r.accountEmails.New()
}

func (r *Repository) accountPasswordsQ() AccountPasswordsQ {
	return r.accountPasswords.New()
}

func (r *Repository) sessionsQ() SessionsQ {
	return r.sessionsSql.New()
}

func (r *Repository) orgMembersQ() OrganizationMembersQ {
	return r.orgMemberSql.New()
}
