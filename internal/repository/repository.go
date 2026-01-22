package repository

import (
	"context"
	"database/sql"

	"github.com/netbill/auth-svc/internal/repository/pgdb"
	"github.com/netbill/pgx"
)

type Repository struct {
	db *sql.DB
}

func New(db *sql.DB) *Repository {
	return &Repository{
		db: db,
	}
}

func (r Repository) accountsQ(ctx context.Context) pgdb.AccountsQ {
	return pgdb.NewAccountsQ(pgx.Exec(r.db, ctx))
}

func (r Repository) sessionsQ(ctx context.Context) pgdb.SessionsQ {
	return pgdb.NewSessionsQ(pgx.Exec(r.db, ctx))
}

func (r Repository) passwordsQ(ctx context.Context) pgdb.AccountPasswordsQ {
	return pgdb.NewAccountPasswordsQ(pgx.Exec(r.db, ctx))
}

func (r Repository) emailsQ(ctx context.Context) pgdb.AccountEmailsQ {
	return pgdb.NewAccountEmailsQ(pgx.Exec(r.db, ctx))
}

func (r Repository) orgMembersQ(ctx context.Context) pgdb.OrganizationMembersQ {
	return pgdb.NewOrganizationMembersQ(pgx.Exec(r.db, ctx))
}

func (r Repository) Transaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return pgx.Transaction(r.db, ctx, fn)
}
