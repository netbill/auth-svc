package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/umisto/pagi"
	"github.com/umisto/sso-svc/internal/domain/models"
	"github.com/umisto/sso-svc/internal/repository/pgdb"
)

func (r Repository) CreateSession(ctx context.Context, sessionID, accountID uuid.UUID, hashToken string) (models.Session, error) {
	now := time.Now().UTC()

	row := pgdb.Session{
		ID:        sessionID,
		AccountID: accountID,
		HashToken: hashToken,
		LastUsed:  now,
		CreatedAt: now,
	}

	err := r.sessionsQ().Insert(ctx, row)
	if err != nil {
		return models.Session{}, err
	}

	return row.ToModel(), nil
}

func (r Repository) GetSession(ctx context.Context, sessionID uuid.UUID) (models.Session, error) {
	row, err := r.sessionsQ().FilterID(sessionID).Get(ctx)
	switch {
	case err != nil:
		return models.Session{}, err
	case row.ID == uuid.Nil:
		return models.Session{}, nil
	}

	return row.ToModel(), nil
}

func (r Repository) GetAccountSession(ctx context.Context, userID, sessionID uuid.UUID) (models.Session, error) {
	row, err := r.sessionsQ().
		FilterID(sessionID).
		FilterAccountID(userID).
		Get(ctx)
	switch {
	case err != nil:
		return models.Session{}, err
	case row.ID == uuid.Nil:
		return models.Session{}, nil
	}

	return row.ToModel(), nil
}

func (r Repository) GetSessionsForAccount(ctx context.Context, userID uuid.UUID, limit, offset uint) (pagi.Page[[]models.Session], error) {
	rows, err := r.sessionsQ().
		FilterAccountID(userID).
		OrderCreatedAt(false).
		Page(limit, offset).
		Select(ctx)
	if err != nil {
		return pagi.Page[[]models.Session]{}, err
	}

	total, err := r.sessionsQ().
		FilterAccountID(userID).
		Count(ctx)
	if err != nil {
		return pagi.Page[[]models.Session]{}, err
	}

	collection := make([]models.Session, 0, len(rows))
	for _, s := range rows {
		collection = append(collection, s.ToModel())
	}

	return pagi.Page[[]models.Session]{
		Data:  collection,
		Page:  uint(offset/limit) + 1,
		Size:  uint(len(collection)),
		Total: total,
	}, nil
}

func (r Repository) GetSessionToken(ctx context.Context, sessionID uuid.UUID) (string, error) {
	row, err := r.sessionsQ().FilterID(sessionID).Get(ctx)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return "", nil
	case err != nil:
		return "", err
	}

	return row.HashToken, nil
}

func (r Repository) UpdateSessionToken(ctx context.Context, sessionID uuid.UUID, token string) (models.Session, error) {
	sess, err := r.sessionsQ().
		FilterID(sessionID).
		UpdateToken(token).
		Update(ctx)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return models.Session{}, nil
	case err != nil:
		return models.Session{}, err
	}

	if len(sess) != 1 {
		return models.Session{}, fmt.Errorf("expected 1 session, got %d", len(sess))
	}
	return sess[0].ToModel(), nil
}

func (r Repository) DeleteSession(ctx context.Context, sessionID uuid.UUID) error {
	return r.sessionsQ().FilterID(sessionID).Delete(ctx)
}

func (r Repository) DeleteSessionsForAccount(ctx context.Context, userID uuid.UUID) error {
	return r.sessionsQ().FilterAccountID(userID).Delete(ctx)
}

func (r Repository) DeleteAccountSession(ctx context.Context, userID, sessionID uuid.UUID) error {
	return r.sessionsQ().
		FilterID(sessionID).
		FilterAccountID(userID).
		Delete(ctx)
}

func toSessionModel(s pgdb.Session) models.Session {
	return models.Session{
		ID:        s.ID,
		AccountID: s.AccountID,
		CreatedAt: s.CreatedAt,
		LastUsed:  s.LastUsed,
	}
}
