package controller

import (
	"errors"
	"net/http"

	"github.com/netbill/auth-svc/internal/core/errx"
	"github.com/netbill/auth-svc/internal/core/modules/auth"
	"github.com/netbill/auth-svc/internal/rest/requests"
	"github.com/netbill/auth-svc/internal/rest/scope"
	"github.com/netbill/restkit/problems"
	"github.com/netbill/restkit/render"
	"github.com/netbill/restkit/tokens"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

const operationRegistration = "registration"

func (c *Controller) Registration(w http.ResponseWriter, r *http.Request) {
	log := scope.Log(r).WithOperation(operationRegistration)

	req, err := requests.Registration(r)
	if err != nil {
		log.WithError(err).Info("invalid registration request")
		render.ResponseError(w, problems.BadRequest(err)...)
		return
	}

	_, err = c.core.Registration(r.Context(), auth.RegistrationParams{
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
