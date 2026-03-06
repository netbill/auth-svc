package auth

import (
	"context"
	"fmt"
	"unicode"

	"github.com/netbill/auth-svc/internal/errx"
	"github.com/netbill/auth-svc/internal/models"
	"github.com/netbill/restkit/tokens"
)

func (s *Service) checkUsernameRequirements(ctx context.Context, username string) error {
	exist, err := s.account.ExistsByUsername(ctx, username)
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
	check, err := s.email.ExistsByEmail(ctx, params.Email)
	if err != nil {
		return models.Account{}, err
	}
	if check {
		return models.Account{}, errx.ErrorEmailAlreadyExist.Raise(
			fmt.Errorf("account with email %s already exists", params.Email),
		)
	}

	check, err = s.account.ExistsByUsername(ctx, params.Username)
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
		return models.Account{}, errx.ErrorRoleNotSupported.Raise(err)
	}

	err = s.passManager.CheckRequirements(params.Password)
	if err != nil {
		return models.Account{}, err
	}

	err = s.checkUsernameRequirements(ctx, params.Username)
	if err != nil {
		return models.Account{}, err
	}

	hash, err := s.passManager.GenerateHash(params.Password)
	if err != nil {
		return models.Account{}, err
	}

	var account models.Account
	err = s.tx.Transaction(ctx, func(ctx context.Context) error {
		account, err = s.account.Create(ctx, params)
		if err != nil {
			return err
		}

		if err := s.email.Create(ctx, models.AccountEmail{
			AccountID: account.ID,
			Email:     params.Email,
		}); err != nil {
			return err
		}

		if err = s.password.Create(ctx, models.AccountPassword{
			AccountID: account.ID,
			Hash:      hash,
		}); err != nil {
			return err
		}

		if err = s.messenger.WriteAccountCreated(ctx, account); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return models.Account{}, err
	}

	return account, nil
}

func (s *Service) GetMyAccountByID(
	ctx context.Context,
	actor models.AccountActor,
) (models.Account, error) {
	account, _, err := s.validateActorSession(ctx, actor)
	if err != nil {
		return models.Account{}, err
	}

	return account, nil
}

func (s *Service) GetMyAccountEmail(
	ctx context.Context,
	actor models.AccountActor,
) (models.AccountEmail, error) {
	_, _, err := s.validateActorSession(ctx, actor)
	if err != nil {
		return models.AccountEmail{}, err
	}

	return s.email.GetByID(ctx, actor.ID)
}

func (s *Service) DeleteMyAccount(
	ctx context.Context,
	actor models.AccountActor,
) error {
	account, _, err := s.validateActorSession(ctx, actor)
	if err != nil {
		return err
	}

	exists, err := s.org.ExistMemberByAccount(ctx, actor.ID)
	if err != nil {
		return err
	}
	if exists {
		return errx.ErrorAccountHaveMembershipInOrg.Raise(
			fmt.Errorf("account %s has a member of organizations", actor.ID),
		)
	}

	return s.tx.Transaction(ctx, func(ctx context.Context) error {
		if err = s.tombstone.BuryAccount(ctx, account.ID); err != nil {
			return err
		}

		if err = s.account.Delete(ctx, actor.ID); err != nil {
			return err
		}

		return s.messenger.WriteAccountDeleted(ctx, account.ID)
	})
}

func (s *Service) UpdateUsername(
	ctx context.Context,
	actor models.AccountActor,
	newUsername string,
) (account models.Account, err error) {
	account, _, err = s.validateActorSession(ctx, actor)
	if err != nil {
		return models.Account{}, err
	}

	if err = s.checkUsernameRequirements(ctx, newUsername); err != nil {
		return models.Account{}, err
	}

	err = s.tx.Transaction(ctx, func(ctx context.Context) error {
		account, err = s.account.UpdateUsername(ctx, actor.ID, newUsername)
		if err != nil {
			return err
		}
		return s.messenger.WriteAccountUsernameUpdated(ctx, account)
	})
	if err != nil {
		return models.Account{}, err
	}

	return account, nil
}

func (s *Service) UpdatePassword(
	ctx context.Context,
	actor models.AccountActor,
	oldPassword, newPassword string,
) error {
	account, _, err := s.validateActorSession(ctx, actor)
	if err != nil {
		return err
	}

	passData, err := s.password.GetByID(ctx, actor.ID)
	if err != nil {
		return err
	}

	if err = passData.CanChangePassword(); err != nil {
		return err
	}

	if err = s.passManager.CheckPasswordMatch(passData.Hash, passData.Hash); err != nil {
		return err
	}

	if err = s.passManager.CheckRequirements(newPassword); err != nil {
		return err
	}

	hash, err := s.passManager.GenerateHash(newPassword)
	if err != nil {
		return err
	}

	return s.tx.Transaction(ctx, func(ctx context.Context) error {
		_, err = s.password.UpdatePassword(ctx, actor.ID, hash)
		if err != nil {
			return err
		}

		if err = s.tombstone.BuryAccountSessions(ctx, actor.ID); err != nil {
			return err
		}

		return s.session.DeleteManyForAccount(ctx, account.ID)
	})
}
