package account

import (
	"context"
	"fmt"
	"strings"
	"unicode"

	"github.com/google/uuid"
	"github.com/netbill/auth-svc/internal/core/errx"
	"github.com/netbill/auth-svc/internal/core/models"
	"github.com/netbill/restkit/auth"
	"github.com/netbill/restkit/pagi"
)

type Service struct {
	repo      repo
	jwt       JWTManager
	messenger messenger
}

func NewService(
	db repo,
	jwt JWTManager,
	event messenger,
) *Service {
	return &Service{
		repo:      db,
		jwt:       jwt,
		messenger: event,
	}
}

type JWTManager interface {
	EncryptAccess(token string) (string, error)
	EncryptRefresh(token string) (string, error)
	DecryptRefresh(encryptedToken string) (string, error)

	ParseRefreshClaims(enc string) (auth.AccountClaims, error)

	GenerateAccess(
		account models.Account, sessionID uuid.UUID,
	) (string, error)

	GenerateRefresh(
		account models.Account, sessionID uuid.UUID,
	) (string, error)
}

type messenger interface {
	WriteAccountCreated(ctx context.Context, account models.Account, email string) error
	WriteAccountPasswordChanged(ctx context.Context, account models.Account) error
	WriteAccountUsernameChanged(ctx context.Context, account models.Account) error
	WriteAccountLogin(ctx context.Context, account models.Account) error
	WriteAccountDeleted(ctx context.Context, account models.Account) error
}

type CreateAccountParams struct {
	Username     string
	Role         string
	Email        string
	PasswordHash string
}

type repo interface {
	CreateAccount(
		ctx context.Context,
		params CreateAccountParams,
	) (models.Account, error)

	GetAccountByID(ctx context.Context, accountID uuid.UUID) (models.Account, error)
	GetAccountByUsername(ctx context.Context, username string) (models.Account, error)
	UpdateAccountUsername(
		ctx context.Context,
		accountID uuid.UUID,
		newUsername string,
	) (models.Account, error)

	GetAccountByEmail(ctx context.Context, email string) (models.Account, error)
	UpdateAccountStatus(
		ctx context.Context,
		accountID uuid.UUID,
		status string,
	) (models.Account, error)

	GetAccountEmail(ctx context.Context, accountID uuid.UUID) (models.AccountEmail, error)
	UpdateAccountEmailVerification(
		ctx context.Context,
		accountID uuid.UUID,
		verified bool,
	) (models.AccountEmail, error)

	GetAccountPassword(ctx context.Context, accountID uuid.UUID) (models.AccountPassword, error)
	UpdateAccountPassword(
		ctx context.Context,
		accountID uuid.UUID,
		passwordHash string,
	) (models.AccountPassword, error)
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

	Transaction(ctx context.Context, fn func(ctx context.Context) error) error
}

func (s Service) CheckPasswordRequirements(password string) error {
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

func (s Service) CheckUsernameRequirements(username string) error {
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

type InitiatorData struct {
	AccountID uuid.UUID
	SessionID uuid.UUID
}
