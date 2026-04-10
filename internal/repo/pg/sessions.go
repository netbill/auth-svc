package pg

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/netbill/auth-svc/internal/errx"
	"github.com/netbill/auth-svc/internal/models"
	"github.com/netbill/pgdbx"
	"github.com/netbill/restkit/pagi"
)

type SessionRepo struct {
	db *pgdbx.DB
}

func NewSessionRepo(db *pgdbx.DB) *SessionRepo {
	return &SessionRepo{db: db}
}

func (r *SessionRepo) Create(
	ctx context.Context,
	sessionID, accountID uuid.UUID,
	hashToken string,
) (models.Session, error) {
	const query = `
		INSERT INTO sessions (id, account_id, hash_token)
		VALUES ($1, $2, $3)
		RETURNING id, account_id, version, created_at, last_used, deleted_at`

	var s models.Session
	row := r.db.QueryRow(ctx, query, sessionID, accountID, hashToken)
	if err := row.Scan(&s.ID, &s.AccountID, &s.Version, &s.CreatedAt, &s.LastUsed, &s.DeletedAt); err != nil {
		return models.Session{}, err
	}

	return s, nil
}

func (r *SessionRepo) GetByID(ctx context.Context, sessionID uuid.UUID) (models.Session, error) {
	const query = `
		SELECT id, account_id, version, created_at, last_used, deleted_at
		FROM sessions
		WHERE id = $1 AND deleted_at IS NULL`

	var s models.Session
	row := r.db.QueryRow(ctx, query, sessionID)
	if err := row.Scan(&s.ID, &s.AccountID, &s.Version, &s.CreatedAt, &s.LastUsed, &s.DeletedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.Session{}, errx.ErrorSessionNotFound.Raise(err)
		}
		return models.Session{}, err
	}

	return s, nil
}

func (r *SessionRepo) GetForAccount(ctx context.Context, accountID, sessionID uuid.UUID) (models.Session, error) {
	const query = `
		SELECT id, account_id, version, created_at, last_used, deleted_at
		FROM sessions
		WHERE id = $1 AND account_id = $2 AND deleted_at IS NULL`

	var s models.Session
	row := r.db.QueryRow(ctx, query, sessionID, accountID)
	if err := row.Scan(&s.ID, &s.AccountID, &s.Version, &s.CreatedAt, &s.LastUsed, &s.DeletedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.Session{}, errx.ErrorSessionNotFound.Raise(err)
		}
		return models.Session{}, err
	}

	return s, nil
}

func (r *SessionRepo) GetListForAccount(
	ctx context.Context,
	accountID uuid.UUID,
	limit, offset uint,
) (pagi.Page[[]models.Session], error) {
	const countQuery = `
		SELECT COUNT(*)
		FROM sessions
		WHERE account_id = $1 AND deleted_at IS NULL`

	var total uint
	if err := r.db.QueryRow(ctx, countQuery, accountID).Scan(&total); err != nil {
		return pagi.Page[[]models.Session]{}, err
	}

	const query = `
		SELECT id, account_id, version, created_at, last_used, deleted_at
		FROM sessions
		WHERE account_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.db.Query(ctx, query, accountID, limit, offset)
	if err != nil {
		return pagi.Page[[]models.Session]{}, err
	}
	defer rows.Close()

	sessions := make([]models.Session, 0, limit)
	for rows.Next() {
		var s models.Session
		if err = rows.Scan(&s.ID, &s.AccountID, &s.Version, &s.CreatedAt, &s.LastUsed, &s.DeletedAt); err != nil {
			return pagi.Page[[]models.Session]{}, err
		}
		sessions = append(sessions, s)
	}

	if err = rows.Err(); err != nil {
		return pagi.Page[[]models.Session]{}, err
	}

	page := uint(1)
	if limit > 0 {
		page = offset/limit + 1
	}

	return pagi.Page[[]models.Session]{
		Data:  sessions,
		Page:  page,
		Size:  uint(len(sessions)),
		Total: total,
	}, nil
}

func (r *SessionRepo) GetToken(ctx context.Context, sessionID uuid.UUID) (string, error) {
	const query = `
		SELECT hash_token
		FROM sessions
		WHERE id = $1 AND deleted_at IS NULL`

	var hash string
	if err := r.db.QueryRow(ctx, query, sessionID).Scan(&hash); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", errx.ErrorSessionNotFound.Raise(err)
		}
		return "", err
	}

	return hash, nil
}

func (r *SessionRepo) UpdateToken(ctx context.Context, sessionID uuid.UUID, token string) (models.Session, error) {
	const query = `
		UPDATE sessions
		SET hash_token = $1, version = version + 1, last_used = now()
		WHERE id = $2 AND deleted_at IS NULL
		RETURNING id, account_id, version, created_at, last_used, deleted_at`

	var s models.Session
	row := r.db.QueryRow(ctx, query, token, sessionID)
	if err := row.Scan(&s.ID, &s.AccountID, &s.Version, &s.CreatedAt, &s.LastUsed, &s.DeletedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.Session{}, errx.ErrorSessionNotFound.Raise(err)
		}
		return models.Session{}, err
	}

	return s, nil
}

func (r *SessionRepo) Delete(ctx context.Context, sessionID uuid.UUID) error {
	const query = `
		UPDATE sessions
		SET deleted_at = now()
		WHERE id = $1 AND deleted_at IS NULL`

	tag, err := r.db.Exec(ctx, query, sessionID)
	if err != nil {
		return err
	}

	if tag.RowsAffected() == 0 {
		return errx.ErrorSessionNotFound.Raise(nil)
	}

	return nil
}

func (r *SessionRepo) DeleteOneForAccount(ctx context.Context, accountID, sessionID uuid.UUID) error {
	const query = `
		UPDATE sessions
		SET deleted_at = now()
		WHERE id = $1 AND account_id = $2 AND deleted_at IS NULL`

	tag, err := r.db.Exec(ctx, query, sessionID, accountID)
	if err != nil {
		return err
	}

	if tag.RowsAffected() == 0 {
		return errx.ErrorSessionNotFound.Raise(nil)
	}

	return nil
}

func (r *SessionRepo) DeleteManyForAccount(ctx context.Context, accountID uuid.UUID) ([]uuid.UUID, error) {
	const query = `
		UPDATE sessions
		SET deleted_at = now()
		WHERE account_id = $1 AND deleted_at IS NULL
		RETURNING id`

	rows, err := r.db.Query(ctx, query, accountID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err = rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}

	return ids, rows.Err()
}
