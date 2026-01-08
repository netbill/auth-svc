package auth

import (
	"context"
	"fmt"

	"github.com/netbill/auth-svc/internal/core/errx"
	"github.com/netbill/auth-svc/internal/core/models"
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

	if err = s.repo.Transaction(ctx, func(txCtx context.Context) error {
		account, err = s.repo.UpdateAccountUsername(ctx, initiator.AccountID, newUsername)
		if err != nil {
			return errx.ErrorInternal.Raise(
				fmt.Errorf("updating username for account %s, cause: %w", initiator.AccountID, err),
			)
		}

		err = s.repo.DeleteSessionsForAccount(ctx, account.ID)
		if err != nil {
			return errx.ErrorInternal.Raise(
				fmt.Errorf("deleting sessions for account %s after username change, cause: %w", initiator.AccountID, err),
			)
		}

		err = s.messenger.WriteAccountUsernameChanged(ctx, account)
		if err != nil {
			return errx.ErrorInternal.Raise(
				fmt.Errorf("failed to write account username changed event for account id: %s, cause: %w", initiator.AccountID, err),
			)
		}

		return nil
	}); err != nil {
		return models.Account{}, err
	}

	return account, nil
}
