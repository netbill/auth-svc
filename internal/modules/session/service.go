package session

import (
	"context"

	"github.com/google/uuid"
	"github.com/netbill/auth-svc/internal/models"
	"github.com/netbill/restkit/pagi"
)

type Service struct {
}

func (s *Service) GetMySession(
	ctx context.Context,
	actor models.AccountActor,
	sessionID uuid.UUID,
) (models.Session, error) {
	return models.Session{}, nil
}

func (s *Service) GetMySessions(
	ctx context.Context,
	actor models.AccountActor,
	limit, offset uint,
) (pagi.Page[[]models.Session], error) {
	return pagi.Page[[]models.Session]{}, nil
}

func (s *Service) Refresh(
	ctx context.Context,
	oldRefreshToken string,
) (models.TokensPair, error) {
	return models.TokensPair{}, nil
}

func (s *Service) Logout(
	ctx context.Context,
	actor models.AccountActor,
) error {
	return nil
}

func (s *Service) DeleteMySession(
	ctx context.Context,
	actor models.AccountActor,
	sessionID uuid.UUID,
) error {
	return nil
}

func (s *Service) DeleteMySessions(
	ctx context.Context,
	actor models.AccountActor,
) error {
	return nil
}
