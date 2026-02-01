package account

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"unicode"

	"github.com/google/uuid"
	"github.com/netbill/auth-svc/internal/core/errx"
	"github.com/netbill/auth-svc/internal/core/models"
	"github.com/netbill/restkit/pagi"
	"github.com/netbill/restkit/tokens"
)

type Module struct {
	repo      repo
	jwt       JWTManager
	messenger messenger
}

func NewService(
	db repo,
	jwt JWTManager,
	event messenger,
) *Module {
	return &Module{
		repo:      db,
		jwt:       jwt,
		messenger: event,
	}
}

type JWTManager interface {
	ParseAccessClaims(tokenStr string) (tokens.AccountClaims, error)
	ParseRefreshClaims(enc string) (tokens.AccountClaims, error)

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

type CreateAccountParams struct {
	Role         string
	Email        string
	Username     string
	PasswordHash string
}

type repo interface {
	CreateAccount(
		ctx context.Context,
		params CreateAccountParams,
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

	Transaction(ctx context.Context, fn func(ctx context.Context) error) error
}

func (m *Module) checkUsernameRequirements(ctx context.Context, username string) error {
	_, err := m.repo.GetAccountByUsername(ctx, username)
	if err != nil {
		return err
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

func (m *Module) checkPasswordRequirements(password string) error {
	if len(password) < 8 || len(password) > 32 {
		return errx.ErrorPasswordIsNotAllowed.Raise(
			fmt.Errorf("password must be between 8 and 32 characters"),
		)
	}

	var (
		hasUpper, hasLower, hasDigit, hasSpecial bool
	)

	allowedSpecials := "-.!#$%&?,@"

	for _, r := range password {
		switch {
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsLower(r):
			hasLower = true
		case unicode.IsDigit(r):
			hasDigit = true
		case strings.ContainsRune(allowedSpecials, r):
			hasSpecial = true
		default:
			return errx.ErrorPasswordIsNotAllowed.Raise(
				fmt.Errorf("password contains invalid characters %s", string(r)),
			)
		}
	}

	if !hasUpper {
		return errx.ErrorPasswordIsNotAllowed.Raise(
			fmt.Errorf("need at least one uppercase letter"),
		)
	}
	if !hasLower {
		return errx.ErrorPasswordIsNotAllowed.Raise(
			fmt.Errorf("need at least one lower case letter"),
		)
	}
	if !hasDigit {
		return errx.ErrorPasswordIsNotAllowed.Raise(
			fmt.Errorf("need at least one digit"),
		)
	}
	if !hasSpecial {
		return errx.ErrorPasswordIsNotAllowed.Raise(
			fmt.Errorf("need at least one special character from %s", allowedSpecials),
		)
	}

	return nil
}

type InitiatorData struct {
	AccountID uuid.UUID
	SessionID uuid.UUID
}

func (m *Module) validateInitiatorSession(
	ctx context.Context,
	initiator InitiatorData,
) (models.Account, models.Session, error) {
	account, err := m.repo.GetAccountByID(ctx, initiator.AccountID)
	switch {
	case errors.Is(err, errx.ErrorAccountNotFound):
		return models.Account{}, models.Session{}, errx.ErrorInitiatorNotFound.Raise(
			fmt.Errorf("account with id '%s' not found", initiator.SessionID),
		)
	case err != nil:
		return models.Account{}, models.Session{}, errx.ErrorInitiatorNotFound.Raise(
			fmt.Errorf("failed to get account with id '%s', cause: %w", initiator.SessionID, err),
		)
	}

	session, err := m.repo.GetSession(ctx, initiator.SessionID)
	switch {
	case errors.Is(err, errx.ErrorSessionNotFound):
		return models.Account{}, models.Session{}, errx.ErrorInitiatorInvalidSession.Raise(
			fmt.Errorf("failed to get session with id '%s', cause: %w", initiator.SessionID, err),
		)
	case session.AccountID != initiator.AccountID:
		return models.Account{}, models.Session{}, errx.ErrorInitiatorInvalidSession.Raise(
			fmt.Errorf("session with id '%s' not found for account '%s'", initiator.SessionID, initiator.AccountID),
		)
	}

	return account, session, nil
}
