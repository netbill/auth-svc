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

type PasswordRepo struct {
	db *pgdbx.DB
}

func NewPasswordRepo(db *pgdbx.DB) *PasswordRepo {
	return &PasswordRepo{db: db}
}

func (r *PasswordRepo) Create(ctx context.Context, params models.AccountPassword) (models.AccountPassword, error) {
	const query = `
		INSERT INTO account_passwords (account_id, hash)
		VALUES ($1, $2)
		RETURNING account_id, hash, version, created_at, updated_at, deleted_at`

	var p models.AccountPassword
	row := r.db.QueryRow(ctx, query, params.AccountID, params.Hash)
	if err := row.Scan(&p.AccountID, &p.Hash, &p.Version, &p.CreatedAt, &p.UpdatedAt, &p.DeletedAt); err != nil {
		return models.AccountPassword{}, err
	}

	return p, nil
}

func (r *PasswordRepo) GetByID(ctx context.Context, accountID uuid.UUID) (models.AccountPassword, error) {
	const query = `
		SELECT account_id, hash, version, created_at, updated_at, deleted_at
		FROM account_passwords
		WHERE account_id = $1 AND deleted_at IS NULL`

	var p models.AccountPassword
	row := r.db.QueryRow(ctx, query, accountID)
	if err := row.Scan(&p.AccountID, &p.Hash, &p.Version, &p.CreatedAt, &p.UpdatedAt, &p.DeletedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.AccountPassword{}, errx.ErrorAccountNotFound.Raise(err)
		}
		return models.AccountPassword{}, err
	}

	return p, nil
}

func (r *PasswordRepo) UpdatePassword(ctx context.Context, accountID uuid.UUID, hash string) (models.AccountPassword, error) {
	const query = `
		UPDATE account_passwords
		SET hash = $1, version = version + 1, updated_at = now()
		WHERE account_id = $2 AND deleted_at IS NULL
		RETURNING account_id, hash, version, created_at, updated_at, deleted_at`

	var p models.AccountPassword
	row := r.db.QueryRow(ctx, query, hash, accountID)
	if err := row.Scan(&p.AccountID, &p.Hash, &p.Version, &p.CreatedAt, &p.UpdatedAt, &p.DeletedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.AccountPassword{}, errx.ErrorAccountNotFound.Raise(err)
		}
		return models.AccountPassword{}, err
	}

	return p, nil
}
