package auth

import (
	"context"
	"errors"
	"fmt"
	"unicode"

	"github.com/google/uuid"
	"github.com/netbill/auth-svc/internal/core/errx"
	"github.com/netbill/auth-svc/internal/core/models"
	"github.com/netbill/restkit/pagi"
	"github.com/netbill/restkit/tokens"
)

type Module struct {
	repo      repo
	token     tokenManager
	messenger messenger
	password  PasswordManager
}

func New(
	db repo,
	jwt tokenManager,
	event messenger,
	passworeder PasswordManager,
) *Module {
	return &Module{
		repo:      db,
		token:     jwt,
		messenger: event,
		password:  passworeder,
	}
}

type repo interface {
	CreateAccount(
		ctx context.Context,
		params RegistrationParams,
	) (models.Account, error)

	GetAccountByID(ctx context.Context, accountID uuid.UUID) (models.Account, error)
	GetAccountByEmail(ctx context.Context, email string) (models.Account, error)
	GetAccountByUsername(ctx context.Context, username string) (models.Account, error)

	ExistsAccountByID(ctx context.Context, accountID uuid.UUID) (bool, error)
	ExistsAccountByEmail(ctx context.Context, email string) (bool, error)
	ExistsAccountByUsername(ctx context.Context, username string) (bool, error)

	GetAccountEmail(ctx context.Context, accountID uuid.UUID) (models.AccountEmail, error)
	GetAccountPassword(ctx context.Context, accountID uuid.UUID) (models.AccountPassword, error)

	UpdateAccountPassword(
		ctx context.Context,
		accountID uuid.UUID,
		passwordHash string,
	) (models.AccountPassword, error)
	UpdateAccountUsername(
		ctx context.Context,
		accountID uuid.UUID,
		newUsername string,
	) (models.Account, error)

	DeleteAccount(ctx context.Context, accountID uuid.UUID) error

	CreateSession(ctx context.Context, sessionID, accountID uuid.UUID, hashToken string) (models.Session, error)
	GetSession(ctx context.Context, sessionID uuid.UUID) (models.Session, error)
	GetAccountSession(
		ctx context.Context,
		accountID, sessionID uuid.UUID,
	) (models.Session, error)
	GetSessionsForAccount(
		ctx context.Context,
		accountID uuid.UUID,
		limit, offset uint,
	) (pagi.Page[[]models.Session], error)
	GetSessionToken(ctx context.Context, sessionID uuid.UUID) (string, error)
	UpdateSessionToken(
		ctx context.Context,
		sessionID uuid.UUID,
		token string,
	) (models.Session, error)

	DeleteSession(ctx context.Context, sessionID uuid.UUID) error
	DeleteSessionsForAccount(ctx context.Context, accountID uuid.UUID) error
	DeleteAccountSession(ctx context.Context, accountID, sessionID uuid.UUID) error

	ExistOrgMemberByAccount(ctx context.Context, accountID uuid.UUID) (bool, error)

	BuryAccount(ctx context.Context, accountID uuid.UUID) error
	BurySession(ctx context.Context, sessionID uuid.UUID) error
	BuryAccountSessions(ctx context.Context, accountID uuid.UUID) error

	AccountIsBuried(ctx context.Context, accountID uuid.UUID) (bool, error)
	SessionIsBuried(ctx context.Context, sessionID uuid.UUID) (bool, error)

	Transaction(ctx context.Context, fn func(ctx context.Context) error) error
}

type tokenManager interface {
	ParseAccountAuthAccess(tokenStr string) (tokens.AccountAuthClaims, error)
	ParseAccountAuthRefresh(enc string) (tokens.AccountAuthClaims, error)

	HashRefresh(rawRefresh string) (string, error)

	GenerateAccess(
		account models.Account, sessionID uuid.UUID,
	) (string, error)

	GenerateRefresh(
		account models.Account, sessionID uuid.UUID,
	) (string, error)
}

type messenger interface {
	WriteAccountCreated(ctx context.Context, account models.Account) error
	WriteAccountUsernameUpdated(ctx context.Context, account models.Account) error
	WriteAccountDeleted(ctx context.Context, accountID uuid.UUID) error
}

type PasswordManager interface {
	CheckRequirements(password string) error
	CheckPasswordMatch(hash, password string) error
	GenerateHash(password string) (string, error)
}

func (m *Module) checkUsernameRequirements(ctx context.Context, username string) error {
	exist, err := m.repo.ExistsAccountByUsername(ctx, username)
	if err != nil {
		return err
	}
	if exist {
		return errx.ErrorUsernameAlreadyTaken.Raise(
			fmt.Errorf("username '%s' is already taken", username),
		)
	}

	if len(username) < 3 || len(username) > 32 {
		return errx.ErrorUsernameIsNotAllowed.Raise(
			fmt.Errorf("username must be between 3 and 32 characters"),
		)
	}

	for _, r := range username {
		if !(unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' || r == '-') {
			return errx.ErrorUsernameIsNotAllowed.Raise(
				fmt.Errorf("username contains invalid characters %s", string(r)),
			)
		}
	}

	return nil
}

func (m *Module) validateActorSession(
	ctx context.Context,
	actor models.AccountActor,
) (models.Account, models.Session, error) {
	account, err := m.repo.GetAccountByID(ctx, actor.ID)
	if errors.Is(err, errx.ErrorAccountNotFound) {
		buried, err := m.repo.AccountIsBuried(ctx, actor.ID)
		if err != nil {
			return models.Account{}, models.Session{}, err
		}
		if buried {
			return models.Account{}, models.Session{}, errx.ErrorAccountInvalidSession.Raise(
				fmt.Errorf("account with id %s is buried", actor.ID),
			)
		}

		return models.Account{}, models.Session{}, errx.ErrorAccountInvalidSession.Raise(
			fmt.Errorf("account with id %s not found", actor.ID),
		)
	}
	if err != nil {
		return models.Account{}, models.Session{}, err
	}

	session, err := m.repo.GetSession(ctx, actor.SessionID)
	if errors.Is(err, errx.ErrorSessionNotFound) {
		buried, err := m.repo.SessionIsBuried(ctx, actor.SessionID)
		if err != nil {
			return models.Account{}, models.Session{}, err
		}
		if buried {
			return models.Account{}, models.Session{}, errx.ErrorAccountInvalidSession.Raise(
				fmt.Errorf("session with id %s is buried", actor.SessionID),
			)
		}

		return models.Account{}, models.Session{}, errx.ErrorAccountInvalidSession.Raise(
			fmt.Errorf("session with id %s not found", actor.SessionID),
		)
	}
	if err != nil {
		return models.Account{}, models.Session{}, err
	}

	return account, session, nil
}
