package models

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/umisto/sso-svc/internal/domain/errx"
	"golang.org/x/crypto/bcrypt"
)

const updateUsernameCooldown = 14 * 24 * time.Hour
const updatePasswordCooldown = 30 * 24 * time.Hour
const updateEmailCooldown = 30 * 24 * time.Hour

const (
	AccountStatusActive      = "active"
	AccountStatusDeactivated = "deactivated"
	AccountStatusSuspended   = "suspended"
)

var accountStatuses = []string{
	AccountStatusActive,
	AccountStatusDeactivated,
	AccountStatusSuspended,
}

var ErrorAccountStatusIsNotSupported = fmt.Errorf("account status is not supported, must be one of: %v", GetAllAccountStatuses())

func CheckAccountStatus(status string) error {
	for _, accountStatus := range accountStatuses {
		if accountStatus == status {
			return nil
		}
	}

	return fmt.Errorf("%s: %w", status, ErrorAccountStatusIsNotSupported)
}

func GetAllAccountStatuses() []string {
	return accountStatuses
}

type Account struct {
	ID       uuid.UUID `json:"id"`
	Username string    `json:"username"`
	Role     string    `json:"role"`
	Status   string    `json:"status"`

	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
	UsernameUpdatedAt time.Time `json:"username_name_updated_at"`
}

func (a Account) IsNil() bool {
	return a.ID == uuid.Nil
}

func (a Account) CanChangeUsername() error {
	if time.Since(a.UsernameUpdatedAt) >= updateUsernameCooldown {
		return nil
	}

	return errx.ErrorCannotChangeUsernameYet.Raise(fmt.Errorf(
		"account with id %s cannot change username yet", a.ID),
	)
}

func (a Account) CanInteract() error {
	if a.Status != AccountStatusActive {
		return errx.ErrorInitiatorIsNotActive.Raise(fmt.Errorf(
			"account with id %s is blocked and cannot interact", a.ID),
		)
	}

	return nil
}

type AccountPassword struct {
	AccountID uuid.UUID `json:"account_id"`
	Hash      string    `json:"hash"`
	UpdatedAt time.Time `json:"updated_at"`
	CreatedAt time.Time `json:"created_at"`
}

func (ap AccountPassword) IsNil() bool {
	return ap.AccountID == uuid.Nil
}

func (ap AccountPassword) CanChangePassword() error {
	if time.Since(ap.UpdatedAt) >= updatePasswordCooldown {
		return nil
	}

	return errx.ErrorCannotChangePasswordYet.Raise(fmt.Errorf(
		"account with id %s cannot change password yet", ap.AccountID),
	)
}

func (ap AccountPassword) CheckPasswordMatch(password string) error {
	if err := bcrypt.CompareHashAndPassword([]byte(ap.Hash), []byte(password)); err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return errx.ErrorPasswordInvalid.Raise(
				fmt.Errorf("invalid credentials, cause: %w", err),
			)
		}

		return errx.ErrorInternal.Raise(
			fmt.Errorf("comparing password hash, cause: %w", err),
		)
	}

	return nil
}

type AccountEmail struct {
	AccountID uuid.UUID `json:"account_id"`
	Email     string    `json:"email"`
	Verified  bool      `json:"verified"`
	UpdatedAt time.Time `json:"updated_at"`
	CreatedAt time.Time `json:"created_at"`
}

func (ae AccountEmail) IsNil() bool {
	return ae.AccountID == uuid.Nil
}

func (ae AccountEmail) CanChangeEmail() error {
	if time.Since(ae.UpdatedAt) >= updateEmailCooldown {
		return nil
	}

	return errx.ErrorCannotChangeEmailYet.Raise(fmt.Errorf(
		"account with id %s cannot change email yet", ae.AccountID),
	)
}

func (ae AccountEmail) IsVerified() error {
	if ae.Verified {
		return nil
	}

	return errx.ErrorEmailNotVerified.Raise(fmt.Errorf(
		"account with id %s has unverified email", ae.AccountID),
	)
}
