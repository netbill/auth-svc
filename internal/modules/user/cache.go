package user

import (
	"context"

	"github.com/google/uuid"
	"github.com/netbill/auth-svc/internal/models"
)

//go:generate mockery --name=userCache --inpackage
type userCache interface {
	Set(ctx context.Context, user models.User) error
	Get(ctx context.Context, userID uuid.UUID) (models.User, error)
	Delete(ctx context.Context, userID uuid.UUID) error
}

//go:generate mockery --name=emailCache --inpackage
type emailCache interface {
	Set(ctx context.Context, email models.UserEmail) error

	GetByID(ctx context.Context, userID uuid.UUID) (models.UserEmail, error)
	GetByEmail(ctx context.Context, email string) (models.UserEmail, error)

	DeleteByID(ctx context.Context, userID uuid.UUID) error
	DeleteByEmail(ctx context.Context, email string) error
}

//go:generate mockery --name=passwordCache --inpackage
type passwordCache interface {
	Set(ctx context.Context, password models.UserPassword) error
	Get(ctx context.Context, userID uuid.UUID) (models.UserPassword, error)
	Delete(ctx context.Context, userID uuid.UUID) error
}

//go:generate mockery --name=sessionsCache --inpackage
type sessionsCache interface {
	Delete(ctx context.Context, sessionID uuid.UUID) error
}
