package account

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/netbill/auth-svc/internal/core/errx"
	"github.com/netbill/auth-svc/internal/core/models"
	"github.com/netbill/restkit/pagi"
)

func (s Service) GetOwnSession(ctx context.Context, initiator InitiatorData, sessionID uuid.UUID) (models.Session, error) {
	_, _, err := s.validateSession(ctx, initiator)
	if err != nil {
		return models.Session{}, err
	}

	session, err := s.repo.GetAccountSession(ctx, initiator.AccountID, sessionID)
	if err != nil {
		return models.Session{}, errx.ErrorInternal.Raise(
			fmt.Errorf("failed to get session with id: %s for account %s, cause: %w", sessionID, initiator.AccountID, err),
		)
	}

	if session.IsNil() {
		return models.Session{}, errx.ErrorSessionNotFound.Raise(
			fmt.Errorf("session with id: %s for account %s not found", sessionID, initiator.AccountID),
		)
	}

	return session, nil
}

func (s Service) GetOwnSessions(
	ctx context.Context,
	initiator InitiatorData,
	limit, offset uint,
) (pagi.Page[[]models.Session], error) {
	_, _, err := s.validateSession(ctx, initiator)
	if err != nil {
		return pagi.Page[[]models.Session]{}, err
	}

	sessions, err := s.repo.GetSessionsForAccount(ctx, initiator.AccountID, limit, offset)
	if err != nil {
		return pagi.Page[[]models.Session]{}, errx.ErrorInternal.Raise(
			fmt.Errorf("failed to list sessions for account %s, cause: %w", initiator.AccountID, err),
		)
	}

	return sessions, nil
}
