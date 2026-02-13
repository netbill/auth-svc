package controller

import (
	"errors"
	"net/http"

	"github.com/netbill/auth-svc/internal/core/errx"
	"github.com/netbill/auth-svc/internal/core/modules/account"
	"github.com/netbill/auth-svc/internal/rest/requests"
	"github.com/netbill/auth-svc/internal/rest/responses"
	"github.com/netbill/auth-svc/internal/rest/scope"
	"github.com/netbill/restkit/problems"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

const operationRegistrationByAdmin = "registration_by_admin"

func (c *Controller) RegistrationByAdmin(w http.ResponseWriter, r *http.Request) {
	log := scope.Log(r).WithOperation(operationRegistrationByAdmin)

	req, err := requests.RegistrationAdmin(r)
	if err != nil {
		log.WithError(err).Info("invalid registration admin request")
		c.responser.RenderErr(w, problems.BadRequest(err)...)
		return
	}

	u, err := c.core.Registration(
		r.Context(),
		account.RegistrationParams{
			Email:    req.Data.Attributes.Email,
			Username: req.Data.Attributes.Username,
			Password: req.Data.Attributes.Password,
			Role:     req.Data.Attributes.Role,
		},
	)
	switch {
	case errors.Is(err, errx.ErrorNotEnoughRights):
		log.Info("not enough rights to register admin")
		c.responser.RenderErr(w, problems.Forbidden("only admins can register new admin accounts"))
	case errors.Is(err, errx.ErrorEmailAlreadyExist):
		log.Info("email already exists")
		c.responser.RenderErr(w, problems.Conflict("user with this email already exists"))
	case errors.Is(err, errx.ErrorUsernameAlreadyTaken):
		log.Info("username already taken")
		c.responser.RenderErr(w, problems.Conflict("user with this username already exists"))
	case errors.Is(err, errx.ErrorUsernameIsNotAllowed):
		log.Info("username is not allowed")
		c.responser.RenderErr(w, problems.BadRequest(validation.Errors{
			"repo/attributes/username": err,
		})...)
	case errors.Is(err, errx.ErrorPasswordIsNotAllowed):
		log.Info("password is not allowed")
		c.responser.RenderErr(w, problems.BadRequest(validation.Errors{
			"repo/attributes/password": err,
		})...)
	case errors.Is(err, errx.ErrorRoleNotSupported):
		log.Info("role is not supported")
		c.responser.RenderErr(w, problems.BadRequest(validation.Errors{
			"repo/attributes/role": err,
		})...)
	case err != nil:
		log.WithError(err).Error("registration by admin failed")
		c.responser.RenderErr(w, problems.InternalError())
	default:
		c.responser.Render(w, http.StatusCreated, responses.Account(u))
	}
}
