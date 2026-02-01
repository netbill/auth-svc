package account

import (
	"context"
	"fmt"

	"github.com/netbill/auth-svc/internal/core/errx"
	"golang.org/x/crypto/bcrypt"
)

func (m *Module) UpdatePassword(
	ctx context.Context,
	initiator InitiatorData,
	oldPassword, newPassword string,
) error {
	account, _, err := m.validateInitiatorSession(ctx, initiator)
	if err != nil {
		return err
	}

	passData, err := m.repo.GetAccountPassword(ctx, initiator.AccountID)
	if err != nil {
		return err
	}

	if err = passData.CanChangePassword(); err != nil {
		return err
	}

	if err = m.checkAccountPassword(ctx, initiator.AccountID, oldPassword); err != nil {
		return err
	}

	if err = m.checkPasswordRequirements(newPassword); err != nil {
		return err
	}

	//TODO remove from here
	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return errx.ErrorInternal.Raise(
			fmt.Errorf("hashing new newPassword for account '%s', cause: %w", initiator.AccountID, err),
		)
	}

	return m.repo.Transaction(ctx, func(ctx context.Context) error {
		_, err = m.repo.UpdateAccountPassword(ctx, initiator.AccountID, string(hash))
		if err != nil {
			return err
		}

		err = m.repo.DeleteSessionsForAccount(ctx, account.ID)
		if err != nil {
			return err
		}

		return nil
	})
}
