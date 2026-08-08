package session

import (
	"context"

	"github.com/google/uuid"
	"github.com/netbill/auth-svc/internal/models"
)

//go:generate mockery --name=passwordCache --inpackage
type passwordCache interface {
	Set(ctx context.Context, password models.UserPassword) error
	Get(ctx context.Context, userID uuid.UUID) (models.UserPassword, error)
}

//go:generate mockery --name=userCache --inpackage
type userCache interface {
	Set(ctx context.Context, user models.User) error
	Get(ctx context.Context, userID uuid.UUID) (models.User, error)
}

//go:generate mockery --name=sessionsCache --inpackage
type sessionsCache interface {
	Set(ctx context.Context, session models.Session) error
	Get(ctx context.Context, sessionID uuid.UUID) (models.Session, error)
	Delete(ctx context.Context, sessionID uuid.UUID) error
}
