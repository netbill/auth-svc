package auth

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/umisto/sso-svc/internal/domain/errx"
	"github.com/umisto/sso-svc/internal/domain/models"
)

func (s Service) GetAccountByID(ctx context.Context, ID uuid.UUID) (models.Account, error) {
	account, err := s.db.GetAccountByID(ctx, ID)
	if err != nil {
		return models.Account{}, errx.ErrorInternal.Raise(
			fmt.Errorf("failed to get account with id '%s', cause: %w", ID, err),
		)
	}

	if account.IsNil() {
		return models.Account{}, errx.ErrorAccountNotFound.Raise(
			fmt.Errorf("account with id '%s' not found", ID),
		)
	}

	return account, nil
}

func (s Service) GetAccountByEmail(ctx context.Context, email string) (models.Account, error) {
	account, err := s.db.GetAccountByEmail(ctx, email)
	if err != nil {
		return models.Account{}, errx.ErrorInternal.Raise(
			fmt.Errorf("failed to get account with email '%s', cause: %w", email, err),
		)
	}

	if account.IsNil() {
		return models.Account{}, errx.ErrorAccountNotFound.Raise(
			fmt.Errorf("account with email '%s' not found", email),
		)
	}

	return account, nil
}

func (s Service) AccountExistsByEmail(ctx context.Context, email string) (bool, error) {
	account, err := s.db.GetAccountByEmail(ctx, email)
	if err != nil {
		return false, errx.ErrorInternal.Raise(
			fmt.Errorf("failed to get account with email '%s', cause: %w", email, err),
		)
	}

	return !account.IsNil(), nil
}

func (s Service) GetAccountByUsername(ctx context.Context, username string) (models.Account, error) {
	account, err := s.db.GetAccountByUsername(ctx, username)
	if err != nil {
		return models.Account{}, errx.ErrorInternal.Raise(
			fmt.Errorf("failed to get account with username '%s', cause: %w", username, err),
		)
	}

	if account.IsNil() {
		return models.Account{}, errx.ErrorAccountNotFound.Raise(
			fmt.Errorf("account with username '%s' not found", username),
		)
	}

	return account, nil
}

func (s Service) AccountExistsByUsername(ctx context.Context, username string) (bool, error) {
	account, err := s.db.GetAccountByUsername(ctx, username)
	if err != nil {
		return false, errx.ErrorInternal.Raise(
			fmt.Errorf("failed to get account with username '%s', cause: %w", username, err),
		)
	}

	return !account.IsNil(), nil
}

func (s Service) GetAccountEmail(ctx context.Context, ID uuid.UUID) (models.AccountEmail, error) {
	accountEmail, err := s.db.GetAccountEmail(ctx, ID)
	if err != nil {
		return models.AccountEmail{}, errx.ErrorInternal.Raise(
			fmt.Errorf("failed to get account email repo with id '%s', cause: %w", ID, err),
		)
	}

	if accountEmail.IsNil() {
		return models.AccountEmail{}, errx.ErrorAccountNotFound.Raise(
			fmt.Errorf("account email repo with id '%s' not found", ID),
		)
	}

	return accountEmail, nil
}
