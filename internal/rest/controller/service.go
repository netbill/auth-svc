package controller

import (
	"context"

	"github.com/google/uuid"
	"github.com/umisto/logium"
	"github.com/umisto/pagi"
	"github.com/umisto/sso-svc/internal/domain/models"
	"github.com/umisto/sso-svc/internal/domain/modules/auth"
	"golang.org/x/oauth2"
)

type core interface {
	Registration(
		ctx context.Context,
		params auth.RegistrationParams,
	) (models.Account, error)
	RegistrationByAdmin(
		ctx context.Context,
		initiatorID uuid.UUID,
		params auth.RegistrationParams,
	) (models.Account, error)

	LoginByEmail(ctx context.Context, email, password string) (models.TokensPair, error)
	LoginByUsername(ctx context.Context, username, password string) (models.TokensPair, error)
	LoginByGoogle(ctx context.Context, email string) (models.TokensPair, error)

	Refresh(ctx context.Context, oldRefreshToken string) (models.TokensPair, error)

	UpdatePassword(
		ctx context.Context,
		initiator auth.InitiatorData,
		oldPassword, newPassword string,
	) error
	UpdateUsername(
		ctx context.Context,
		initiator auth.InitiatorData,
		password string,
		newUsername string,
	) (models.Account, error)

	GetAccountByID(ctx context.Context, ID uuid.UUID) (models.Account, error)
	GetAccountEmail(ctx context.Context, ID uuid.UUID) (models.AccountEmail, error)

	GetOwnSession(ctx context.Context, initiator auth.InitiatorData, sessionID uuid.UUID) (models.Session, error)
	GetOwnSessions(
		ctx context.Context,
		initiator auth.InitiatorData,
		limit, offset uint,
	) (pagi.Page[[]models.Session], error)

	DeleteOwnAccount(ctx context.Context, initiator auth.InitiatorData) error

	Logout(ctx context.Context, initiator auth.InitiatorData) error
	DeleteOwnSession(ctx context.Context, initiator auth.InitiatorData, sessionID uuid.UUID) error
	DeleteOwnSessions(ctx context.Context, initiator auth.InitiatorData) error
}

type Service struct {
	google oauth2.Config
	domain core
	log    logium.Logger
}

func New(log logium.Logger, google oauth2.Config, domain core) *Service {
	return &Service{
		log:    log,
		google: google,
		domain: domain,
	}
}
