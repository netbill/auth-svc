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

func (r Repository) accountsQ() pgdb.AccountsQ {
	return pgdb.NewAccountsQ(r.db)
}

func (r Repository) sessionsQ() pgdb.SessionsQ {
	return pgdb.NewSessionsQ(r.db)
}

func (r Repository) passwordsQ() pgdb.AccountPasswordsQ {
	return pgdb.NewAccountPasswordsQ(r.db)
}

func (r Repository) emailsQ() pgdb.AccountEmailsQ {
	return pgdb.NewAccountEmailsQ(r.db)
}

func (r Repository) Transaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return pgx.Transaction(r.db, ctx, fn)
}
