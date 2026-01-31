package controller

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/netbill/auth-svc/internal/core/models"
	"github.com/netbill/auth-svc/internal/core/modules/account"
	"github.com/netbill/logium"
	"github.com/netbill/restkit/pagi"
	"golang.org/x/oauth2"
)

type core interface {
	Registration(
		ctx context.Context,
		params account.RegistrationParams,
	) (models.Account, error)

	LoginByEmail(ctx context.Context, email, password string) (models.TokensPair, error)
	LoginByGoogle(ctx context.Context, email string) (models.TokensPair, error)
	LoginByUsername(ctx context.Context, username, password string) (models.TokensPair, error)

	Refresh(ctx context.Context, oldRefreshToken string) (models.TokensPair, error)

	UpdatePassword(
		ctx context.Context,
		initiator account.InitiatorData,
		oldPassword, newPassword string,
	) error
	UpdateUsername(
		ctx context.Context,
		initiator account.InitiatorData,
		newUsername string,
	) (account models.Account, err error)

	GetAccountByID(ctx context.Context, ID uuid.UUID) (models.Account, error)
	GetAccountEmail(ctx context.Context, ID uuid.UUID) (models.AccountEmail, error)

	GetOwnSession(ctx context.Context, initiator account.InitiatorData, sessionID uuid.UUID) (models.Session, error)
	GetOwnSessions(
		ctx context.Context,
		initiator account.InitiatorData,
		limit, offset uint,
	) (pagi.Page[[]models.Session], error)

	DeleteOwnAccount(ctx context.Context, initiator account.InitiatorData) error

	Logout(ctx context.Context, initiator account.InitiatorData) error
	DeleteOwnSession(ctx context.Context, initiator account.InitiatorData, sessionID uuid.UUID) error
	DeleteOwnSessions(ctx context.Context, initiator account.InitiatorData) error
}

type responser interface {
	Render(w http.ResponseWriter, status int, res ...interface{})
	RenderErr(w http.ResponseWriter, errs ...error)
}

type Controller struct {
	log    *logium.Logger
	google oauth2.Config

	core      core
	responser responser
}

func New(log *logium.Logger, google oauth2.Config, core core, responser responser) *Controller {
	return &Controller{
		log:       log,
		google:    google,
		core:      core,
		responser: responser,
	}
}
