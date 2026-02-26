package controller

import (
	"context"

	"github.com/google/uuid"
	"github.com/netbill/auth-svc/internal/core/models"
	"github.com/netbill/auth-svc/internal/core/modules/auth"
	"github.com/netbill/restkit/pagi"
	"golang.org/x/oauth2"
)

type core interface {
	Registration(
		ctx context.Context,
		params auth.RegistrationParams,
	) (models.Account, error)

	LoginByEmail(ctx context.Context, email, password string) (models.TokensPair, error)
	LoginByGoogle(ctx context.Context, email string) (models.TokensPair, error)
	LoginByUsername(ctx context.Context, username, password string) (models.TokensPair, error)

	Refresh(ctx context.Context, oldRefreshToken string) (models.TokensPair, error)

	UpdatePassword(
		ctx context.Context,
		actor models.AccountActor,
		oldPassword, newPassword string,
	) error
	UpdateUsername(
		ctx context.Context,
		actor models.AccountActor,
		newUsername string,
	) (account models.Account, err error)

	GetMyAccountByID(ctx context.Context, actor models.AccountActor) (models.Account, error)
	GetMyAccountEmail(ctx context.Context, actor models.AccountActor) (models.AccountEmail, error)

	GetMySession(ctx context.Context, actor models.AccountActor, sessionID uuid.UUID) (models.Session, error)
	GetMySessions(
		ctx context.Context,
		actor models.AccountActor,
		limit, offset uint,
	) (pagi.Page[[]models.Session], error)

	DeleteMyAccount(ctx context.Context, actor models.AccountActor) error

	Logout(ctx context.Context, actor models.AccountActor) error
	DeleteMySession(ctx context.Context, actor models.AccountActor, sessionID uuid.UUID) error
	DeleteMySessions(ctx context.Context, actor models.AccountActor) error
}

type Controller struct {
	google oauth2.Config
	core   core
}

func New(core core, google oauth2.Config) *Controller {
	return &Controller{
		google: google,
		core:   core,
	}
}
