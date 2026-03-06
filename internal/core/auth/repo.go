package auth

import (
	"context"

	"github.com/google/uuid"
	"github.com/netbill/auth-svc/internal/models"
	"github.com/netbill/restkit/pagi"
)

type accountRepo interface {
	Create(
		ctx context.Context,
		params RegistrationParams,
	) (models.Account, error)

	GetByID(ctx context.Context, accountID uuid.UUID) (models.Account, error)
	GetByEmail(ctx context.Context, email string) (models.Account, error)
	GetByUsername(ctx context.Context, username string) (models.Account, error)

	ExistsByID(ctx context.Context, accountID uuid.UUID) (bool, error)
	ExistsByUsername(ctx context.Context, username string) (bool, error)

	UpdateUsername(
		ctx context.Context,
		accountID uuid.UUID,
		newUsername string,
	) (models.Account, error)

	Delete(ctx context.Context, accountID uuid.UUID) error
}

type accountEmailRepo interface {
	Create(ctx context.Context, params models.AccountEmail) error

	GetByID(ctx context.Context, accountID uuid.UUID) (models.AccountEmail, error)

	ExistsByEmail(ctx context.Context, email string) (bool, error)
}

type accountPasswordRepo interface {
	Create(ctx context.Context, params models.AccountPassword) error

	GetByID(ctx context.Context, accountID uuid.UUID) (models.AccountPassword, error)

	UpdatePassword(
		ctx context.Context,
		accountID uuid.UUID,
		passwordHash string,
	) (models.AccountPassword, error)
}

type accountSessionRepo interface {
	Create(
		ctx context.Context,
		sessionID, accountID uuid.UUID,
		hashToken string,
	) (models.Session, error)

	GetByID(ctx context.Context, sessionID uuid.UUID) (models.Session, error)
	GetForAccount(
		ctx context.Context,
		accountID, sessionID uuid.UUID,
	) (models.Session, error)
	GetListForAccount(
		ctx context.Context,
		accountID uuid.UUID,
		limit, offset uint,
	) (pagi.Page[[]models.Session], error)
	GetToken(ctx context.Context, sessionID uuid.UUID) (string, error)
	UpdateToken(
		ctx context.Context,
		sessionID uuid.UUID,
		token string,
	) (models.Session, error)

	Delete(ctx context.Context, sessionID uuid.UUID) error
	DeleteManyForAccount(ctx context.Context, accountID uuid.UUID) error
	DeleteOneForAccount(ctx context.Context, accountID, sessionID uuid.UUID) error
}

type accountOrgRepo interface {
	ExistMemberByAccount(ctx context.Context, accountID uuid.UUID) (bool, error)
}

type accountTombstoneRepo interface {
	BuryAccount(ctx context.Context, accountID uuid.UUID) error
	BurySession(ctx context.Context, sessionID uuid.UUID) error
	BuryAccountSessions(ctx context.Context, accountID uuid.UUID) error

	AccountIsBuried(ctx context.Context, accountID uuid.UUID) (bool, error)
	SessionIsBuried(ctx context.Context, sessionID uuid.UUID) (bool, error)
}

type transaction interface {
	Transaction(ctx context.Context, fn func(ctx context.Context) error) error
}
