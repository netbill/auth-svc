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

//go:generate mockery --name=userRepo --inpackage
type userRepo interface {
	GetByID(ctx context.Context, userID uuid.UUID) (models.User, error)
}

//go:generate mockery --name=emailRepo --inpackage
type emailRepo interface {
	GetByEmail(ctx context.Context, email string) (models.UserEmail, error)
}

//go:generate mockery --name=passwordRepo --inpackage
type passwordRepo interface {
	GetByID(ctx context.Context, userID uuid.UUID) (models.UserPassword, error)
}

//go:generate mockery --name=sessionRepo --inpackage
type sessionRepo interface {
	Create(
		ctx context.Context,
		sessionID, userID uuid.UUID,
		hashToken string,
	) (models.Session, error)

	GetByID(ctx context.Context, sessionID uuid.UUID) (models.Session, error)
	GetForUser(ctx context.Context, userID, sessionID uuid.UUID) (models.Session, error)
	GetListForUser(
		ctx context.Context,
		userID uuid.UUID,
		opts ...ListSessionsOption,
	) (pagi.Page[[]models.Session], error)

	GetToken(ctx context.Context, sessionID uuid.UUID) (string, error)

	UpdateToken(ctx context.Context, sessionID uuid.UUID, token string) (models.Session, error)

	Delete(ctx context.Context, sessionID uuid.UUID) error
	DeleteOneForUser(ctx context.Context, userID, sessionID uuid.UUID) error
	DeleteManyForUser(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error)
}
