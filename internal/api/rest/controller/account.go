package controller

import (
	"context"
	"errors"
	"net/http"
	"slices"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/netbill/auth-svc/internal/api/rest/requests"
	"github.com/netbill/auth-svc/internal/api/rest/responses"
	"github.com/netbill/auth-svc/internal/api/rest/scope"
	"github.com/netbill/auth-svc/internal/errx"
	"github.com/netbill/auth-svc/internal/models"
	"github.com/netbill/auth-svc/internal/modules/account"
	"github.com/netbill/restkit"
	"github.com/netbill/restkit/problems"
	"github.com/netbill/restkit/render"
	"github.com/netbill/restkit/tokens"
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

const operationRegistration = "registration"

func (c *AccountController) Registration(w http.ResponseWriter, r *http.Request) {
	log := scope.Log(r).WithOperation(operationRegistration)

	req, err := requests.Registration(r)
	if err != nil {
		log.WithError(err).Info("invalid registration request")
		render.ResponseError(w, problems.BadRequest(err)...)
		return
	}

	_, err = c.accounts.Registration(r.Context(), account.RegistrationParams{
		Email:    req.Data.Attributes.Email,
		Password: req.Data.Attributes.Password,
		Username: req.Data.Attributes.Username,
		Role:     tokens.RoleSystemUser,
	})

	switch {
	case errors.Is(err, errx.ErrorEmailAlreadyExist):
		log.WithError(err).Warn("email already exists")
		render.ResponseError(w, problems.Conflict("user with this email already exists"))
	case errors.Is(err, errx.ErrorUsernameAlreadyTaken):
		log.WithError(err).Warn("username already taken")
		render.ResponseError(w, problems.Conflict("user with this username already exists"))
	case errors.Is(err, errx.ErrorUsernameIsNotAllowed):
		log.WithError(err).Warn("username is not allowed")
		render.ResponseError(w, problems.BadRequest(validation.Errors{
			"data/attributes/username": err,
		})...)
	case errors.Is(err, errx.ErrorPasswordIsNotAllowed):
		log.WithError(err).Warn("password is not allowed")
		render.ResponseError(w, problems.BadRequest(validation.Errors{
			"data/attributes/password": err,
		})...)
	case err != nil:
		log.WithError(err).Error("unexpected error")
		render.ResponseError(w, problems.InternalError())
	default:
		log.Info("registration successful")
		w.WriteHeader(http.StatusCreated)
	}
}

const operationRegistrationByAdmin = "registration_by_admin"

func (c *AccountController) RegistrationByAdmin(w http.ResponseWriter, r *http.Request) {
	log := scope.Log(r).WithOperation(operationRegistrationByAdmin)

	req, err := requests.RegistrationAdmin(r)
	if err != nil {
		log.WithError(err).Info("invalid registration admin request")
		render.ResponseError(w, problems.BadRequest(err)...)
		return
	}

	log = log.WithFields(map[string]interface{}{
		"email":    req.Data.Attributes.Email,
		"username": req.Data.Attributes.Username,
		"role":     req.Data.Attributes.Role,
	})

	u, err := c.accounts.Registration(r.Context(), account.RegistrationParams{
		Email:    req.Data.Attributes.Email,
		Username: req.Data.Attributes.Username,
		Password: req.Data.Attributes.Password,
		Role:     req.Data.Attributes.Role,
	})
	switch {
	case errors.Is(err, errx.ErrorEmailAlreadyExist):
		log.WithError(err).Warn("email already exists")
		render.ResponseError(w, problems.Conflict("user with this email already exists"))
	case errors.Is(err, errx.ErrorUsernameAlreadyTaken):
		log.WithError(err).Warn("username already taken")
		render.ResponseError(w, problems.Conflict("user with this username already exists"))
	case errors.Is(err, errx.ErrorUsernameIsNotAllowed):
		log.WithError(err).Warn("username is not allowed")
		render.ResponseError(w, problems.BadRequest(validation.Errors{
			"data/attributes/username": err,
		})...)
	case errors.Is(err, errx.ErrorPasswordIsNotAllowed):
		log.WithError(err).Warn("password is not allowed")
		render.ResponseError(w, problems.BadRequest(validation.Errors{
			"data/attributes/password": err,
		})...)
	case errors.Is(err, errx.ErrorRoleNotSupported):
		log.WithError(err).Warn("role is not supported")
		render.ResponseError(w, problems.BadRequest(validation.Errors{
			"data/attributes/role": err,
		})...)
	case err != nil:
		log.WithError(err).Error("unexpected error")
		render.ResponseError(w, problems.InternalError())
	default:
		log.Info("account registered successfully by admin")
		render.Response(w, http.StatusCreated, responses.Account(u))
	}
}

const operationGetMyAccount = "get_my_account"

func (c *AccountController) GetMyAccount(w http.ResponseWriter, r *http.Request) {
	log := scope.Log(r).WithOperation(operationGetMyAccount)

	account, err := c.accounts.GetMyAccountByID(r.Context(), scope.AccountActor(r))
	switch {
	case errors.Is(err, errx.ErrorAccountDeleted):
		log.WithError(err).Warn("account deleted")
		render.ResponseError(w, problems.Unauthorized())
		return
	case errors.Is(err, errx.ErrorAccountNotFound):
		log.WithError(err).Warn("account not found")
		render.ResponseError(w, problems.NotFound("account not found"))
		return
	case err != nil:
		log.WithError(err).Error("unexpected error")
		render.ResponseError(w, problems.InternalError())
		return
	}

	opts := make([]responses.AccountOption, 0, 1)
	includes := restkit.ParseIncludes(r)

	if slices.Contains(includes, "email") {
		email, err := c.accounts.GetMyEmailByID(r.Context(), scope.AccountActor(r))
		switch {
		case errors.Is(err, errx.ErrorAccountDeleted):
			log.WithError(err).Warn("account deleted")
			render.ResponseError(w, problems.Unauthorized())
			return
		case errors.Is(err, errx.ErrorAccountNotFound):
			log.WithError(err).Warn("account not found")
			render.ResponseError(w, problems.NotFound("account not found"))
			return
		case err != nil:
			log.WithError(err).Error("unexpected error")
			render.ResponseError(w, problems.InternalError())
			return
		}

		opts = append(opts, responses.WithAccountEmail(email))
	}

	log.Info("account retrieved")
	render.Response(w, http.StatusOK, responses.Account(account, opts...))
}

const operationUpdatePassword = "update_password"

func (c *AccountController) UpdatePassword(w http.ResponseWriter, r *http.Request) {
	log := scope.Log(r).WithOperation(operationUpdatePassword)

	req, err := requests.UpdatePassword(r)
	if err != nil {
		log.WithError(err).Info("failed to parse update password request")
		render.ResponseError(w, problems.BadRequest(err)...)
		return
	}

	err = c.accounts.UpdatePassword(
		r.Context(),
		scope.AccountActor(r),
		req.Data.Attributes.OldPassword,
		req.Data.Attributes.NewPassword,
	)
	switch {
	case errors.Is(err, errx.ErrorAccountInvalidSession),
		errors.Is(err, errx.ErrorSessionNotFound):
		log.WithError(err).Warn("invalid credentials")
		render.ResponseError(w, problems.Unauthorized())
	case errors.Is(err, errx.ErrorPasswordInvalid):
		log.WithError(err).Warn("invalid old password")
		render.ResponseError(w, problems.Unauthorized())
	case errors.Is(err, errx.ErrorPasswordIsNotAllowed):
		log.WithError(err).Warn("new password is not allowed")
		render.ResponseError(w, problems.BadRequest(validation.Errors{
			"data/attributes/new_password": err,
		})...)
	case err != nil:
		log.WithError(err).Error("unexpected error")
		render.ResponseError(w, problems.InternalError())
	default:
		log.Info("password updated successfully")
		render.Response(w, http.StatusNoContent, nil)
	}
}

const operationDeleteMyAccount = "delete_my_account"

func (c *AccountController) DeleteMyAccount(w http.ResponseWriter, r *http.Request) {
	log := scope.Log(r).WithOperation(operationDeleteMyAccount)

	err := c.accounts.DeleteMyAccount(r.Context(), scope.AccountActor(r))
	switch {
	case errors.Is(err, errx.ErrorAccountInvalidSession),
		errors.Is(err, errx.ErrorSessionNotFound):
		log.WithError(err).Warn("invalid credentials")
		render.ResponseError(w, problems.Unauthorized())
	case errors.Is(err, errx.ErrorAccountNotFound):
		log.WithError(err).Warn("account not found")
		render.ResponseError(w, problems.NotFound("account not found"))
	case err != nil:
		log.WithError(err).Error("unexpected error")
		render.ResponseError(w, problems.InternalError())
	default:
		log.Info("account deleted")
		render.Response(w, http.StatusNoContent, nil)
	}
}
