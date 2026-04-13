package account

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"
	"unicode"

	"github.com/google/uuid"
	"github.com/netbill/auth-svc/internal/errx"
	"github.com/netbill/auth-svc/internal/models"
	"github.com/netbill/restkit/tokens"
)

type auth interface {
	ValidateSession(ctx context.Context, actor models.AccountActor) (models.Account, models.Session, error)
}

type Service struct {
	auth auth

	accountRepo  accountRepo
	emailRepo    emailRepo
	passwordRepo passwordRepo
	sessionRepo  sessionRepo
	tx           transaction

	accountCache  accountCache
	emailCache    emailCache
	passwordCache passwordCache
	sessionsCache sessionsCache

	passManager passwordManager

	messenger messenger

	log *slog.Logger
}

type ServiceDeps struct {
	Auth          auth
	AccountRepo   accountRepo
	EmailRepo     emailRepo
	PasswordRepo  passwordRepo
	SessionRepo   sessionRepo
	Tx            transaction
	AccountCache  accountCache
	EmailCache    emailCache
	PasswordCache passwordCache
	SessionsCache sessionsCache
	PassManager   passwordManager
	Messenger     messenger
	Log           *slog.Logger
}

func New(deps ServiceDeps) *Service {
	return &Service{
		auth:          deps.Auth,
		accountRepo:   deps.AccountRepo,
		emailRepo:     deps.EmailRepo,
		passwordRepo:  deps.PasswordRepo,
		sessionRepo:   deps.SessionRepo,
		tx:            deps.Tx,
		accountCache:  deps.AccountCache,
		emailCache:    deps.EmailCache,
		passwordCache: deps.PasswordCache,
		sessionsCache: deps.SessionsCache,
		passManager:   deps.PassManager,
		messenger:     deps.Messenger,
		log:           deps.Log,
	}
}

type passwordManager interface {
	CheckMatch(password, hash string) error
	GenerateHash(password string) (string, error)
}

type messenger interface {
	WriteAccountCreated(ctx context.Context, account models.Account, email models.AccountEmail) error

	WriteAccountUsernameUpdated(ctx context.Context, account models.Account) error

	WriteAccountDeleted(ctx context.Context, account models.Account, email models.AccountEmail) error
}

type RegistrationParams struct {
	Email    string
	Username string
	Password string
	Role     string
}

func (s *Service) Registration(
	ctx context.Context,
	params RegistrationParams,
) (models.Account, error) {
	if err := tokens.ValidateUserSystemRole(params.Role); err != nil {
		return models.Account{}, errx.ErrorRoleNotSupported.Raise(err)
	}

	if err := s.checkUsernameRequirements(params.Username); err != nil {
		return models.Account{}, err
	}

	if err := s.checkPasswordRequirements(params.Password); err != nil {
		return models.Account{}, err
	}

	hash, err := s.passManager.GenerateHash(params.Password)
	if err != nil {
		return models.Account{}, err
	}

	var account models.Account
	var email models.AccountEmail
	var password models.AccountPassword

	if err = s.tx.Transaction(ctx, func(ctx context.Context) error {
		account, err = s.accountRepo.Create(ctx, params)
		if err != nil {
			return err
		}

		email, err = s.emailRepo.Create(ctx, models.AccountEmail{
			AccountID: account.ID,
			Email:     params.Email,
		})
		if err != nil {
			return err
		}

		password, err = s.passwordRepo.Create(ctx, models.AccountPassword{
			AccountID: account.ID,
			Hash:      hash,
		})
		if err != nil {
			return err
		}

		return s.messenger.WriteAccountCreated(ctx, account, email)
	}); err != nil {
		return models.Account{}, err
	}

	go func() {
		ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
		defer cancel()

		if err := s.accountCache.Set(ctx, account); err != nil {
			s.log.Error("failed to set account", "error", err)
		}

		if err := s.emailCache.Set(ctx, email); err != nil {
			s.log.Error("failed to set account email", "error", err)
		}

		if err := s.passwordCache.SetByID(ctx, password); err != nil {
			s.log.Error("failed to set account password", "error", err)
		}
	}()

	return account, nil
}

func (s *Service) GetMyAccountByID(
	ctx context.Context,
	actor models.AccountActor,
) (models.Account, error) {
	account, err := s.accountCache.GetByID(ctx, actor.ID)
	switch {
	case errors.Is(err, errx.ErrCacheMiss):
		s.log.Debug("account cache miss", "account_id", actor.ID)
	case err != nil:
		s.log.Error("failed to get account", "error", err)
	default:
		return account, nil
	}

	account, err = s.accountRepo.GetByID(ctx, actor.ID)
	if err != nil {
		return models.Account{}, err
	}

	go func() {
		ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
		defer cancel()

		if err := s.accountCache.Set(ctx, account); err != nil {
			s.log.Error("failed to set account", "error", err)
		}
	}()

	return account, nil
}

func (s *Service) GetMyEmailByID(
	ctx context.Context,
	actor models.AccountActor,
) (models.AccountEmail, error) {
	email, err := s.emailCache.GetByID(ctx, actor.ID)
	switch {
	case errors.Is(err, errx.ErrCacheMiss):
		s.log.Debug("account cache miss", "account_id", actor.ID)
	case err != nil:
		s.log.Error("failed to get account", "error", err)
	default:
		return email, nil
	}

	email, err = s.emailRepo.GetByID(ctx, actor.ID)
	if err != nil {
		return models.AccountEmail{}, err
	}

	go func() {
		ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
		defer cancel()

		if err := s.emailCache.Set(ctx, email); err != nil {
			s.log.Error("failed to set account", "error", err)
		}
	}()

	return email, nil
}

func (s *Service) UpdateUsername(
	ctx context.Context,
	actor models.AccountActor,
	newUsername string,
) (models.Account, error) {
	if err := s.checkUsernameRequirements(newUsername); err != nil {
		return models.Account{}, err
	}

	current, err := s.accountCache.GetByID(ctx, actor.ID)
	switch {
	case errors.Is(err, errx.ErrCacheMiss):
		s.log.Debug("account cache miss", "account_id", actor.ID)
	case err != nil:
		s.log.Error("failed to get account from cache", "error", err)
	}

	if err != nil {
		current, err = s.accountRepo.GetByID(ctx, actor.ID)
		if err != nil {
			return models.Account{}, err
		}
	}

	if current.Username == newUsername {
		return current, nil
	}

	var updated models.Account
	if err = s.tx.Transaction(ctx, func(ctx context.Context) error {
		updated, err = s.accountRepo.UpdateUsername(ctx, actor.ID, newUsername, current.Version)
		if err != nil {
			return err
		}

		return s.messenger.WriteAccountUsernameUpdated(ctx, updated)
	}); err != nil {
		return models.Account{}, err
	}

	go func() {
		ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
		defer cancel()

		if err := s.accountCache.Delete(ctx, actor.ID); err != nil {
			s.log.Error("failed to delete account cache by id", "error", err)
		}

		if err := s.accountCache.Set(ctx, updated); err != nil {
			s.log.Error("failed to set account cache by id", "error", err)
		}
	}()

	return updated, nil
}

func (s *Service) UpdatePassword(
	ctx context.Context,
	actor models.AccountActor,
	oldPassword, newPassword string,
) error {
	if _, _, err := s.auth.ValidateSession(ctx, actor); err != nil {
		return err
	}

	pwd, err := s.passwordCache.GetByID(ctx, actor.ID)
	switch {
	case errors.Is(err, errx.ErrCacheMiss):
		s.log.Debug("password cache miss", "account_id", actor.ID)
	case err != nil:
		s.log.Error("failed to get password from cache", "error", err)
	}

	if err != nil {
		pwd, err = s.passwordRepo.GetByID(ctx, actor.ID)
		if err != nil {
			return err
		}
	}

	if err = s.passManager.CheckMatch(oldPassword, pwd.Hash); err != nil {
		return err
	}

	if err = s.checkPasswordRequirements(newPassword); err != nil {
		return err
	}

	hash, err := s.passManager.GenerateHash(newPassword)
	if err != nil {
		return err
	}

	var updated models.AccountPassword
	if err = s.tx.Transaction(ctx, func(ctx context.Context) error {
		updated, err = s.passwordRepo.UpdatePassword(ctx, actor.ID, hash)
		return err
	}); err != nil {
		return err
	}

	go func() {
		ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
		defer cancel()

		if err := s.passwordCache.DeleteByID(ctx, actor.ID); err != nil {
			s.log.Error("failed to delete password cache", "error", err)
		}
		if err := s.passwordCache.SetByID(ctx, updated); err != nil {
			s.log.Error("failed to set password cache", "error", err)
		}
	}()

	return nil
}

func (s *Service) DeleteMyAccount(
	ctx context.Context,
	actor models.AccountActor,
) error {
	if _, _, err := s.auth.ValidateSession(ctx, actor); err != nil {
		return err
	}

	var (
		account    models.Account
		email      models.AccountEmail
		sessionIDs []uuid.UUID
		err        error
	)

	if err = s.tx.Transaction(ctx, func(ctx context.Context) error {
		account, err = s.accountRepo.Delete(ctx, actor.ID)
		if err != nil {
			return err
		}

		email, err = s.emailRepo.GetByID(ctx, actor.ID, WithDeleted(DeletedFilterAll))
		if err != nil {
			return err
		}

		sessionIDs, err = s.sessionRepo.DeleteManyForAccount(ctx, actor.ID)
		if err != nil {
			return err
		}

		return s.messenger.WriteAccountDeleted(ctx, account, email)
	}); err != nil {
		return err
	}

	go func() {
		ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
		defer cancel()

		if err := s.accountCache.Delete(ctx, actor.ID); err != nil {
			s.log.Error("failed to delete account cache by id", "error", err)
		}

		if err := s.emailCache.DeleteByID(ctx, actor.ID); err != nil {
			s.log.Error("failed to delete email cache by id", "error", err)
		}

		if err := s.passwordCache.DeleteByID(ctx, actor.ID); err != nil {
			s.log.Error("failed to delete password cache by id", "error", err)
		}

		for _, id := range sessionIDs {
			if err := s.sessionsCache.DeleteByID(ctx, id); err != nil {
				s.log.Error("failed to delete session cache", "session_id", id, "error", err)
			}
		}
	}()

	return nil
}

func (s *Service) checkUsernameRequirements(username string) error {
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

func (s *Service) checkPasswordRequirements(password string) error {
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
