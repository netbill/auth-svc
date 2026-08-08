package pg

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/netbill/auth-svc/internal/errx"
	"github.com/netbill/auth-svc/internal/models"
	"github.com/netbill/auth-svc/internal/modules/account"
	"github.com/netbill/pgdbx"
)

const (
	accountsTable = "accounts"
	accountsCols  = "id, role, version, created_at, updated_at, deleted_at"
)

type AccountRepo struct {
	db *pgdbx.DB
}

func NewAccountRepo(db *pgdbx.DB) *AccountRepo {
	return &AccountRepo{db: db}
}

func scanAccount(row pgx.Row) (r models.Account, err error) {
	err = row.Scan(
		&r.ID,
		&r.Role,
		&r.Version,
		&r.CreatedAt,
		&r.UpdatedAt,
		&r.DeletedAt,
	)
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		return models.Account{}, errx.ErrorAccountNotFound.Raise(err)
	case err != nil:
		return models.Account{}, fmt.Errorf("scan account: %w", err)
		//case r.DeletedAt != nil:
		//	return models.Account{}, errx.ErrorAccountDeleted.Raise(fmt.Errorf("account %v is deleted", r.ID))
	}
	return r, nil
}

func (r *AccountRepo) Create(ctx context.Context, params account.RegistrationParams) (models.Account, error) {
	const query = `
		INSERT INTO ` + accountsTable + ` (role)
		VALUES ($1)
		RETURNING ` + accountsCols

	return scanAccount(r.db.QueryRow(ctx, query, params.Role))
}

func (r *AccountRepo) GetByID(ctx context.Context, accountID uuid.UUID) (models.Account, error) {
	const query = `
		SELECT ` + accountsCols + `
		FROM ` + accountsTable + `
		WHERE id = $1 AND deleted_at IS NULL`

	return scanAccount(r.db.QueryRow(ctx, query, accountID))
}

func (r *AccountRepo) Delete(ctx context.Context, accountID uuid.UUID) (models.Account, error) {
	const query = `
		UPDATE ` + accountsTable + `
		SET
			deleted_at = now(),
			updated_at = now(),
			version    = version + 1
		WHERE id = $1 AND deleted_at IS NULL
		RETURNING ` + accountsCols

	var a models.Account
	err := r.db.QueryRow(ctx, query, accountID).Scan(
		&a.ID, &a.Role, &a.Version, &a.CreatedAt, &a.UpdatedAt, &a.DeletedAt,
	)
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		return models.Account{}, errx.ErrorAccountNotFound.Raise(err)
	case err != nil:
		return models.Account{}, fmt.Errorf("delete account: %w", err)
	}
	return a, nil
}
