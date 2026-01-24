package account

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/netbill/auth-svc/internal/core/errx"
	"github.com/netbill/auth-svc/internal/core/models"
	"github.com/netbill/restkit/tokens/roles"
	"golang.org/x/crypto/bcrypt"
)

type RegistrationParams struct {
	Email    string
	Username string
	Password string
	Role     string
}

func (s Service) Registration(
	ctx context.Context,
	params RegistrationParams,
) (models.Account, error) {
	check, err := s.accountExistsByEmail(ctx, params.Email)
	if err != nil {
		return models.Account{}, err
	}
	if check {
		return models.Account{}, errx.ErrorEmailAlreadyExist.Raise(
			fmt.Errorf("account with email '%s' already exists", params.Email),
		)
	}

	err = roles.ValidateUserSystemRole(params.Role)
	if err != nil {
		return models.Account{}, errx.ErrorRoleNotSupported.Raise(
			fmt.Errorf("failed to parsing role for new account with email '%s', cause: %w", params.Email, err),
		)
	}

	err = s.checkPasswordRequirements(params.Password)
	if err != nil {
		return models.Account{}, err
	}

	err = s.checkUsernameRequirements(ctx, params.Username)
	if err != nil {
		return models.Account{}, err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(params.Password), bcrypt.DefaultCost)
	if err != nil {
		return models.Account{}, errx.ErrorInternal.Raise(
			fmt.Errorf("failed to hashing password, cause: %w", err),
		)
	}

	var account models.Account
	err = s.repo.Transaction(ctx, func(ctx context.Context) error {
		account, err = s.repo.CreateAccount(ctx, CreateAccountParams{
			Role:         params.Role,
			Username:     params.Username,
			Email:        params.Email,
			PasswordHash: string(hash),
		})
		if err != nil {
			return errx.ErrorInternal.Raise(
				fmt.Errorf("failed to inserting new account with email '%s', cause: %w", params.Email, err),
			)
		}

		if err = s.messenger.WriteAccountCreated(ctx, account); err != nil {
			return errx.ErrorInternal.Raise(
				fmt.Errorf("failed to publish account created messenger for account '%s', cause: %w", account.ID, err),
			)
		}

		return nil
	})
	if err != nil {
		return models.Account{}, err
	}

	return account, nil
}

func (s Service) RegistrationByAdmin(
	ctx context.Context,
	initiatorID uuid.UUID,
	params RegistrationParams,
) (models.Account, error) {
	initiator, err := s.repo.GetAccountByID(ctx, initiatorID)
	if err != nil {
		return models.Account{}, errx.ErrorInternal.Raise(
			fmt.Errorf("failed to get initiator with id '%s', cause: %w", initiatorID, err),
		)
	}
	if initiator.IsNil() {
		return models.Account{}, errx.ErrorInitiatorNotFound.Raise(
			fmt.Errorf("initiator with id '%s' not found", initiatorID),
		)
	}

	if initiator.Role != roles.SystemAdmin {
		return models.Account{}, errx.ErrorNotEnoughRights.Raise(
			fmt.Errorf("account %s has insufficient permissions to register admin accounts", initiatorID),
		)
	}

	account, err := s.Registration(ctx, params)
	if err != nil {
		return models.Account{}, err
	}

	err = s.checkUsernameRequirements(ctx, params.Username)
	if err != nil {
		return models.Account{}, err
	}

	err = s.messenger.WriteAccountCreated(ctx, account)
	if err != nil {
		return models.Account{}, errx.ErrorInternal.Raise(
			fmt.Errorf("failed to publish admin created messenger for account '%s', cause: %w", account.ID, err),
		)
	}

	return account, nil
}
