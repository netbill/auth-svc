package auth

import (
	"context"

	"github.com/netbill/auth-svc/internal/core/models"
)

func (m *Module) GetMyAccountByID(ctx context.Context, actor models.AccountActor) (models.Account, error) {
	account, _, err := m.validateActorSession(ctx, actor)
	if err != nil {
		return models.Account{}, err
	}

	return account, nil
}

func (m *Module) GetMyAccountEmail(ctx context.Context, actor models.AccountActor) (models.AccountEmail, error) {
	_, _, err := m.validateActorSession(ctx, actor)
	if err != nil {
		return models.AccountEmail{}, err
	}

	return m.repo.GetAccountEmail(ctx, actor.ID)
}
