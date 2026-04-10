package account

import (
	"context"

	"github.com/google/uuid"
	"github.com/netbill/auth-svc/internal/models"
	"github.com/netbill/restkit/pagi"
)

type transaction interface {
	Transaction(ctx context.Context, fn func(ctx context.Context) error) error
}

type accountRepo interface {
	Create(ctx context.Context, params RegistrationParams) (models.Account, error)

	GetByID(ctx context.Context, accountID uuid.UUID) (models.Account, error)
	GetByEmail(ctx context.Context, email string) (models.Account, error)
	GetByUsername(ctx context.Context, username string) (models.Account, error)

	ExistsByID(ctx context.Context, accountID uuid.UUID) (bool, error)
	ExistsByUsername(ctx context.Context, username string) (bool, error)

	UpdateUsername(
		ctx context.Context,
		accountID uuid.UUID,
		newUsername string,
		version int32,
	) (models.Account, error)

	Delete(ctx context.Context, accountID uuid.UUID) error
}

type emailRepo interface {
	Create(ctx context.Context, params models.AccountEmail) error

	GetByID(ctx context.Context, accountID uuid.UUID) (models.AccountEmail, error)

	ExistsByEmail(ctx context.Context, email string) (bool, error)
}

type passwordRepo interface {
	Create(ctx context.Context, params models.AccountPassword) error

	GetByID(ctx context.Context, accountID uuid.UUID) (models.AccountPassword, error)

	UpdatePassword(
		ctx context.Context,
		accountID uuid.UUID,
		passwordHash string,
	) (models.AccountPassword, error)
}

type sessionRepo interface {
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
	DeleteOneForAccount(ctx context.Context, accountID, sessionID uuid.UUID) error
	DeleteManyForAccount(ctx context.Context, accountID uuid.UUID) error
}

type accountCache interface {
	SetByID(ctx context.Context, account models.Account) error
	SetByUsername(ctx context.Context, account models.Account) error
	SetByEmail(ctx context.Context, email string, account models.Account) error

	GetByID(ctx context.Context, accountID uuid.UUID) (models.Account, error)
	GetByUsername(ctx context.Context, username string) (models.Account, error)
	GetByEmail(ctx context.Context, email string) (models.Account, error)

	Delete(ctx context.Context, accountID uuid.UUID) error
	DeleteByUsername(ctx context.Context, username string) error
	DeleteByEmail(ctx context.Context, email string) error
}

type emailCache interface {
	SetByID(ctx context.Context, account models.AccountEmail) error
	SetByEmail(ctx context.Context, account models.AccountEmail) error
	SetByUsername(ctx context.Context, username string, account models.AccountEmail) error

	GetByID(ctx context.Context, accountID uuid.UUID) (models.AccountEmail, error)
	GetByEmail(ctx context.Context, email string) (models.AccountEmail, error)
	GetByUsername(ctx context.Context, username string) (models.AccountEmail, error)

	DeleteByID(ctx context.Context, accountID uuid.UUID) error
	DeleteByUsername(ctx context.Context, username string) error
	DeleteByEmail(ctx context.Context, email string) error
}

type passwordCache interface {
	SetByID(ctx context.Context, account models.AccountPassword) error
	GetByID(ctx context.Context, accountID uuid.UUID) (models.AccountPassword, error)
	DeleteByID(ctx context.Context, accountID uuid.UUID) error
}

type sessionsCache interface {
	Set(ctx context.Context, session models.Session) error
	GetByID(ctx context.Context, sessionID uuid.UUID) (models.Session, error)
	DeleteByID(ctx context.Context, sessionID uuid.UUID) error
}
