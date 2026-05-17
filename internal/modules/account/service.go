package account

import (
	"context"
	"fmt"
	"strings"
	"unicode"

	"github.com/google/uuid"
	"github.com/netbill/auth-svc/internal/errx"
	"github.com/netbill/auth-svc/internal/models"
	"github.com/netbill/restkit/tokens"
)

type DeletedFilter uint8

const (
	DeletedFilterActive  DeletedFilter = iota // 0 — default: only active (deleted_at IS NULL)
	DeletedFilterAll                          // active + deleted (no filter)
	DeletedFilterDeleted                      // only deleted (deleted_at IS NOT NULL)
)

type GetAccountOptions struct {
	Deleted DeletedFilter
}

type GetAccountOption func(*GetAccountOptions)

func ApplyGetAccountOptions(optFns []GetAccountOption) GetAccountOptions {
	var opts GetAccountOptions
	for _, fn := range optFns {
		fn(&opts)
	}
	return opts
}

func WithDeleted(f DeletedFilter) GetAccountOption {
	return func(o *GetAccountOptions) {
		o.Deleted = f
	}
}

//go:generate mockery --name=auth --inpackage
type auth interface {
	ValidateSession(ctx context.Context, actor models.AccountActor) (models.Account, models.Session, error)
}

type Service struct {
	auth auth

	accountRepo  accountRepo
	emailRepo    emailRepo
	passwordRepo passwordRepo
	sessionRepo  sessionRepo

	tx transaction

	accountCache  accountCache
	emailCache    emailCache
	passwordCache passwordCache
	sessionsCache sessionsCache

	passManager passwordManager

	messenger messenger
}

type ServiceDeps struct {
	Auth auth

	AccountRepo  accountRepo
	EmailRepo    emailRepo
	PasswordRepo passwordRepo
	SessionRepo  sessionRepo

	Tx transaction

	AccountCache  accountCache
	EmailCache    emailCache
	PasswordCache passwordCache
	SessionsCache sessionsCache

	PassManager passwordManager

	Messenger messenger
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
	}
}

//go:generate mockery --name=passwordManager --inpackage
type passwordManager interface {
	CheckMatch(password, hash string) error
	GenerateHash(password string) (string, error)
}

//go:generate mockery --name=messenger --inpackage
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

	detached := context.WithoutCancel(ctx)
	go s.accountCache.Set(detached, account)
	go s.emailCache.Set(detached, email)
	go s.passwordCache.Set(detached, password)

	return account, nil
}

func (s *Service) GetMyAccountByID(
	ctx context.Context,
	actor models.AccountActor,
) (models.Account, error) {
	if cached, err := s.accountCache.Get(ctx, actor.ID); err == nil {
		return cached, nil
	}

	account, err := s.accountRepo.GetByID(ctx, actor.ID)
	if err != nil {
		return models.Account{}, err
	}

	go s.accountCache.Set(context.WithoutCancel(ctx), account)

	return account, nil
}

func (s *Service) GetMyEmailByID(
	ctx context.Context,
	actor models.AccountActor,
) (models.AccountEmail, error) {
	if cached, err := s.emailCache.GetByID(ctx, actor.ID); err == nil {
		return cached, nil
	}

	email, err := s.emailRepo.GetByID(ctx, actor.ID)
	if err != nil {
		return models.AccountEmail{}, err
	}

	go s.emailCache.Set(context.WithoutCancel(ctx), email)

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

	current, err := s.accountRepo.GetByID(ctx, actor.ID)
	if err != nil {
		return models.Account{}, err
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

	go s.accountCache.Set(context.WithoutCancel(ctx), updated)

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

	pwd, err := s.passwordCache.Get(ctx, actor.ID)
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

	go s.passwordCache.Set(context.WithoutCancel(ctx), updated)

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

	detached := context.WithoutCancel(ctx)

	go s.accountCache.Delete(detached, actor.ID)
	go s.emailCache.DeleteByID(detached, actor.ID)
	go s.passwordCache.Delete(detached, actor.ID)

	for _, id := range sessionIDs {
		go s.sessionsCache.Delete(detached, id)
	}

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
		return errx.ErrorPasswordIsNotAllowed.Raise(fmt.Errorf("need at least one uppercase letter"))
	}
	if !hasLower {
		return errx.ErrorPasswordIsNotAllowed.Raise(fmt.Errorf("need at least one lower case letter"))
	}
	if !hasDigit {
		return errx.ErrorPasswordIsNotAllowed.Raise(fmt.Errorf("need at least one digit"))
	}
	if !hasSpecial {
		return errx.ErrorPasswordIsNotAllowed.Raise(
			fmt.Errorf("need at least one special character from %s", allowedSpecials),
		)
	}

	return nil
}
