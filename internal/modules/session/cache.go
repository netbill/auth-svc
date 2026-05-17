package session

import (
	"context"

	"github.com/google/uuid"
	"github.com/netbill/auth-svc/internal/models"
)

//go:generate mockery --name=passwordCache --inpackage
type passwordCache interface {
	Set(ctx context.Context, password models.AccountPassword) error
	Get(ctx context.Context, accountID uuid.UUID) (models.AccountPassword, error)
}

//go:generate mockery --name=accountCache --inpackage
type accountCache interface {
	Set(ctx context.Context, account models.Account) error
	Get(ctx context.Context, accountID uuid.UUID) (models.Account, error)
}

//go:generate mockery --name=sessionsCache --inpackage
type sessionsCache interface {
	Set(ctx context.Context, session models.Session) error
	Get(ctx context.Context, sessionID uuid.UUID) (models.Session, error)
	Delete(ctx context.Context, sessionID uuid.UUID) error
}
