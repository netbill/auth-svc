package auth

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/netbill/auth-svc/internal/errx"
	"github.com/netbill/auth-svc/internal/models"
)

//go:generate mockery --name=userRepo --inpackage
type userRepo interface {
	GetByID(ctx context.Context, userID uuid.UUID) (models.User, error)
}

//go:generate mockery --name=sessionRepo --inpackage
type sessionRepo interface {
	GetByID(ctx context.Context, sessionID uuid.UUID) (models.Session, error)
}

type Service struct {
	userRepo    userRepo
	sessionRepo sessionRepo
}

type ServiceDeps struct {
	UserRepo    userRepo
	SessionRepo sessionRepo
}

func New(deps ServiceDeps) *Service {
	return &Service{
		userRepo:    deps.UserRepo,
		sessionRepo: deps.SessionRepo,
	}
}

func (s *Service) ValidateSession(
	ctx context.Context,
	actor models.UserActor,
) (user models.User, session models.Session, err error) {
	user, err = s.userRepo.GetByID(ctx, actor.ID)
	switch {
	case errors.Is(err, errx.ErrorUserNotFound):
		return models.User{}, models.Session{}, errx.ErrorUserInvalidSession.Raise(
			fmt.Errorf("user %s not found", actor.ID),
		)
	case err != nil:
		return models.User{}, models.Session{}, err
	}

	session, err = s.sessionRepo.GetByID(ctx, actor.SessionID)
	switch {
	case errors.Is(err, errx.ErrorSessionNotFound):
		return models.User{}, models.Session{}, errx.ErrorUserInvalidSession.Raise(
			fmt.Errorf("session %s not found", actor.SessionID),
		)
	case err != nil:
		return models.User{}, models.Session{}, err
	}

	return user, session, nil
}
