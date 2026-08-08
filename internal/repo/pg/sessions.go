package pg

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/netbill/auth-svc/internal/errx"
	"github.com/netbill/auth-svc/internal/models"
	"github.com/netbill/auth-svc/internal/modules/session"
	"github.com/netbill/pgdbx"
	"github.com/netbill/restkit/pagi"
)

const (
	sessionsTable = "sessions"
	sessionsCols  = "id, user_id, version, created_at, updated_at, last_used, deleted_at"
)

type SessionRepo struct {
	db *pgdbx.DB
}

func NewSessionRepo(db *pgdbx.DB) *SessionRepo {
	return &SessionRepo{db: db}
}

func scanSession(row pgx.Row) (s models.Session, err error) {
	err = row.Scan(
		&s.ID,
		&s.UserID,
		&s.Version,
		&s.CreatedAt,
		&s.UpdatedAt,
		&s.LastUsed,
		&s.DeletedAt,
	)
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		return models.Session{}, errx.ErrorSessionNotFound.Raise(err)
	case err != nil:
		return models.Session{}, fmt.Errorf("scan session: %w", err)
	}

	return s, nil
}

func (r *SessionRepo) Create(
	ctx context.Context,
	sessionID, userID uuid.UUID,
	hashToken string,
) (models.Session, error) {
	const query = `
		INSERT INTO ` + sessionsTable + ` (id, user_id, hash_token)
		VALUES ($1, $2, $3)
		RETURNING ` + sessionsCols

	return scanSession(r.db.QueryRow(ctx, query, sessionID, userID, hashToken))
}

func (r *SessionRepo) GetByID(ctx context.Context, sessionID uuid.UUID) (models.Session, error) {
	const query = `
		SELECT ` + sessionsCols + `
		FROM ` + sessionsTable + `
		WHERE id = $1`

	return scanSession(r.db.QueryRow(ctx, query, sessionID))
}

func (r *SessionRepo) GetForUser(ctx context.Context, userID, sessionID uuid.UUID) (models.Session, error) {
	const query = `
		SELECT ` + sessionsCols + `
		FROM ` + sessionsTable + `
		WHERE id = $1 AND user_id = $2`

	return scanSession(r.db.QueryRow(ctx, query, sessionID, userID))
}

func (r *SessionRepo) GetListForUser(
	ctx context.Context,
	userID uuid.UUID,
	optFns ...session.ListSessionsOption,
) (pagi.Page[[]models.Session], error) {
	opts := session.ApplyListOptions(optFns)

	var deletedCond string
	switch opts.Deleted {
	case session.DeletedFilterActive:
		deletedCond = " AND deleted_at IS NULL"
	case session.DeletedFilterDeleted:
		deletedCond = " AND deleted_at IS NOT NULL"
	}

	orderDir := "DESC"
	if opts.LastUsed == session.LastUsedAsc {
		orderDir = "ASC"
	}

	const countQuery = `SELECT COUNT(*) FROM ` + sessionsTable + ` WHERE user_id = $1`
	const listQuery = `SELECT ` + sessionsCols + ` FROM ` + sessionsTable + ` WHERE user_id = $1`

	var total uint
	if err := r.db.QueryRow(ctx,
		countQuery+deletedCond,
		userID,
	).Scan(&total); err != nil {
		return pagi.Page[[]models.Session]{}, fmt.Errorf("count sessions: %w", err)
	}

	rows, err := r.db.Query(ctx,
		listQuery+deletedCond+
			" ORDER BY last_used "+orderDir+
			" LIMIT $2 OFFSET $3",
		userID, opts.Limit, opts.Offset,
	)
	if err != nil {
		return pagi.Page[[]models.Session]{}, fmt.Errorf("list sessions: %w", err)
	}
	defer rows.Close()

	sessions := make([]models.Session, 0, opts.Limit)

	for rows.Next() {
		var s models.Session
		s, err = scanSession(rows)
		if err != nil {
			return pagi.Page[[]models.Session]{}, fmt.Errorf("list sessions: %w", err)
		}
		sessions = append(sessions, s)
	}

	if err = rows.Err(); err != nil {
		return pagi.Page[[]models.Session]{}, fmt.Errorf("iterate sessions: %w", err)
	}

	page := uint(1)
	if opts.Limit > 0 {
		page = opts.Offset/opts.Limit + 1
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
		FROM ` + sessionsTable + `
		WHERE id = $1 AND deleted_at IS NULL`

	var hash string
	if err := r.db.QueryRow(ctx, query, sessionID).Scan(&hash); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", errx.ErrorSessionNotFound.Raise(err)
		}
		return "", fmt.Errorf("get session token: %w", err)
	}

	return hash, nil
}

func (r *SessionRepo) UpdateToken(ctx context.Context, sessionID uuid.UUID, token string) (models.Session, error) {
	const query = `
		UPDATE ` + sessionsTable + `
		SET
		    hash_token = $1,
		    version    = version + 1,
		    updated_at = now(),
		    last_used  = now()
		WHERE id = $2 AND deleted_at IS NULL
		RETURNING ` + sessionsCols

	return scanSession(r.db.QueryRow(ctx, query, token, sessionID))
}

func (r *SessionRepo) Delete(ctx context.Context, sessionID uuid.UUID) error {
	const query = `
		UPDATE ` + sessionsTable + `
		SET
		    deleted_at = now(),
		    updated_at = now(),
		    version    = version + 1
		WHERE id = $1 AND deleted_at IS NULL`

	tag, err := r.db.Exec(ctx, query, sessionID)
	if err != nil {
		return fmt.Errorf("delete session: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return errx.ErrorSessionNotFound.Raise(fmt.Errorf("session %v not found on delete", sessionID))
	}

	return nil
}

func (r *SessionRepo) DeleteOneForUser(ctx context.Context, userID, sessionID uuid.UUID) error {
	const query = `
		UPDATE ` + sessionsTable + `
		SET
		    deleted_at = now(),
		    updated_at = now(),
		    version    = version + 1
		WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL`

	tag, err := r.db.Exec(ctx, query, sessionID, userID)
	if err != nil {
		return fmt.Errorf("delete session for user: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return errx.ErrorSessionNotFound.Raise(fmt.Errorf("session %v not found for user %v on delete", sessionID, userID))
	}

	return nil
}

func (r *SessionRepo) DeleteManyForUser(
	ctx context.Context,
	userID uuid.UUID,
) ([]uuid.UUID, error) {
	const query = `
		UPDATE ` + sessionsTable + `
		SET
		    deleted_at = now(),
		    updated_at = now(),
		    version    = version + 1
		WHERE user_id = $1 AND deleted_at IS NULL
		RETURNING id`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("delete sessions for user: %w", err)
	}
	defer rows.Close()

	var ids []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err = rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scan deleted session id: %w", err)
		}
		ids = append(ids, id)
	}

	return ids, rows.Err()
}
