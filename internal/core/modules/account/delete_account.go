package account

import (
	"context"
	"fmt"

	"github.com/netbill/auth-svc/internal/core/errx"
)

func (s Service) DeleteOwnAccount(ctx context.Context, initiator InitiatorData) error {
	account, _, err := s.validateSession(ctx, initiator)
	if err != nil {
		return err
	}

	exists, err := s.repo.ExistOrgMemberByAccount(ctx, initiator.AccountID)
	if err != nil {
		return errx.ErrorInternal.Raise(
			fmt.Errorf("failed to check if account have membership in organizations, cause %s", err),
		)
	}

	if exists {
		return errx.AccountHaveMembershipInOrg.Raise(
			fmt.Errorf("account %s has a member of organizations", initiator.AccountID),
		)
	}

	return s.repo.Transaction(ctx, func(txCtx context.Context) error {
		err = s.repo.DeleteAccount(ctx, initiator.AccountID)
		if err != nil {
			return errx.ErrorInternal.Raise(
				fmt.Errorf("failed to delete account with id: %s, cause: %w", initiator.AccountID, err),
			)
		}

		err = s.messenger.WriteAccountDeleted(ctx, account.ID)
		if err != nil {
			return errx.ErrorInternal.Raise(
				fmt.Errorf("failed to write account deleted event for account id: %s, cause: %w", initiator.AccountID, err),
			)
		}

		return nil
	})
}
