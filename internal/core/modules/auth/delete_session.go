package auth

import (
	"context"

	"github.com/google/uuid"
	"github.com/netbill/auth-svc/internal/core/models"
)

func (m *Module) Logout(
	ctx context.Context,
	actor models.AccountActor,
) error {
	return m.repo.Transaction(ctx, func(ctx context.Context) error {
		if err := m.repo.BurySession(ctx, actor.SessionID); err != nil {
			return err
		}

		return m.repo.DeleteAccountSession(ctx, actor.ID, actor.SessionID)
	})
}

func (m *Module) DeleteMySession(
	ctx context.Context,
	actor models.AccountActor,
	sessionID uuid.UUID,
) error {
	_, _, err := m.validateActorSession(ctx, actor)
	if err != nil {
		return err
	}

	return m.repo.Transaction(ctx, func(ctx context.Context) error {
		if err := m.repo.BurySession(ctx, sessionID); err != nil {
			return err
		}

		return m.repo.DeleteAccountSession(ctx, actor.ID, sessionID)
	})
}

func (m *Module) DeleteMySessions(ctx context.Context, actor models.AccountActor) error {
	_, _, err := m.validateActorSession(ctx, actor)
	if err != nil {
		return err
	}

	return m.repo.Transaction(ctx, func(ctx context.Context) error {
		if err := m.repo.BuryAccountSessions(ctx, actor.ID); err != nil {
			return err
		}

		return m.repo.DeleteSessionsForAccount(ctx, actor.ID)
	})
}
