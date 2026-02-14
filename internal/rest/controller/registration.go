package controller

import (
	"errors"
	"net/http"

	"github.com/netbill/auth-svc/internal/core/errx"
	"github.com/netbill/auth-svc/internal/core/modules/auth"
	"github.com/netbill/auth-svc/internal/rest/requests"
	"github.com/netbill/auth-svc/internal/rest/scope"
	"github.com/netbill/restkit/problems"
	"github.com/netbill/restkit/tokens"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

const operationRegistration = "registration"

func (c *Controller) Registration(w http.ResponseWriter, r *http.Request) {
	log := scope.Log(r).WithOperation(operationRegistration)

	req, err := requests.Registration(r)
	if err != nil {
		log.WithError(err).Info("invalid registration request")
		c.responser.RenderErr(w, problems.BadRequest(err)...)
		return
	}

	_, err = c.core.Registration(r.Context(), auth.RegistrationParams{
		Email:    req.Data.Attributes.Email,
		Password: req.Data.Attributes.Password,
		Username: req.Data.Attributes.Username,
		Role:     tokens.RoleSystemAdmin,
	})

	switch {
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
	case err != nil:
		log.WithError(err).Error("registration failed")
		c.responser.RenderErr(w, problems.InternalError())
	default:
		w.WriteHeader(http.StatusCreated)
	}
}
