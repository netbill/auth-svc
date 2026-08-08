package session

import (
	"context"

	"github.com/google/uuid"
	"github.com/netbill/auth-svc/internal/models"
	"github.com/netbill/restkit/pagi"
)

//go:generate mockery --name=transaction --inpackage
type transaction interface {
	Transaction(ctx context.Context, fn func(ctx context.Context) error) error
}

//go:generate mockery --name=accountRepo --inpackage
type accountRepo interface {
	GetByID(ctx context.Context, accountID uuid.UUID) (models.Account, error)
}

//go:generate mockery --name=emailRepo --inpackage
type emailRepo interface {
	GetByEmail(ctx context.Context, email string) (models.AccountEmail, error)
}

//go:generate mockery --name=passwordRepo --inpackage
type passwordRepo interface {
	GetByID(ctx context.Context, accountID uuid.UUID) (models.AccountPassword, error)
}

//go:generate mockery --name=sessionRepo --inpackage
type sessionRepo interface {
	Create(
		ctx context.Context,
		sessionID, accountID uuid.UUID,
		hashToken string,
	) (models.Session, error)

	GetByID(ctx context.Context, sessionID uuid.UUID) (models.Session, error)
	GetForAccount(ctx context.Context, accountID, sessionID uuid.UUID) (models.Session, error)
	GetListForAccount(
		ctx context.Context,
		accountID uuid.UUID,
		opts ...ListSessionsOption,
	) (pagi.Page[[]models.Session], error)

	GetToken(ctx context.Context, sessionID uuid.UUID) (string, error)

	UpdateToken(ctx context.Context, sessionID uuid.UUID, token string) (models.Session, error)

	Delete(ctx context.Context, sessionID uuid.UUID) error
	DeleteOneForAccount(ctx context.Context, accountID, sessionID uuid.UUID) error
	DeleteManyForAccount(ctx context.Context, accountID uuid.UUID) ([]uuid.UUID, error)
}
