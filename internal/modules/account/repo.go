package account

import (
	"context"

	"github.com/google/uuid"
	"github.com/netbill/auth-svc/internal/models"
)

type transaction interface {
	Transaction(ctx context.Context, fn func(ctx context.Context) error) error
}

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

type DeletedFilter uint8

const (
	DeletedFilterActive  DeletedFilter = iota // 0 — default: only active (deleted_at IS NULL)
	DeletedFilterAll                          // active + deleted (no filter)
	DeletedFilterDeleted                      // only deleted (deleted_at IS NOT NULL)
)

type GetAccountOptions struct {
	Deleted DeletedFilter
}

type GetAccountOption func(*GetAccountOptions)

func ApplyGetAccountOptions(optFns []GetAccountOption) GetAccountOptions {
	var opts GetAccountOptions
	for _, fn := range optFns {
		fn(&opts)
	}
	return opts
}

func WithDeleted(f DeletedFilter) GetAccountOption {
	return func(o *GetAccountOptions) {
		o.Deleted = f
	}
}

type emailRepo interface {
	Create(ctx context.Context, params models.AccountEmail) (models.AccountEmail, error)

	GetByID(ctx context.Context, accountID uuid.UUID, opts ...GetAccountOption) (models.AccountEmail, error)
}

type sessionRepo interface {
	DeleteManyForAccount(ctx context.Context, accountID uuid.UUID) ([]uuid.UUID, error)
}

type passwordRepo interface {
	Create(ctx context.Context, params models.AccountPassword) (models.AccountPassword, error)

	GetByID(ctx context.Context, accountID uuid.UUID) (models.AccountPassword, error)

	UpdatePassword(
		ctx context.Context,
		accountID uuid.UUID,
		passwordHash string,
	) (models.AccountPassword, error)
}

type accountCache interface {
	Set(ctx context.Context, account models.Account) error
	GetByID(ctx context.Context, accountID uuid.UUID) (models.Account, error)
	Delete(ctx context.Context, accountID uuid.UUID) error
}

type emailCache interface {
	Set(ctx context.Context, account models.AccountEmail) error

	GetByID(ctx context.Context, accountID uuid.UUID) (models.AccountEmail, error)
	GetByEmail(ctx context.Context, email string) (models.AccountEmail, error)

	DeleteByID(ctx context.Context, accountID uuid.UUID) error
	DeleteByEmail(ctx context.Context, email string) error
}

type passwordCache interface {
	SetByID(ctx context.Context, account models.AccountPassword) error
	GetByID(ctx context.Context, accountID uuid.UUID) (models.AccountPassword, error)
	DeleteByID(ctx context.Context, accountID uuid.UUID) error
}

type sessionsCache interface {
	DeleteByID(ctx context.Context, sessionID uuid.UUID) error
}
