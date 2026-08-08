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
	"github.com/netbill/auth-svc/internal/modules/user"
	"github.com/netbill/pgdbx"
)

const (
	emailsTable = "user_emails"
	emailsCols  = "user_id, email, verified, version, created_at, updated_at, deleted_at"
)

type EmailRepo struct {
	db *pgdbx.DB
}

func NewEmailRepo(db *pgdbx.DB) *EmailRepo {
	return &EmailRepo{db: db}
}

func scanEmail(row pgx.Row) (e models.UserEmail, err error) {
	err = row.Scan(
		&e.UserID,
		&e.Email,
		&e.Verified,
		&e.Version,
		&e.CreatedAt,
		&e.UpdatedAt,
		&e.DeletedAt,
	)
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		return models.UserEmail{}, errx.ErrorUserNotFound.Raise(err)
	case err != nil:
		return models.UserEmail{}, fmt.Errorf("scan email: %w", err)
	}
	return e, nil
}

func (r *EmailRepo) Create(ctx context.Context, params models.UserEmail) (models.UserEmail, error) {
	const query = `
		INSERT INTO ` + emailsTable + ` (user_id, email)
		VALUES ($1, $2)
		RETURNING ` + emailsCols

	result, err := scanEmail(r.db.QueryRow(ctx, query, params.UserID, params.Email))
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return models.UserEmail{}, errx.ErrorEmailAlreadyExist.Raise(err)
		}
		return models.UserEmail{}, err
	}
	return result, nil
}

func (r *EmailRepo) GetByID(ctx context.Context, userID uuid.UUID, optFns ...user.GetUserOption) (models.UserEmail, error) {
	opts := user.ApplyGetUserOptions(optFns)

	query := `SELECT ` + emailsCols + ` FROM ` + emailsTable + ` WHERE user_id = $1`
	switch opts.Deleted {
	case user.DeletedFilterAll:
		// no additional filter
	case user.DeletedFilterDeleted:
		query += ` AND deleted_at IS NOT NULL`
	default: // DeletedFilterActive (0)
		query += ` AND deleted_at IS NULL`
	}

	return scanEmail(r.db.QueryRow(ctx, query, userID))
}

func (r *EmailRepo) GetByEmail(ctx context.Context, email string) (models.UserEmail, error) {
	const query = `
		SELECT ` + emailsCols + `
		FROM ` + emailsTable + `
		WHERE email = $1 AND deleted_at IS NULL`

	return scanEmail(r.db.QueryRow(ctx, query, email))
}
