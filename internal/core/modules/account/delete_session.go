package account

import (
	"context"

	"github.com/google/uuid"
	"github.com/netbill/auth-svc/internal/core/models"
)

func (m *Module) Logout(
	ctx context.Context,
	actor models.AccountActor,
) error {
	err := m.repo.DeleteAccountSession(ctx, actor.ID, actor.SessionID)
	if err != nil {
		return err
	}

	return nil
}

func (m *Module) DeleteOwnSession(
	ctx context.Context,
	actor models.AccountActor,
	sessionID uuid.UUID,
) error {
	_, _, err := m.validateActorSession(ctx, actor)
	if err != nil {
		return err
	}

	err = m.repo.DeleteAccountSession(ctx, actor.ID, sessionID)
	if err != nil {
		return err
	}

	return nil
}

func (m *Module) DeleteOwnSessions(ctx context.Context, actor models.AccountActor) error {
	_, _, err := m.validateActorSession(ctx, actor)
	if err != nil {
		return err
	}

	err = m.repo.DeleteSessionsForAccount(ctx, actor.ID)
	if err != nil {
		return err
	}

	return nil
}
