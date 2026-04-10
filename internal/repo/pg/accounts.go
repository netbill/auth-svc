package pg

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/netbill/auth-svc/internal/errx"
	"github.com/netbill/auth-svc/internal/models"
	"github.com/netbill/auth-svc/internal/modules/account"
	"github.com/netbill/pgdbx"
)

type AccountRepo struct {
	db *pgdbx.DB
}

func NewAccountRepo(db *pgdbx.DB) *AccountRepo {
	return &AccountRepo{db: db}
}

func (r *AccountRepo) Create(ctx context.Context, params account.RegistrationParams) (models.Account, error) {
	const query = `
		INSERT INTO accounts (username, role)
		VALUES ($1, $2)
		RETURNING id, username, role, version, created_at, updated_at, deleted_at`

	var a models.Account
	row := r.db.QueryRow(ctx, query, params.Username, params.Role)
	if err := row.Scan(&a.ID, &a.Username, &a.Role, &a.Version, &a.CreatedAt, &a.UpdatedAt, &a.DeletedAt); err != nil {
		return models.Account{}, err
	}

	return a, nil
}

func (r *AccountRepo) GetByID(ctx context.Context, accountID uuid.UUID) (models.Account, error) {
	const query = `
		SELECT id, username, role, version, created_at, updated_at, deleted_at
		FROM accounts
		WHERE id = $1 AND deleted_at IS NULL`

	var a models.Account
	row := r.db.QueryRow(ctx, query, accountID)
	if err := row.Scan(&a.ID, &a.Username, &a.Role, &a.Version, &a.CreatedAt, &a.UpdatedAt, &a.DeletedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.Account{}, errx.ErrorAccountNotFound.Raise(err)
		}
		return models.Account{}, err
	}

	return a, nil
}

func (r *AccountRepo) GetByUsername(ctx context.Context, username string) (models.Account, error) {
	const query = `
		SELECT id, username, role, version, created_at, updated_at, deleted_at
		FROM accounts
		WHERE username = $1 AND deleted_at IS NULL`

	var a models.Account
	row := r.db.QueryRow(ctx, query, username)
	if err := row.Scan(&a.ID, &a.Username, &a.Role, &a.Version, &a.CreatedAt, &a.UpdatedAt, &a.DeletedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.Account{}, errx.ErrorAccountNotFound.Raise(err)
		}
		return models.Account{}, err
	}

	return a, nil
}

func (r *AccountRepo) UpdateUsername(
	ctx context.Context,
	accountID uuid.UUID,
	newUsername string,
	version int32,
) (models.Account, error) {
	const query = `
		UPDATE accounts
		SET username = $1, version = version + 1, updated_at = now()
		WHERE id = $2 AND version = $3 AND deleted_at IS NULL
		RETURNING id, username, role, version, created_at, updated_at, deleted_at`

	var a models.Account
	row := r.db.QueryRow(ctx, query, newUsername, accountID, version)
	if err := row.Scan(&a.ID, &a.Username, &a.Role, &a.Version, &a.CreatedAt, &a.UpdatedAt, &a.DeletedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.Account{}, errx.ErrorAccountNotFound.Raise(err)
		}
		return models.Account{}, err
	}

	return a, nil
}

func (r *AccountRepo) Delete(ctx context.Context, accountID uuid.UUID) error {
	const query = `
		UPDATE accounts
		SET deleted_at = now(), updated_at = now()
		WHERE id = $1 AND deleted_at IS NULL`

	tag, err := r.db.Exec(ctx, query, accountID)
	if err != nil {
		return err
	}

	if tag.RowsAffected() == 0 {
		return errx.ErrorAccountNotFound.Raise(nil)
	}

	return nil
}
