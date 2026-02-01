package account

import (
	"context"

	"github.com/netbill/auth-svc/internal/core/models"
)

func (m *Module) UpdateUsername(ctx context.Context, initiator InitiatorData, newUsername string) (account models.Account, err error) {
	account, _, err = m.validateInitiatorSession(ctx, initiator)
	if err != nil {
		return models.Account{}, err
	}

	if err = m.checkUsernameRequirements(ctx, newUsername); err != nil {
		return models.Account{}, err
	}

	err = m.repo.Transaction(ctx, func(ctx context.Context) error {
		account, err = m.repo.UpdateAccountUsername(ctx, initiator.AccountID, newUsername)
		if err != nil {
			return err
		}

		if err = m.messenger.WriteAccountUsernameUpdated(ctx, account); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return models.Account{}, err
	}

	return account, nil
}
