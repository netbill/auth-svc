package session

import (
	"context"

	"github.com/google/uuid"
	"github.com/netbill/auth-svc/internal/models"
)

type passwordCache interface {
	Set(ctx context.Context, password models.AccountPassword)
	Get(ctx context.Context, accountID uuid.UUID) (models.AccountPassword, bool)
}

type accountCache interface {
	Set(ctx context.Context, account models.Account)
	Get(ctx context.Context, accountID uuid.UUID) (models.Account, bool)
}

type sessionsCache interface {
	Set(ctx context.Context, session models.Session)
	Get(ctx context.Context, sessionID uuid.UUID) (models.Session, bool)
	Delete(ctx context.Context, sessionID uuid.UUID)
}
