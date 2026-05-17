package account

import (
	"context"

	"github.com/google/uuid"
	"github.com/netbill/auth-svc/internal/models"
)

type accountCache interface {
	Set(ctx context.Context, account models.Account)
	Get(ctx context.Context, accountID uuid.UUID) (models.Account, bool)
	Delete(ctx context.Context, accountID uuid.UUID)
}

type emailCache interface {
	Set(ctx context.Context, account models.AccountEmail)

	GetByID(ctx context.Context, accountID uuid.UUID) (models.AccountEmail, bool)
	GetByEmail(ctx context.Context, email string) (models.AccountEmail, bool)

	DeleteByID(ctx context.Context, accountID uuid.UUID)
	DeleteByEmail(ctx context.Context, email string)
}

type passwordCache interface {
	Set(ctx context.Context, account models.AccountPassword)
	Get(ctx context.Context, accountID uuid.UUID) (models.AccountPassword, bool)
	Delete(ctx context.Context, accountID uuid.UUID)
}

type sessionsCache interface {
	Delete(ctx context.Context, sessionID uuid.UUID)
}
