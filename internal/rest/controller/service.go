package controller

import (
	"context"

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
	RegistrationByAdmin(
		ctx context.Context,
		initiatorID uuid.UUID,
		params account.RegistrationParams,
	) (models.Account, error)

	LoginByEmail(ctx context.Context, email, password string) (models.TokensPair, error)
	LoginByGoogle(ctx context.Context, email string) (models.TokensPair, error)

	Refresh(ctx context.Context, oldRefreshToken string) (models.TokensPair, error)

	UpdatePassword(
		ctx context.Context,
		initiator account.InitiatorData,
		oldPassword, newPassword string,
	) error

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

type Service struct {
	google oauth2.Config
	core   core
	log    logium.Logger
}

func New(log logium.Logger, google oauth2.Config, domain core) *Service {
	return &Service{
		log:    log,
		google: google,
		core:   domain,
	}
}
