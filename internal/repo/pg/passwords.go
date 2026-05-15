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
	passwordsTable = "account_passwords"
	passwordsCols  = "account_id, hash, version, created_at, updated_at, deleted_at"
)

type PasswordRepo struct {
	db *pgdbx.DB
}

func NewPasswordRepo(db *pgdbx.DB) *PasswordRepo {
	return &PasswordRepo{db: db}
}

func scanPassword(row pgx.Row) (p models.AccountPassword, err error) {
	err = row.Scan(
		&p.AccountID,
		&p.Hash,
		&p.Version,
		&p.CreatedAt,
		&p.UpdatedAt,
		&p.DeletedAt,
	)
	switch {
	case p.DeletedAt != nil:
		return models.AccountPassword{}, errx.ErrorAccountDeleted.Raise(fmt.Errorf("account %v is deleted", p.AccountID))
	case errors.Is(err, pgx.ErrNoRows):
		return models.AccountPassword{}, errx.ErrorAccountNotFound.Raise(err)
	case err != nil:
		return models.AccountPassword{}, fmt.Errorf("scan password: %w", err)
	}
	return p, nil
}

func (r *PasswordRepo) Create(ctx context.Context, params models.AccountPassword) (models.AccountPassword, error) {
	const query = `
		INSERT INTO ` + passwordsTable + ` (account_id, hash)
		VALUES ($1, $2)
		RETURNING ` + passwordsCols

	return scanPassword(r.db.QueryRow(ctx, query, params.AccountID, params.Hash))
}

func (r *PasswordRepo) GetByID(ctx context.Context, accountID uuid.UUID) (models.AccountPassword, error) {
	const query = `
		SELECT ` + passwordsCols + `
		FROM ` + passwordsTable + `
		WHERE account_id = $1 AND deleted_at IS NULL`

	return scanPassword(r.db.QueryRow(ctx, query, accountID))
}

func (r *PasswordRepo) UpdatePassword(ctx context.Context, accountID uuid.UUID, hash string) (models.AccountPassword, error) {
	const query = `
		UPDATE ` + passwordsTable + `
		SET hash = $1, version = version + 1, updated_at = now()
		WHERE account_id = $2 AND deleted_at IS NULL
		RETURNING ` + passwordsCols

	return scanPassword(r.db.QueryRow(ctx, query, hash, accountID))
}

func (r *PasswordRepo) Delete(ctx context.Context, accountID uuid.UUID) error {
	const query = `
		UPDATE ` + passwordsTable + `
		SET
			deleted_at = now(),
			updated_at = now(),
			version    = version + 1
		WHERE account_id = $1 AND deleted_at IS NULL`

	tag, err := r.db.Exec(ctx, query, accountID)
	if err != nil {
		return err
	}

	if tag.RowsAffected() == 0 {
		return errx.ErrorAccountNotFound.Raise(fmt.Errorf("account not found on password soft-delete"))
	}

	return nil
}
