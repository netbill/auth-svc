package pg

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/netbill/auth-svc/internal/errx"
	"github.com/netbill/auth-svc/internal/models"
	"github.com/netbill/pgdbx"
)

type EmailRepo struct {
	db *pgdbx.DB
}

func NewEmailRepo(db *pgdbx.DB) *EmailRepo {
	return &EmailRepo{db: db}
}

func (r *EmailRepo) Create(ctx context.Context, params models.AccountEmail) (models.AccountEmail, error) {
	const query = `
		INSERT INTO account_emails (account_id, email)
		VALUES ($1, $2)
		RETURNING account_id, email, verified, version, created_at, updated_at, deleted_at`

	var e models.AccountEmail
	row := r.db.QueryRow(ctx, query, params.AccountID, params.Email)
	if err := row.Scan(&e.AccountID, &e.Email, &e.Verified, &e.Version, &e.CreatedAt, &e.UpdatedAt, &e.DeletedAt); err != nil {
		return models.AccountEmail{}, err
	}

	return e, nil
}

func (r *EmailRepo) GetByID(ctx context.Context, accountID uuid.UUID) (models.AccountEmail, error) {
	const query = `
		SELECT account_id, email, verified, version, created_at, updated_at, deleted_at
		FROM account_emails
		WHERE account_id = $1 AND deleted_at IS NULL`

	var e models.AccountEmail
	row := r.db.QueryRow(ctx, query, accountID)
	if err := row.Scan(&e.AccountID, &e.Email, &e.Verified, &e.Version, &e.CreatedAt, &e.UpdatedAt, &e.DeletedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.AccountEmail{}, errx.ErrorAccountNotFound.Raise(err)
		}
		return models.AccountEmail{}, err
	}

	return e, nil
}

func (r *EmailRepo) GetByEmail(ctx context.Context, email string) (models.AccountEmail, error) {
	const query = `
		SELECT account_id, email, verified, version, created_at, updated_at, deleted_at
		FROM account_emails
		WHERE email = $1 AND deleted_at IS NULL`

	var e models.AccountEmail
	row := r.db.QueryRow(ctx, query, email)
	if err := row.Scan(&e.AccountID, &e.Email, &e.Verified, &e.Version, &e.CreatedAt, &e.UpdatedAt, &e.DeletedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.AccountEmail{}, errx.ErrorAccountNotFound.Raise(err)
		}
		return models.AccountEmail{}, err
	}

	return e, nil
}

func (r *EmailRepo) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	const query = `
		SELECT EXISTS(
			SELECT 1 FROM account_emails
			WHERE email = $1 AND deleted_at IS NULL
		)`

	var exists bool
	if err := r.db.QueryRow(ctx, query, email).Scan(&exists); err != nil {
		return false, err
	}

	return exists, nil
}
