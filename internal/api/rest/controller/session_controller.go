package controller

import (
	"context"

	"github.com/google/uuid"
	"github.com/netbill/auth-svc/internal/models"
	"github.com/netbill/auth-svc/internal/modules/session"
	"github.com/netbill/restkit/pagi"
	"golang.org/x/oauth2"
)

type sessionCore interface {
	LoginByEmail(ctx context.Context, email, password string) (models.TokensPair, error)
	LoginByGoogle(ctx context.Context, email string) (models.TokensPair, error)
	LoginByUsername(ctx context.Context, username, password string) (models.TokensPair, error)

	Refresh(ctx context.Context, oldRefreshToken string) (models.TokensPair, error)

	GetMySession(ctx context.Context, actor models.AccountActor, sessionID uuid.UUID) (models.Session, error)
	GetMySessions(
		ctx context.Context,
		actor models.AccountActor,
		opts ...session.ListSessionsOption,
	) (pagi.Page[[]models.Session], error)

	Logout(ctx context.Context, actor models.AccountActor) error
	DeleteMySession(ctx context.Context, actor models.AccountActor, sessionID uuid.UUID) error
	DeleteMySessions(ctx context.Context, actor models.AccountActor) error
}

type SessionController struct {
	google   oauth2.Config
	sessions sessionCore
}

func NewSessionController(sessions sessionCore, google oauth2.Config) *SessionController {
	return &SessionController{
		google:   google,
		sessions: sessions,
	}
}
