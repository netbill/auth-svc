package session

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
	GetByID(ctx context.Context, accountID uuid.UUID) (models.Account, error)
	GetByUsername(ctx context.Context, username string) (models.Account, error)
}

type emailRepo interface {
	GetByEmail(ctx context.Context, email string) (models.AccountEmail, error)
}

type passwordRepo interface {
	GetByID(ctx context.Context, accountID uuid.UUID) (models.AccountPassword, error)
}

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
		limit, offset uint,
	) (pagi.Page[[]models.Session], error)

	GetToken(ctx context.Context, sessionID uuid.UUID) (string, error)

	UpdateToken(ctx context.Context, sessionID uuid.UUID, token string) (models.Session, error)

	Delete(ctx context.Context, sessionID uuid.UUID) error
	DeleteOneForAccount(ctx context.Context, accountID, sessionID uuid.UUID) error
	DeleteManyForAccount(ctx context.Context, accountID uuid.UUID) ([]uuid.UUID, error)
}

type auth interface {
	ValidateSession(ctx context.Context, actor models.AccountActor) (models.Account, models.Session, error)
}

type passwordCache interface {
	SetByID(ctx context.Context, password models.AccountPassword) error
	GetByID(ctx context.Context, accountID uuid.UUID) (models.AccountPassword, error)
}

type accountCache interface {
	Set(ctx context.Context, account models.Account) error
	GetByID(ctx context.Context, accountID uuid.UUID) (models.Account, error)
}

type sessionsCache interface {
	Set(ctx context.Context, session models.Session) error
	GetByID(ctx context.Context, sessionID uuid.UUID) (models.Session, error)
	DeleteByID(ctx context.Context, sessionID uuid.UUID) error
}
