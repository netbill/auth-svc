package pg

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/netbill/auth-svc/internal/errx"
	"github.com/netbill/auth-svc/internal/models"
	"github.com/netbill/pgdbx"
)

const (
	passwordsTable = "user_passwords"
	passwordsCols  = "user_id, hash, version, created_at, updated_at, deleted_at"
)

type PasswordRepo struct {
	db *pgdbx.DB
}

func NewPasswordRepo(db *pgdbx.DB) *PasswordRepo {
	return &PasswordRepo{db: db}
}

func scanPassword(row pgx.Row) (p models.UserPassword, err error) {
	err = row.Scan(
		&p.UserID,
		&p.Hash,
		&p.Version,
		&p.CreatedAt,
		&p.UpdatedAt,
		&p.DeletedAt,
	)
	switch {
	case p.DeletedAt != nil:
		return models.UserPassword{}, errx.ErrorUserDeleted.Raise(fmt.Errorf("user %v is deleted", p.UserID))
	case errors.Is(err, pgx.ErrNoRows):
		return models.UserPassword{}, errx.ErrorUserNotFound.Raise(err)
	case err != nil:
		return models.UserPassword{}, fmt.Errorf("scan password: %w", err)
	}
	return p, nil
}

func (r *PasswordRepo) Create(ctx context.Context, params models.UserPassword) (models.UserPassword, error) {
	const query = `
		INSERT INTO ` + passwordsTable + ` (user_id, hash)
		VALUES ($1, $2)
		RETURNING ` + passwordsCols

	return scanPassword(r.db.QueryRow(ctx, query, params.UserID, params.Hash))
}

func (r *PasswordRepo) GetByID(ctx context.Context, userID uuid.UUID) (models.UserPassword, error) {
	const query = `
		SELECT ` + passwordsCols + `
		FROM ` + passwordsTable + `
		WHERE user_id = $1 AND deleted_at IS NULL`

	return scanPassword(r.db.QueryRow(ctx, query, userID))
}

func (r *PasswordRepo) UpdatePassword(ctx context.Context, userID uuid.UUID, hash string) (models.UserPassword, error) {
	const query = `
		UPDATE ` + passwordsTable + `
		SET hash = $1, version = version + 1, updated_at = now()
		WHERE user_id = $2 AND deleted_at IS NULL
		RETURNING ` + passwordsCols

	return scanPassword(r.db.QueryRow(ctx, query, hash, userID))
}

func (r *PasswordRepo) Delete(ctx context.Context, userID uuid.UUID) error {
	const query = `
		UPDATE ` + passwordsTable + `
		SET
			deleted_at = now(),
			updated_at = now(),
			version    = version + 1
		WHERE user_id = $1 AND deleted_at IS NULL`

	tag, err := r.db.Exec(ctx, query, userID)
	if err != nil {
		return err
	}

	if tag.RowsAffected() == 0 {
		return errx.ErrorUserNotFound.Raise(fmt.Errorf("user not found on password soft-delete"))
	}

	return nil
}
