package account

import (
	"context"

	"github.com/google/uuid"
	"github.com/netbill/auth-svc/internal/models"
)

//go:generate mockery --name=accountCache --inpackage
type accountCache interface {
	Set(ctx context.Context, account models.Account) error
	Get(ctx context.Context, accountID uuid.UUID) (models.Account, error)
	Delete(ctx context.Context, accountID uuid.UUID) error
}

//go:generate mockery --name=emailCache --inpackage
type emailCache interface {
	Set(ctx context.Context, email models.AccountEmail) error

	GetByID(ctx context.Context, accountID uuid.UUID) (models.AccountEmail, error)
	GetByEmail(ctx context.Context, email string) (models.AccountEmail, error)

	DeleteByID(ctx context.Context, accountID uuid.UUID) error
	DeleteByEmail(ctx context.Context, email string) error
}

//go:generate mockery --name=passwordCache --inpackage
type passwordCache interface {
	Set(ctx context.Context, password models.AccountPassword) error
	Get(ctx context.Context, accountID uuid.UUID) (models.AccountPassword, error)
	Delete(ctx context.Context, accountID uuid.UUID) error
}

//go:generate mockery --name=sessionsCache --inpackage
type sessionsCache interface {
	Delete(ctx context.Context, sessionID uuid.UUID) error
}
