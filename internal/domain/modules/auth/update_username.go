package auth

import (
	"context"
	"fmt"

	"github.com/umisto/sso-svc/internal/domain/errx"
	"github.com/umisto/sso-svc/internal/domain/models"
)

func (s Service) UpdateUsername(
	ctx context.Context,
	initiator InitiatorData,
	password string,
	newUsername string,
) (models.Account, error) {
	account, _, err := s.ValidateSession(ctx, initiator)
	if err != nil {
		return models.Account{}, err
	}

	if err = account.CanChangeUsername(); err != nil {
		return models.Account{}, err
	}

	if err = s.CheckUsernameRequirements(newUsername); err != nil {
		return models.Account{}, err
	}

	if err = s.checkAccountPassword(ctx, initiator.AccountID, password); err != nil {
		return models.Account{}, err
	}

	if err = s.db.Transaction(ctx, func(txCtx context.Context) error {
		account, err = s.db.UpdateAccountUsername(ctx, initiator.AccountID, newUsername)
		if err != nil {
			return errx.ErrorInternal.Raise(
				fmt.Errorf("updating username for account %s, cause: %w", initiator.AccountID, err),
			)
		}

		err = s.db.DeleteSessionsForAccount(ctx, account.ID)
		if err != nil {
			return errx.ErrorInternal.Raise(
				fmt.Errorf("deleting sessions for account %s after username change, cause: %w", initiator.AccountID, err),
			)
		}

		return nil
	}); err != nil {
		return models.Account{}, err
	}

	email, err := s.GetAccountEmail(ctx, account.ID)
	if err != nil {
		return models.Account{}, err
	}

	err = s.event.WriteAccountUsernameChanged(ctx, account, email.Email)
	if err != nil {
		return models.Account{}, err
	}

	return account, nil
}
