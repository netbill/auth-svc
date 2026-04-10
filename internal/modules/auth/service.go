package auth

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/netbill/auth-svc/internal/errx"
	"github.com/netbill/auth-svc/internal/models"
)

type accountRepo interface {
	GetByID(ctx context.Context, accountID uuid.UUID) (models.Account, error)
}

type sessionRepo interface {
	GetByID(ctx context.Context, sessionID uuid.UUID) (models.Session, error)
}

type accountCache interface {
	Set(ctx context.Context, account models.Account) error

	GetByID(ctx context.Context, accountID uuid.UUID) (models.Account, error)
}

type sessionsCache interface {
	Set(ctx context.Context, session models.Session) error

	GetByID(ctx context.Context, sessionID uuid.UUID) (models.Session, error)
}

type Service struct {
	accountRepo accountRepo
	sessionRepo sessionRepo

	accountCache  accountCache
	sessionsCache sessionsCache

	log *slog.Logger
}

func (s *Service) ValidateSession(
	ctx context.Context,
	actor models.AccountActor,
) (account models.Account, session models.Session, err error) {
	account, err = s.accountCache.GetByID(ctx, actor.ID)
	switch {
	case errors.Is(err, errx.ErrCacheMiss):
		s.log.Debug("account cache miss", "account_id", actor.ID)
	case err != nil:
		s.log.Error("failed to get account", "error", err)
	}

	if err != nil {
		account, err = s.accountRepo.GetByID(ctx, actor.ID)
		if err != nil {
			return models.Account{}, models.Session{}, err
		}
	}

	session, err = s.sessionsCache.GetByID(ctx, actor.SessionID)
	switch {
	case errors.Is(err, errx.ErrCacheMiss):
		s.log.Debug("session cache miss", "session_id", actor.SessionID)
	case err != nil:
		s.log.Error("failed to get session from cache", "error", err)
	}

	if err != nil {
		session, err = s.sessionRepo.GetByID(ctx, actor.SessionID)
		if err != nil {
			return models.Account{}, models.Session{}, err
		}
	}

	if session.AccountID != actor.ID {
		return models.Account{}, models.Session{}, errx.ErrorSessionNotFound.Raise(
			fmt.Errorf("session %s does not belong to account %s", session.ID, actor.ID),
		)
	}

	go func() {
		ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
		defer cancel()

		if err := s.accountCache.Set(ctx, account); err != nil {
			s.log.Error("failed to set account", "error", err)
		}

		if err := s.sessionsCache.Set(ctx, session); err != nil {
			s.log.Error("failed to set session", "error", err)
		}
	}()

	return account, session, nil
}
