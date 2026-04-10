package account

import (
	"context"
	"errors"
	"log/slog"

	"github.com/google/uuid"
	"github.com/netbill/auth-svc/internal/errx"
	"github.com/netbill/auth-svc/internal/models"
	"github.com/netbill/restkit/tokens"
)

type Service struct {
	accountRepo  accountRepo
	emailRepo    emailRepo
	passwordRepo passwordRepo
	sessionRepo  sessionRepo
	tx           transaction

	accountCache  accountCache
	emailCache    emailCache
	passwordCache passwordCache
	sessionsCache sessionsCache

	passManager  passwordManager
	tokenManager tokenManager

	messenger messenger

	log *slog.Logger
}

type passwordManager interface {
	CheckRequirements(password string) error
	CheckMatch(password, hash string) error

	GenerateHash(password string) (string, error)
}

type tokenManager interface {
	ParseAccountAuthAccess(token string) (tokens.AccountAuthClaims, error)
	ParseAccountAuthRefresh(token string) (tokens.AccountAuthClaims, error)

	HashRefresh(token string) (string, error)

	GenerateAccess(account models.Account, sessionID uuid.UUID) (string, error)
	GenerateRefresh(account models.Account, sessionID uuid.UUID) (string, error)
}

type messenger interface {
	WriteAccountCreated(ctx context.Context, account models.Account) error

	WriteAccountUsernameUpdated(ctx context.Context, account models.Account) error
	WriteAccountEmailVerified(ctx context.Context, account models.Account) error

	WriteAccountDeleted(ctx context.Context, accountID uuid.UUID) error
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
) (account models.Account, err error) {
	if err = tokens.ValidateUserSystemRole(params.Role); err != nil {
		return models.Account{}, errx.ErrorRoleNotSupported.Raise(err)
	}

	if err = s.passManager.CheckRequirements(params.Password); err != nil {
		return models.Account{}, err
	}

	hash, err := s.passManager.GenerateHash(params.Password)
	if err != nil {
		return models.Account{}, err
	}

	if err = s.tx.Transaction(ctx, func(ctx context.Context) error {
		account, err = s.accountRepo.Create(ctx, params)
		if err != nil {
			return err
		}

		err = s.emailRepo.Create(ctx, models.AccountEmail{
			AccountID: account.ID,
			Email:     params.Email,
		})
		if err != nil {
			return err
		}

		return s.passwordRepo.Create(ctx, models.AccountPassword{
			AccountID: account.ID,
			Hash:      hash,
		})
	}); err != nil {
		return models.Account{}, err
	}

	err = s.accountCache.SetByID(ctx, account)
	if err != nil {
		s.log.Error("failed to set account", "error", err)
	}

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

	err = s.accountCache.SetByID(ctx, account)
	if err != nil {
		s.log.Error("failed to set account", "error", err)
	}

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

	err = s.emailCache.SetByID(ctx, email)
	if err != nil {
		s.log.Error("failed to set account", "error", err)
	}

	return email, nil
}

func (s *Service) UpdateUsername(
	ctx context.Context,
	actor models.AccountActor,
	newUsername string,
) (models.Account, error) {
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

	if err = s.accountCache.DeleteByUsername(ctx, current.Username); err != nil {
		s.log.Error("failed to delete account cache by username", "error", err)
	}
	if err = s.accountCache.Delete(ctx, actor.ID); err != nil {
		s.log.Error("failed to delete account cache by id", "error", err)
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

	if err = s.accountCache.SetByID(ctx, updated); err != nil {
		s.log.Error("failed to set account cache by id", "error", err)
	}
	if err = s.accountCache.SetByUsername(ctx, updated); err != nil {
		s.log.Error("failed to set account cache by username", "error", err)
	}

	return updated, nil
}

func (s *Service) DeleteMyAccount(
	ctx context.Context,
	actor models.AccountActor,
) error {
	if err := s.accountRepo.Delete(ctx, actor.ID); err != nil {
		return err
	}

	if err := s.accountCache.Delete(ctx, actor.ID); err != nil {
		s.log.Error("failed to delete account cache by id", "error", err)
	}

	return nil
}
