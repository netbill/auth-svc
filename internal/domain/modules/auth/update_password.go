package auth

import (
	"context"
	"fmt"

	"github.com/umisto/sso-svc/internal/domain/errx"
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

	passData, err := s.db.GetAccountPassword(ctx, initiator.AccountID)
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

	if err = s.db.Transaction(ctx, func(txCtx context.Context) error {
		_, err = s.db.UpdateAccountPassword(ctx, initiator.AccountID, string(hash))
		if err != nil {
			return errx.ErrorInternal.Raise(
				fmt.Errorf("updating password for account %s, cause: %w", initiator.AccountID, err),
			)
		}

		err = s.db.DeleteSessionsForAccount(ctx, account.ID)
		if err != nil {
			return errx.ErrorInternal.Raise(
				fmt.Errorf("deleting sessions for account %s after password change, cause: %w", initiator.AccountID, err),
			)
		}

		return nil
	}); err != nil {
		return err
	}

	email, err := s.GetAccountEmail(ctx, account.ID)
	if err != nil {
		return err
	}

	err = s.event.WriteAccountPasswordChanged(ctx, account, email.Email)
	if err != nil {
		return err
	}

	return nil
}
