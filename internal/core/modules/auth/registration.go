package auth

import (
	"context"
	"fmt"

	"github.com/netbill/auth-svc/internal/core/errx"
	"github.com/netbill/auth-svc/internal/core/models"
	"github.com/netbill/restkit/tokens"
	"golang.org/x/crypto/bcrypt"
)

type RegistrationParams struct {
	Email    string
	Username string
	Password string
	Role     string
}

func (m *Module) Registration(
	ctx context.Context,
	params RegistrationParams,
) (models.Account, error) {
	check, err := m.repo.ExistsAccountByEmail(ctx, params.Email)
	if err != nil {
		return models.Account{}, err
	}
	if check {
		return models.Account{}, errx.ErrorEmailAlreadyExist.Raise(
			fmt.Errorf("account with email %s already exists", params.Email),
		)
	}

	check, err = m.repo.ExistsAccountByUsername(ctx, params.Username)
	if err != nil {
		return models.Account{}, err
	}
	if check {
		return models.Account{}, errx.ErrorUsernameAlreadyTaken.Raise(
			fmt.Errorf("account with username %s already exists", params.Username),
		)
	}

	err = tokens.ValidateUserSystemRole(params.Role)
	if err != nil {
		return models.Account{}, err
	}

	err = m.password.CheckRequirements(params.Password)
	if err != nil {
		return models.Account{}, err
	}

	err = m.checkUsernameRequirements(ctx, params.Username)
	if err != nil {
		return models.Account{}, err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(params.Password), bcrypt.DefaultCost)
	if err != nil {
		return models.Account{}, err
	}

	var account models.Account
	err = m.repo.Transaction(ctx, func(ctx context.Context) error {
		account, err = m.repo.CreateAccount(ctx, CreateAccountParams{
			Role:         params.Role,
			Username:     params.Username,
			Email:        params.Email,
			PasswordHash: string(hash),
		})
		if err != nil {
			return err
		}

		if err = m.messenger.WriteAccountCreated(ctx, account); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return models.Account{}, err
	}

	return account, nil
}
