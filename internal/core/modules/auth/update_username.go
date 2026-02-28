package auth

import (
	"context"

	"github.com/netbill/auth-svc/internal/core/models"
)

func (m *Module) UpdateUsername(
	ctx context.Context,
	actor models.AccountActor,
	newUsername string,
) (account models.Account, err error) {
	account, _, err = m.validateActorSession(ctx, actor)
	if err != nil {
		return models.Account{}, err
	}

	if err = m.checkUsernameRequirements(ctx, newUsername); err != nil {
		return models.Account{}, err
	}

	err = m.repo.Transaction(ctx, func(ctx context.Context) error {
		account, err = m.repo.UpdateAccountUsername(ctx, actor.ID, newUsername)
		if err != nil {
			return err
		}
		return m.messenger.WriteAccountUsernameUpdated(ctx, account)
	})
	if err != nil {
		return models.Account{}, err
	}

	return account, nil
}
