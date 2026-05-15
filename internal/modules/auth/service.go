package auth

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

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

type Service struct {
	accountRepo accountRepo
	sessionRepo sessionRepo

	log *slog.Logger
}

type ServiceDeps struct {
	AccountRepo accountRepo
	SessionRepo sessionRepo

	Log *slog.Logger
}

func New(deps ServiceDeps) *Service {
	return &Service{
		accountRepo: deps.AccountRepo,
		sessionRepo: deps.SessionRepo,

		log: deps.Log,
	}
}

func (s *Service) ValidateSession(
	ctx context.Context,
	actor models.AccountActor,
) (account models.Account, session models.Session, err error) {
	account, err = s.accountRepo.GetByID(ctx, actor.ID)
	switch {
	case errors.Is(err, errx.ErrorAccountNotFound):
		return models.Account{}, models.Session{}, errx.ErrorAccountInvalidSession.Raise(
			fmt.Errorf("account %s not found", actor.ID),
		)
	case err != nil:
		return models.Account{}, models.Session{}, err
	}

	session, err = s.sessionRepo.GetByID(ctx, actor.SessionID)
	switch {
	case errors.Is(err, errx.ErrorSessionNotFound):
		return models.Account{}, models.Session{}, errx.ErrorAccountInvalidSession.Raise(
			fmt.Errorf("session %s not found", actor.SessionID),
		)
	case err != nil:
		return models.Account{}, models.Session{}, err
	}

	return account, session, nil
}
