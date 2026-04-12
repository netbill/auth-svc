package pg

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/netbill/auth-svc/internal/errx"
	"github.com/netbill/auth-svc/internal/models"
	"github.com/netbill/auth-svc/internal/modules/account"
	"github.com/netbill/pgdbx"
)

const (
	emailsTable = "account_emails"
	emailsCols  = "account_id, email, verified, version, created_at, updated_at, deleted_at"
)

type EmailRepo struct {
	db *pgdbx.DB
}

func NewEmailRepo(db *pgdbx.DB) *EmailRepo {
	return &EmailRepo{db: db}
}

func scanEmail(row pgx.Row) (e models.AccountEmail, err error) {
	err = row.Scan(
		&e.AccountID,
		&e.Email,
		&e.Verified,
		&e.Version,
		&e.CreatedAt,
		&e.UpdatedAt,
		&e.DeletedAt,
	)
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		return models.AccountEmail{}, errx.ErrorAccountNotFound.Raise(err)
	case err != nil:
		return models.AccountEmail{}, fmt.Errorf("scan email: %w", err)
	}
	return e, nil
}

func (r *EmailRepo) Create(ctx context.Context, params models.AccountEmail) (models.AccountEmail, error) {
	const query = `
		INSERT INTO ` + emailsTable + ` (account_id, email)
		VALUES ($1, $2)
		RETURNING ` + emailsCols

	result, err := scanEmail(r.db.QueryRow(ctx, query, params.AccountID, params.Email))
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return models.AccountEmail{}, errx.ErrorEmailAlreadyExist.Raise(err)
		}
		return models.AccountEmail{}, err
	}
	return result, nil
}

func (r *EmailRepo) GetByID(ctx context.Context, accountID uuid.UUID, optFns ...account.GetAccountOption) (models.AccountEmail, error) {
	opts := account.ApplyGetAccountOptions(optFns)

	query := `SELECT ` + emailsCols + ` FROM ` + emailsTable + ` WHERE account_id = $1`
	switch opts.Deleted {
	case account.DeletedFilterAll:
		// no additional filter
	case account.DeletedFilterDeleted:
		query += ` AND deleted_at IS NOT NULL`
	default: // DeletedFilterActive (0)
		query += ` AND deleted_at IS NULL`
	}

	return scanEmail(r.db.QueryRow(ctx, query, accountID))
}

func (r *EmailRepo) GetByEmail(ctx context.Context, email string) (models.AccountEmail, error) {
	const query = `
		SELECT ` + emailsCols + `
		FROM ` + emailsTable + `
		WHERE email = $1 AND deleted_at IS NULL`

	return scanEmail(r.db.QueryRow(ctx, query, email))
}

