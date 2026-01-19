package account

import (
	"context"
	"fmt"

	"github.com/netbill/auth-svc/internal/core/errx"
	"golang.org/x/crypto/bcrypt"
)

func (s Service) UpdatePassword(
	ctx context.Context,
	initiator InitiatorData,
	oldPassword, newPassword string,
) error {
	account, _, err := s.ValidateSession(ctx, initiator)
	if err != nil {
		return err
	}

	passData, err := s.repo.GetAccountPassword(ctx, initiator.AccountID)
	if err != nil {
		return errx.ErrorInternal.Raise(
			fmt.Errorf("getting password for account %s, cause: %w", initiator.AccountID, err),
		)
	}
	if passData.IsNil() {
		return errx.ErrorInitiatorNotFound.Raise(
			fmt.Errorf("password for account %s not found, cause: %w", initiator.AccountID, err),
		)
	}

	if err = passData.CanChangePassword(); err != nil {
		return err
	}

	if err = s.checkAccountPassword(ctx, initiator.AccountID, oldPassword); err != nil {
		return err
	}

	if err = s.CheckPasswordRequirements(newPassword); err != nil {
		return err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return errx.ErrorInternal.Raise(
			fmt.Errorf("hashing new newPassword for account '%s', cause: %w", initiator.AccountID, err),
		)
	}

	return s.repo.Transaction(ctx, func(txCtx context.Context) error {
		_, err = s.repo.UpdateAccountPassword(ctx, initiator.AccountID, string(hash))
		if err != nil {
			return errx.ErrorInternal.Raise(
				fmt.Errorf("updating password for account %s, cause: %w", initiator.AccountID, err),
			)
		}

		err = s.repo.DeleteSessionsForAccount(ctx, account.ID)
		if err != nil {
			return errx.ErrorInternal.Raise(
				fmt.Errorf("deleting sessions for account %s after password change, cause: %w", initiator.AccountID, err),
			)
		}

		err = s.messenger.WriteAccountPasswordChanged(ctx, account)
		if err != nil {
			return errx.ErrorInternal.Raise(
				fmt.Errorf("failed to write account password changed event for account id: %s, cause: %w", initiator.AccountID, err),
			)
		}

		return nil
	})
}
