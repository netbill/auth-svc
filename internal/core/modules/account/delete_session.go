package account

import (
	"context"

	"github.com/google/uuid"
)

func (m Module) Logout(ctx context.Context, initiator InitiatorData) error {
	err := m.repo.DeleteAccountSession(ctx, initiator.AccountID, initiator.SessionID)
	if err != nil {
		return err
	}

	return nil
}

func (m Module) DeleteOwnSession(ctx context.Context, initiator InitiatorData, sessionID uuid.UUID) error {
	_, _, err := m.validateInitiatorSession(ctx, initiator)
	if err != nil {
		return err
	}

	err = m.repo.DeleteAccountSession(ctx, initiator.AccountID, sessionID)
	if err != nil {
		return err
	}

	return nil
}

func (m Module) DeleteOwnSessions(ctx context.Context, initiator InitiatorData) error {
	_, _, err := m.validateInitiatorSession(ctx, initiator)
	if err != nil {
		return err
	}

	err = m.repo.DeleteSessionsForAccount(ctx, initiator.AccountID)
	if err != nil {
		return err
	}

	return nil
}
