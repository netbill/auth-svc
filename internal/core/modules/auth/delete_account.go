package auth

import (
	"context"
	"fmt"

	"github.com/netbill/auth-svc/internal/core/errx"
	"github.com/netbill/auth-svc/internal/core/models"
)

func (m *Module) DeleteMyAccount(
	ctx context.Context,
	actor models.AccountActor,
) error {
	account, _, err := m.validateActorSession(ctx, actor)
	if err != nil {
		return err
	}

	exists, err := m.repo.ExistOrgMemberByAccount(ctx, actor.ID)
	if err != nil {
		return err
	}
	if exists {
		return errx.ErrorAccountHaveMembershipInOrg.Raise(
			fmt.Errorf("account %s has a member of organizations", actor.ID),
		)
	}

	return m.repo.Transaction(ctx, func(ctx context.Context) error {
		if err = m.repo.BuryAccount(ctx, account.ID); err != nil {
			return err
		}

		if err = m.repo.DeleteAccount(ctx, actor.ID); err != nil {
			return err
		}

		return m.messenger.WriteAccountDeleted(ctx, account.ID)
	})
}
