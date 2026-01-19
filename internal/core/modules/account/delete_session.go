package account

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/netbill/auth-svc/internal/core/errx"
)

func (s Service) Logout(ctx context.Context, initiator InitiatorData) error {
	err := s.repo.DeleteAccountSession(ctx, initiator.AccountID, initiator.SessionID)
	if err != nil {
		return errx.ErrorInternal.Raise(
			fmt.Errorf("failed to delete session with id: %s, cause: %w", initiator.SessionID, err),
		)
	}

	return nil
}

func (s Service) DeleteOwnSession(ctx context.Context, initiator InitiatorData, sessionID uuid.UUID) error {
	_, _, err := s.ValidateSession(ctx, initiator)
	if err != nil {
		return err
	}

	err = s.repo.DeleteAccountSession(ctx, initiator.AccountID, sessionID)
	if err != nil {
		return errx.ErrorInternal.Raise(
			fmt.Errorf("failed to delete session with id: %s for account %s, cause: %w", sessionID, initiator.AccountID, err),
		)
	}

	return nil
}

func (s Service) DeleteOwnSessions(ctx context.Context, initiator InitiatorData) error {
	_, _, err := s.ValidateSession(ctx, initiator)
	if err != nil {
		return err
	}

	err = s.repo.DeleteSessionsForAccount(ctx, initiator.AccountID)
	if err != nil {
		return errx.ErrorInternal.Raise(
			fmt.Errorf("failed to delete sessions for account %s, cause: %w", initiator.AccountID, err),
		)
	}

	return nil
}
