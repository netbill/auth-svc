package auth

import (
	"context"
	"fmt"

	"github.com/netbill/auth-svc/internal/domain/errx"
)

func (s Service) DeleteOwnAccount(ctx context.Context, initiator InitiatorData) error {
	account, _, err := s.ValidateSession(ctx, initiator)
	if err != nil {
		return err
	}

	return s.repo.Transaction(ctx, func(txCtx context.Context) error {
		err = s.repo.DeleteAccount(ctx, initiator.AccountID)
		if err != nil {
			return errx.ErrorInternal.Raise(
				fmt.Errorf("failed to delete account with id: %s, cause: %w", initiator.AccountID, err),
			)
		}

		err = s.messanger.WriteAccountDeleted(ctx, account)
		if err != nil {
			return errx.ErrorInternal.Raise(
				fmt.Errorf("failed to write account deleted event for account id: %s, cause: %w", initiator.AccountID, err),
			)
		}

		return nil
	})
}
