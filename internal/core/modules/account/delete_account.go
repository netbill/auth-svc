package account

import (
	"context"
	"fmt"

	"github.com/netbill/auth-svc/internal/core/errx"
)

func (m Module) DeleteOwnAccount(ctx context.Context, initiator InitiatorData) error {
	account, _, err := m.validateInitiatorSession(ctx, initiator)
	if err != nil {
		return err
	}

	exists, err := m.repo.ExistOrgMemberByAccount(ctx, initiator.AccountID)
	if err != nil {
		return err
	}
	if exists {
		return errx.AccountHaveMembershipInOrg.Raise(
			fmt.Errorf("account %s has a member of organizations", initiator.AccountID),
		)
	}

	return m.repo.Transaction(ctx, func(ctx context.Context) error {
		err = m.repo.DeleteAccount(ctx, initiator.AccountID)
		if err != nil {
			return err
		}

		err = m.messenger.WriteAccountDeleted(ctx, account.ID)
		if err != nil {
			return err
		}

		return nil
	})
}
