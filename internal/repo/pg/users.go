package pg

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/netbill/auth-svc/internal/errx"
	"github.com/netbill/auth-svc/internal/models"
	"github.com/netbill/auth-svc/internal/modules/user"
	"github.com/netbill/pgdbx"
	"github.com/netbill/restkit/pagi"
)

const (
	usersTable = "users"
	usersCols  = "id, role, username, pseudonym, description, avatar_key, version, created_at, updated_at, deleted_at"
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
		&r.Username,
		&r.Pseudonym,
		&r.Description,
		&r.AvatarKey,
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
	}
	return r, nil
}

// nullIfEmpty maps an update param's cleared-value convention ("" clears the
// field) onto SQL NULL; pgx sends a nil *string as NULL and a non-nil one as
// its value.
func nullIfEmpty(v *string) *string {
	if v == nil || *v == "" {
		return nil
	}
	return v
}

func (r *UserRepo) Create(ctx context.Context, params user.RegistrationParams, username string) (models.User, error) {
	const query = `
		INSERT INTO ` + usersTable + ` (role, username)
		VALUES ($1, $2)
		RETURNING ` + usersCols

	res, err := scanUser(r.db.QueryRow(ctx, query, params.Role, username))
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" && pgErr.ConstraintName != "users_pkey" {
			return models.User{}, errx.ErrorUsernameTaken.Raise(err)
		}
		return models.User{}, fmt.Errorf("insert user, cause: %w", err)
	}

	return res, nil
}

func (r *UserRepo) GetByID(ctx context.Context, userID uuid.UUID) (models.User, error) {
	const query = `
		SELECT ` + usersCols + `
		FROM ` + usersTable + `
		WHERE id = $1 AND deleted_at IS NULL`

	return scanUser(r.db.QueryRow(ctx, query, userID))
}

func (r *UserRepo) GetByUsername(ctx context.Context, username string) (models.User, error) {
	const query = `
		SELECT ` + usersCols + `
		FROM ` + usersTable + `
		WHERE username = $1 AND deleted_at IS NULL`

	return scanUser(r.db.QueryRow(ctx, query, username))
}

func (r *UserRepo) ExistByUsername(ctx context.Context, username string) (bool, error) {
	const query = `SELECT EXISTS(SELECT 1 FROM ` + usersTable + ` WHERE username = $1 AND deleted_at IS NULL)`

	var exists bool
	if err := r.db.QueryRow(ctx, query, username).Scan(&exists); err != nil {
		return false, fmt.Errorf("failed to check user existence by username %s, cause: %w", username, err)
	}

	return exists, nil
}

func (r *UserRepo) Update(
	ctx context.Context,
	userID uuid.UUID,
	input user.UpdateParams,
) (models.User, error) {
	sets := []string{"updated_at = now()", "version = version + 1"}
	args := make([]interface{}, 0, 4)

	if input.Pseudonym != nil {
		args = append(args, nullIfEmpty(input.Pseudonym))
		sets = append(sets, fmt.Sprintf("pseudonym = $%d", len(args)))
	}
	if input.Description != nil {
		args = append(args, nullIfEmpty(input.Description))
		sets = append(sets, fmt.Sprintf("description = $%d", len(args)))
	}
	if input.AvatarKey != nil {
		args = append(args, nullIfEmpty(input.AvatarKey))
		sets = append(sets, fmt.Sprintf("avatar_key = $%d", len(args)))
	}

	args = append(args, userID)
	query := `UPDATE ` + usersTable + ` SET ` + strings.Join(sets, ", ") +
		fmt.Sprintf(" WHERE id = $%d AND deleted_at IS NULL RETURNING ", len(args)) + usersCols

	res, err := scanUser(r.db.QueryRow(ctx, query, args...))
	if err != nil {
		return models.User{}, fmt.Errorf("failed to update user by id %s, cause: %w", userID, err)
	}

	return res, nil
}

func (r *UserRepo) UpdateUsername(
	ctx context.Context,
	userID uuid.UUID,
	username string,
) (models.User, error) {
	const query = `
		UPDATE ` + usersTable + `
		SET username = $1, updated_at = now(), version = version + 1
		WHERE id = $2 AND deleted_at IS NULL
		RETURNING ` + usersCols

	res, err := scanUser(r.db.QueryRow(ctx, query, username, userID))
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return models.User{}, errx.ErrorUsernameTaken.Raise(err)
		}
		return models.User{}, fmt.Errorf("failed to update username by user id %s, cause: %w", userID, err)
	}

	return res, nil
}

// Filter powers the public user search. Ranking mirrors a simple relevance
// order: exact match first, then prefix match, then substring match,
// username taking priority over pseudonym within each tier.
func (r *UserRepo) Filter(
	ctx context.Context,
	params user.FilterParams,
	limit, offset uint,
) (pagi.Page[[]models.User], error) {
	if limit == 0 {
		limit = 10
	}

	where := " WHERE deleted_at IS NULL"
	var args []interface{}

	if params.Text != nil && *params.Text != "" {
		term := *params.Text
		like := "%" + term + "%"
		prefix := term + "%"

		where += ` AND (username ILIKE $1 OR pseudonym ILIKE $1)`
		args = []interface{}{like, term, prefix}
	}

	const countQuery = `SELECT COUNT(*) FROM ` + usersTable
	var total uint
	if err := r.db.QueryRow(ctx, countQuery+where, args...).Scan(&total); err != nil {
		return pagi.Page[[]models.User]{}, fmt.Errorf("failed to count users: %w", err)
	}

	order := " ORDER BY username ASC"
	if len(args) > 0 {
		order = ` ORDER BY
			CASE
				WHEN lower(username) = lower($2) THEN 0
				WHEN lower(pseudonym) = lower($2) THEN 1
				WHEN lower(username) LIKE lower($3) THEN 2
				WHEN lower(pseudonym) LIKE lower($3) THEN 3
				WHEN lower(username) LIKE lower($1) THEN 4
				WHEN lower(pseudonym) LIKE lower($1) THEN 5
				ELSE 6
			END,
			username ASC`
	}

	listQuery := `SELECT ` + usersCols + ` FROM ` + usersTable + where + order +
		fmt.Sprintf(" LIMIT $%d OFFSET $%d", len(args)+1, len(args)+2)

	rows, err := r.db.Query(ctx, listQuery, append(append([]interface{}{}, args...), limit, offset)...)
	if err != nil {
		return pagi.Page[[]models.User]{}, fmt.Errorf("failed to filter users: %w", err)
	}
	defer rows.Close()

	collection := make([]models.User, 0, limit)
	for rows.Next() {
		u, err := scanUser(rows)
		if err != nil {
			return pagi.Page[[]models.User]{}, fmt.Errorf("failed to filter users: %w", err)
		}
		collection = append(collection, u)
	}
	if err = rows.Err(); err != nil {
		return pagi.Page[[]models.User]{}, fmt.Errorf("failed to filter users: %w", err)
	}

	return pagi.Page[[]models.User]{
		Data:  collection,
		Page:  uint(offset/limit) + 1,
		Size:  uint(len(collection)),
		Total: total,
	}, nil
}

// Delete soft-deletes the user: the row is kept, and username is anonymized
// to free it up for reuse despite the UNIQUE constraint. A second call
// against an already-deleted row matches zero rows and surfaces as
// errx.ErrorUserNotFound.
func (r *UserRepo) Delete(ctx context.Context, userID uuid.UUID) (models.User, error) {
	// username is VARCHAR(32): "deleted_user" (12) + 20 hex chars fits exactly.
	const query = `
		UPDATE ` + usersTable + `
		SET
			username   = 'deleted_user' || substr(replace(gen_random_uuid()::text, '-', ''), 1, 20),
			deleted_at = now(),
			updated_at = now(),
			version    = version + 1
		WHERE id = $1 AND deleted_at IS NULL
		RETURNING ` + usersCols

	res, err := scanUser(r.db.QueryRow(ctx, query, userID))
	if err != nil {
		return models.User{}, fmt.Errorf("delete user %s, cause: %w", userID, err)
	}

	return res, nil
}
