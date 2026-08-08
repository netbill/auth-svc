package pg

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/netbill/auth-svc/internal/errx"
	"github.com/netbill/auth-svc/internal/models"
	"github.com/netbill/auth-svc/internal/modules/user"
	"github.com/netbill/pgdbx"
)

const (
	usersTable = "users"
	usersCols  = "id, role, version, created_at, updated_at, deleted_at"
)

type UserRepo struct {
	db *pgdbx.DB
}

func NewUserRepo(db *pgdbx.DB) *UserRepo {
	return &UserRepo{db: db}
}

func scanUser(row pgx.Row) (r models.User, err error) {
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
		return models.User{}, errx.ErrorUserNotFound.Raise(err)
	case err != nil:
		return models.User{}, fmt.Errorf("scan user: %w", err)
		//case r.DeletedAt != nil:
		//	return models.User{}, errx.ErrorUserDeleted.Raise(fmt.Errorf("user %v is deleted", r.ID))
	}
	return r, nil
}

func (r *UserRepo) Create(ctx context.Context, params user.RegistrationParams) (models.User, error) {
	const query = `
		INSERT INTO ` + usersTable + ` (role)
		VALUES ($1)
		RETURNING ` + usersCols

	return scanUser(r.db.QueryRow(ctx, query, params.Role))
}

func (r *UserRepo) GetByID(ctx context.Context, userID uuid.UUID) (models.User, error) {
	const query = `
		SELECT ` + usersCols + `
		FROM ` + usersTable + `
		WHERE id = $1 AND deleted_at IS NULL`

	return scanUser(r.db.QueryRow(ctx, query, userID))
}

func (r *UserRepo) Delete(ctx context.Context, userID uuid.UUID) (models.User, error) {
	const query = `
		UPDATE ` + usersTable + `
		SET
			deleted_at = now(),
			updated_at = now(),
			version    = version + 1
		WHERE id = $1 AND deleted_at IS NULL
		RETURNING ` + usersCols

	var a models.User
	err := r.db.QueryRow(ctx, query, userID).Scan(
		&a.ID, &a.Role, &a.Version, &a.CreatedAt, &a.UpdatedAt, &a.DeletedAt,
	)
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		return models.User{}, errx.ErrorUserNotFound.Raise(err)
	case err != nil:
		return models.User{}, fmt.Errorf("delete user: %w", err)
	}
	return a, nil
}
