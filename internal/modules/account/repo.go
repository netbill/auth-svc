package account

import (
	"context"

	"github.com/google/uuid"
	"github.com/netbill/auth-svc/internal/models"
)

//go:generate mockery --name=transaction --inpackage
type transaction interface {
	Transaction(ctx context.Context, fn func(ctx context.Context) error) error
}

//go:generate mockery --name=accountRepo --inpackage
type accountRepo interface {
	Create(ctx context.Context, params RegistrationParams) (models.Account, error)
	GetByID(ctx context.Context, accountID uuid.UUID) (models.Account, error)
	UpdateUsername(
		ctx context.Context,
		accountID uuid.UUID,
		newUsername string,
		version int32,
	) (models.Account, error)
	Delete(ctx context.Context, accountID uuid.UUID) (models.Account, error)
}

//go:generate mockery --name=emailRepo --inpackage
type emailRepo interface {
	Create(ctx context.Context, params models.AccountEmail) (models.AccountEmail, error)
	GetByID(ctx context.Context, accountID uuid.UUID, opts ...GetAccountOption) (models.AccountEmail, error)
}

//go:generate mockery --name=sessionRepo --inpackage
type sessionRepo interface {
	DeleteManyForAccount(ctx context.Context, accountID uuid.UUID) ([]uuid.UUID, error)
}

//go:generate mockery --name=passwordRepo --inpackage
type passwordRepo interface {
	Create(ctx context.Context, params models.AccountPassword) (models.AccountPassword, error)
	GetByID(ctx context.Context, accountID uuid.UUID) (models.AccountPassword, error)
	UpdatePassword(
		ctx context.Context,
		accountID uuid.UUID,
		passwordHash string,
	) (models.AccountPassword, error)
}
