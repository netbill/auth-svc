package controller

import (
	"context"

	"github.com/netbill/auth-svc/internal/models"
	"github.com/netbill/auth-svc/internal/modules/account"
)

type accountCore interface {
	Registration(ctx context.Context, params account.RegistrationParams) (models.Account, error)

	GetMyAccountByID(ctx context.Context, actor models.AccountActor) (models.Account, error)
	GetMyEmailByID(ctx context.Context, actor models.AccountActor) (models.AccountEmail, error)

	UpdatePassword(ctx context.Context, actor models.AccountActor, oldPassword, newPassword string) error
	UpdateUsername(ctx context.Context, actor models.AccountActor, newUsername string) (models.Account, error)

	DeleteMyAccount(ctx context.Context, actor models.AccountActor) error
}

type AccountController struct {
	accounts accountCore
}

func NewAccountController(accounts accountCore) *AccountController {
	return &AccountController{accounts: accounts}
}
