package repository

import (
	"context"
	"database/sql"

	"github.com/umisto/pgx"
	"github.com/umisto/sso-svc/internal/repository/pgdb"
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
