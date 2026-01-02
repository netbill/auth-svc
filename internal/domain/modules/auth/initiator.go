package auth

import (
	"context"
	"fmt"

	"github.com/netbill/auth-svc/internal/domain/errx"
	"github.com/netbill/auth-svc/internal/domain/models"
)

func (s Service) ValidateSession(
	ctx context.Context,
	initiator InitiatorData,
) (models.Account, models.Session, error) {
	account, err := s.repo.GetAccountByID(ctx, initiator.AccountID)
	if err != nil {
		return models.Account{}, models.Session{}, errx.ErrorInitiatorNotFound.Raise(
			fmt.Errorf("failed to get account with id '%s', cause: %w", initiator.SessionID, err),
		)
	}
	if account.IsNil() {
		return models.Account{}, models.Session{}, errx.ErrorInitiatorNotFound.Raise(
			fmt.Errorf("account with id '%s' not found", initiator.SessionID),
		)
	}

	if err = account.CanInteract(); err != nil {
		return models.Account{}, models.Session{}, errx.ErrorInitiatorIsNotActive.Raise(
			fmt.Errorf("account with id '%s' cannot interact, cause: %w", initiator.AccountID, err),
		)
	}

	session, err := s.repo.GetSession(ctx, initiator.SessionID)
	if err != nil {
		return models.Account{}, models.Session{}, errx.ErrorInitiatorInvalidSession.Raise(
			fmt.Errorf("failed to get session with id '%s', cause: %w", initiator.SessionID, err),
		)
	}
	if session.IsNil() || session.AccountID != initiator.AccountID {
		return models.Account{}, models.Session{}, errx.ErrorInitiatorInvalidSession.Raise(
			fmt.Errorf("session with id '%s' not found for account '%s'", initiator.SessionID, initiator.AccountID),
		)
	}

	return account, session, nil
}
