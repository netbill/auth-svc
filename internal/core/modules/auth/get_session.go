package auth

import (
	"context"

	"github.com/google/uuid"
	"github.com/netbill/auth-svc/internal/core/models"
	"github.com/netbill/restkit/pagi"
)

func (m *Module) GetMySession(
	ctx context.Context,
	actor models.AccountActor,
	sessionID uuid.UUID,
) (models.Session, error) {
	_, _, err := m.validateActorSession(ctx, actor)
	if err != nil {
		return models.Session{}, err
	}

	session, err := m.repo.GetAccountSession(ctx, actor.ID, sessionID)
	if err != nil {
		return models.Session{}, err
	}

	return session, nil
}

func (m *Module) GetMySessions(
	ctx context.Context,
	actor models.AccountActor,
	limit, offset uint,
) (pagi.Page[[]models.Session], error) {
	_, _, err := m.validateActorSession(ctx, actor)
	if err != nil {
		return pagi.Page[[]models.Session]{}, err
	}

	sessions, err := m.repo.GetSessionsForAccount(ctx, actor.ID, limit, offset)
	if err != nil {
		return pagi.Page[[]models.Session]{}, err
	}

	return sessions, nil
}
