package auth

import (
	"context"

	"github.com/netbill/auth-svc/internal/core/models"
)

func (m *Module) UpdatePassword(
	ctx context.Context,
	actor models.AccountActor,
	oldPassword, newPassword string,
) error {
	account, _, err := m.validateActorSession(ctx, actor)
	if err != nil {
		return err
	}

	passData, err := m.repo.GetAccountPassword(ctx, actor.ID)
	if err != nil {
		return err
	}

	if err = passData.CanChangePassword(); err != nil {
		return err
	}

	if err = m.checkAccountPassword(ctx, actor.ID, oldPassword); err != nil {
		return err
	}

	if err = m.password.CheckRequirements(newPassword); err != nil {
		return err
	}

	hash, err := m.password.GenerateHash(newPassword)
	if err != nil {
		return err
	}

	return m.repo.Transaction(ctx, func(ctx context.Context) error {
		_, err = m.repo.UpdateAccountPassword(ctx, actor.ID, hash)
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
