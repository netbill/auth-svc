package auth

import (
	"context"

	"github.com/google/uuid"
	"github.com/netbill/auth-svc/internal/core/models"
)

func (m *Module) GetAccountByID(ctx context.Context, ID uuid.UUID) (models.Account, error) {
	return m.repo.GetAccountByID(ctx, ID)
}

func (m *Module) GetAccountByEmail(ctx context.Context, email string) (models.Account, error) {
	return m.repo.GetAccountByEmail(ctx, email)
}

func (m *Module) GetAccountByUsername(ctx context.Context, username string) (models.Account, error) {
	return m.repo.GetAccountByUsername(ctx, username)
}

func (m *Module) GetAccountEmail(ctx context.Context, ID uuid.UUID) (models.AccountEmail, error) {
	return m.repo.GetAccountEmail(ctx, ID)
}
