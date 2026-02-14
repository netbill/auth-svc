package models

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/netbill/auth-svc/internal/core/errx"
)

const updatePasswordCooldown = 30 * 24 * time.Hour

type AccountActor struct {
	ID        uuid.UUID `json:"id"`
	SessionID uuid.UUID `json:"session_id"`
	Role      string    `json:"role"`
}

type Account struct {
	ID       uuid.UUID `json:"id"`
	Username string    `json:"username"`
	Role     string    `json:"role"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type AccountPassword struct {
	AccountID uuid.UUID `json:"account_id"`
	Hash      string    `json:"hash"`
	UpdatedAt time.Time `json:"updated_at"`
	CreatedAt time.Time `json:"created_at"`
}

func (ap AccountPassword) CanChangePassword() error {
	if time.Since(ap.UpdatedAt) >= updatePasswordCooldown {
		return nil
	}

	return errx.ErrorCannotChangePasswordYet.Raise(fmt.Errorf(
		"account with id %s cannot change password yet", ap.AccountID),
	)
}

type AccountEmail struct {
	AccountID uuid.UUID `json:"account_id"`
	Email     string    `json:"email"`
	Verified  bool      `json:"verified"`
	UpdatedAt time.Time `json:"updated_at"`
	CreatedAt time.Time `json:"created_at"`
}

func (ae AccountEmail) IsVerified() error {
	if ae.Verified {
		return nil
	}

	return errx.ErrorEmailNotVerified.Raise(fmt.Errorf(
		"account with id %s has unverified email", ae.AccountID),
	)
}
