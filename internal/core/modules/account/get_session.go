package account

import (
	"context"

	"github.com/google/uuid"
	"github.com/netbill/auth-svc/internal/core/models"
	"github.com/netbill/restkit/pagi"
)

func (m Module) GetOwnSession(ctx context.Context, initiator InitiatorData, sessionID uuid.UUID) (models.Session, error) {
	_, _, err := m.validateInitiatorSession(ctx, initiator)
	if err != nil {
		return models.Session{}, err
	}

	session, err := m.repo.GetAccountSession(ctx, initiator.AccountID, sessionID)
	if err != nil {
		return models.Session{}, err
	}

	return session, nil
}

func (m Module) GetOwnSessions(
	ctx context.Context,
	initiator InitiatorData,
	limit, offset uint,
) (pagi.Page[[]models.Session], error) {
	_, _, err := m.validateInitiatorSession(ctx, initiator)
	if err != nil {
		return pagi.Page[[]models.Session]{}, err
	}

	sessions, err := m.repo.GetSessionsForAccount(ctx, initiator.AccountID, limit, offset)
	if err != nil {
		return pagi.Page[[]models.Session]{}, err
	}

	return sessions, nil
}
