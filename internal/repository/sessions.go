package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/netbill/auth-svc/internal/core/errx"
	"github.com/netbill/auth-svc/internal/core/models"
	"github.com/netbill/restkit/pagi"
)

type SessionRow struct {
	ID        uuid.UUID `db:"id"`
	AccountID uuid.UUID `db:"account_id"`
	HashToken string    `db:"hash_token"`
	Version   int32     `db:"version"`
	LastUsed  time.Time `db:"last_used"`
	CreatedAt time.Time `db:"created_at"`
}

func (s SessionRow) IsNil() bool {
	return s.ID == uuid.Nil
}

func (s SessionRow) ToModel() models.Session {
	return models.Session{
		ID:        s.ID,
		AccountID: s.AccountID,
		Version:   s.Version,
		LastUsed:  s.LastUsed,
		CreatedAt: s.CreatedAt,
	}
}

type SessionsQ interface {
	New() SessionsQ
	Insert(ctx context.Context, input SessionRow) (SessionRow, error)
	Get(ctx context.Context) (SessionRow, error)
	Select(ctx context.Context) ([]SessionRow, error)

	UpdateToken(token string) SessionsQ
	UpdateOne(ctx context.Context) (SessionRow, error)

	Delete(ctx context.Context) error

	FilterID(id uuid.UUID) SessionsQ
	FilterAccountID(accountID uuid.UUID) SessionsQ
	OrderCreatedAt(asc bool) SessionsQ

	Page(limit, offset uint) SessionsQ
	Count(ctx context.Context) (uint, error)
}

func (r *Repository) CreateSession(ctx context.Context, sessionID, accountID uuid.UUID, hashToken string) (models.Session, error) {
	row, err := r.SessionsSql.New().Insert(ctx, SessionRow{
		ID:        sessionID,
		AccountID: accountID,
		HashToken: hashToken,
	})
	if err != nil {
		return models.Session{}, fmt.Errorf("failed to insert session, cause: %w", err)
	}

	return row.ToModel(), nil
}

func (r *Repository) GetSession(ctx context.Context, sessionID uuid.UUID) (models.Session, error) {
	row, err := r.SessionsSql.New().FilterID(sessionID).Get(ctx)
	switch {
	case row.IsNil():
		return models.Session{}, errx.ErrorSessionNotFound.Raise(
			fmt.Errorf("session with id %s not found", sessionID),
		)
	case err != nil:
		return models.Session{}, err
	}

	return row.ToModel(), nil
}

func (r *Repository) GetAccountSession(ctx context.Context, userID, sessionID uuid.UUID) (models.Session, error) {
	row, err := r.SessionsSql.New().
		FilterID(sessionID).
		FilterAccountID(userID).
		Get(ctx)
	switch {
	case row.IsNil():
		return models.Session{}, errx.ErrorSessionNotFound.Raise(
			fmt.Errorf("failed to get session with id %s for account %s, cause: %w", sessionID, userID, err),
		)
	case err != nil:
		return models.Session{}, err
	}

	return row.ToModel(), nil
}

func (r *Repository) GetSessionsForAccount(ctx context.Context, userID uuid.UUID, limit, offset uint) (pagi.Page[[]models.Session], error) {
	rows, err := r.SessionsSql.New().
		FilterAccountID(userID).
		OrderCreatedAt(false).
		Page(limit, offset).
		Select(ctx)
	if err != nil {
		return pagi.Page[[]models.Session]{}, fmt.Errorf(
			"failed to get sessions for account %s, cause: %w", userID, err,
		)
	}

	total, err := r.SessionsSql.New().
		FilterAccountID(userID).
		Count(ctx)
	if err != nil {
		return pagi.Page[[]models.Session]{}, fmt.Errorf(
			"failed to count sessions for account %s, cause: %w", userID, err,
		)
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

func (r *Repository) GetSessionToken(ctx context.Context, sessionID uuid.UUID) (string, error) {
	row, err := r.SessionsSql.New().FilterID(sessionID).Get(ctx)
	switch {
	case err != nil:
		return "", fmt.Errorf("failed to get session token for session %s, cause: %w", sessionID, err)
	case row.IsNil():
		return "", errx.ErrorSessionNotFound.Raise(
			fmt.Errorf("session with id %s not found", sessionID),
		)
	}

	return row.HashToken, nil
}

func (r *Repository) UpdateSessionToken(ctx context.Context, sessionID uuid.UUID, token string) (models.Session, error) {
	row, err := r.SessionsSql.New().
		FilterID(sessionID).
		UpdateToken(token).
		UpdateOne(ctx)
	switch {
	case err != nil:
		return models.Session{}, fmt.Errorf("failed to update session token for session %s, cause: %w", sessionID, err)
	case row.IsNil():
		return models.Session{}, errx.ErrorSessionNotFound.Raise(
			fmt.Errorf("session with id %s not found", sessionID),
		)
	}

	return row.ToModel(), nil
}

func (r *Repository) DeleteSession(ctx context.Context, sessionID uuid.UUID) error {
	err := r.SessionsSql.New().FilterID(sessionID).Delete(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete session with id %s, cause: %w", sessionID, err)
	}

	return nil
}

func (r *Repository) DeleteSessionsForAccount(ctx context.Context, userID uuid.UUID) error {
	err := r.SessionsSql.New().FilterAccountID(userID).Delete(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete sessions for account %s, cause: %w", userID, err)
	}

	return nil
}

func (r *Repository) DeleteAccountSession(ctx context.Context, userID, sessionID uuid.UUID) error {
	err := r.SessionsSql.New().
		FilterID(sessionID).
		FilterAccountID(userID).
		Delete(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete session with id %s, cause: %w", sessionID, err)
	}

	return nil
}
